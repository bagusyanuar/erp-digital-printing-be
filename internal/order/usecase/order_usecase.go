package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	cashFlowDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/domain"
	orderDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/order/domain"
	productDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/product/domain"
	resellerDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/reseller/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type orderUsecase struct {
	orderRepo    orderDomain.OrderRepository
	productRepo  productDomain.ProductRepository
	resellerRepo resellerDomain.ResellerRepository
	cashFlowRepo cashFlowDomain.CashFlowRepository
	db           *gorm.DB
	logger       *zap.Logger
}

func NewOrderUsecase(
	orderRepo orderDomain.OrderRepository,
	productRepo productDomain.ProductRepository,
	resellerRepo resellerDomain.ResellerRepository,
	cashFlowRepo cashFlowDomain.CashFlowRepository,
	db *gorm.DB,
	logger *zap.Logger,
) orderDomain.OrderUsecase {
	return &orderUsecase{
		orderRepo:    orderRepo,
		productRepo:  productRepo,
		resellerRepo: resellerRepo,
		cashFlowRepo: cashFlowRepo,
		db:           db,
		logger:       logger,
	}
}


func (u *orderUsecase) validateUOMAndGetQty(item *orderDomain.OrderItem) (float64, error) {
	if item.Quantity <= 0 {
		return 0, errors.New("quantity must be greater than 0")
	}

	switch item.UOM {
	case productDomain.UomM2:
		if item.LengthCM == nil || *item.LengthCM <= 0 {
			return 0, errors.New("length_cm is required and must be greater than 0 for UOM m2")
		}
		if item.WidthCM == nil || *item.WidthCM <= 0 {
			return 0, errors.New("width_cm is required and must be greater than 0 for UOM m2")
		}
		areaM2 := (*item.LengthCM / 100.0) * (*item.WidthCM / 100.0) * float64(item.Quantity)
		return areaM2, nil

	case productDomain.UomMLari:
		if item.LengthCM == nil || *item.LengthCM <= 0 {
			return 0, errors.New("length_cm is required and must be greater than 0 for UOM m_lari")
		}
		lengthM := (*item.LengthCM / 100.0) * float64(item.Quantity)
		return lengthM, nil

	case productDomain.UomBox, productDomain.UomPcs:
		item.LengthCM = nil
		item.WidthCM = nil
		return float64(item.Quantity), nil

	default:
		return 0, fmt.Errorf("invalid UOM: %s", item.UOM)
	}
}

// createOrder is the shared internal logic for SaveDraft and SubmitToCashier,
// eliminating code duplication between the two public methods.
func (u *orderUsecase) createOrder(ctx context.Context, order *orderDomain.Order, status string) error {
	order.Status = status
	order.PaymentStatus = orderDomain.PaymentStatusUnpaid
	order.InvoiceNumber = nil

	// Validate items
	if len(order.OrderItems) == 0 {
		return errors.New("order must have at least one item")
	}

	// Determine Customer Level
	var customerLevelID uuid.UUID
	if order.ResellerID != nil && *order.ResellerID != uuid.Nil {
		reseller, err := u.resellerRepo.FindByID(ctx, *order.ResellerID)
		if err != nil {
			return fmt.Errorf("failed to find reseller: %w", err)
		}
		if reseller.CustomerLevelID != nil {
			customerLevelID = *reseller.CustomerLevelID
		} else {
			// Fallback default reseller level UUID
			customerLevelID = uuid.MustParse("d2c67ef8-82e4-4d8b-968b-5a1e2f5b6154")
		}
	} else {
		// Default End User level UUID
		customerLevelID = uuid.MustParse("b3c8f3a3-b26a-4638-b7f2-841a54774844")
	}

	var totalProductPrice float64
	var totalAdditionalCost float64

	for i := range order.OrderItems {
		item := &order.OrderItems[i]

		calcQty, err := u.validateUOMAndGetQty(item)
		if err != nil {
			return fmt.Errorf("item[%d]: %w", i, err)
		}

		var finishingIDs []uuid.UUID
		for _, f := range item.Finishings {
			finishingIDs = append(finishingIDs, f.ID)
		}

		var finishings []orderDomain.Finishing
		if len(finishingIDs) > 0 {
			var err error
			finishings, err = u.orderRepo.FindFinishingsByIDs(ctx, finishingIDs)
			if err != nil {
				return fmt.Errorf("failed to fetch finishings for item %d: %w", i, err)
			}
		}
		item.Finishings = finishings

		var finishingCost float64
		for _, f := range finishings {
			finishingCost += f.Price
		}

		qtyInt := int(calcQty)
		if qtyInt < 1 {
			qtyInt = 1
		}

		priceRes, err := u.productRepo.CheckPrice(ctx, item.ProductVariantID, customerLevelID, qtyInt)
		if err != nil {
			return fmt.Errorf("failed to check price tier for variant %s: %w", item.ProductVariantID, err)
		}

		item.PricePerUnit = priceRes.PricePerUnit
		item.AdditionalCost = finishingCost
		item.Subtotal = (priceRes.PricePerUnit * calcQty) + (finishingCost * float64(item.Quantity))

		totalProductPrice += priceRes.PricePerUnit * calcQty
		totalAdditionalCost += finishingCost * float64(item.Quantity)
	}

	order.TotalProductPrice = totalProductPrice
	order.TotalAdditionalCost = totalAdditionalCost
	order.GrandTotal = totalProductPrice + totalAdditionalCost

	// Generate JOB Number
	dateStr := time.Now().Format("20060102")
	seq, err := u.orderRepo.GetNextJobSeq(ctx, dateStr)
	if err != nil {
		return fmt.Errorf("failed to generate job number: %w", err)
	}
	order.JobNumber = fmt.Sprintf("JOB/%s/%04d", dateStr, seq)

	return u.orderRepo.Create(ctx, order)
}

func (u *orderUsecase) SaveDraft(ctx context.Context, order *orderDomain.Order) error {
	return u.createOrder(ctx, order, orderDomain.StatusDraft)
}

func (u *orderUsecase) SubmitToCashier(ctx context.Context, order *orderDomain.Order) error {
	return u.createOrder(ctx, order, orderDomain.StatusPendingPayment)
}

func (u *orderUsecase) SubmitExistingToCashier(ctx context.Context, orderID uuid.UUID) error {
	order, err := u.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	if order.Status != orderDomain.StatusDraft {
		return fmt.Errorf("cannot submit order in %s status, only DRAFT status allowed", order.Status)
	}

	order.Status = orderDomain.StatusPendingPayment
	return u.orderRepo.Update(ctx, order)
}

func (u *orderUsecase) FindByID(ctx context.Context, id uuid.UUID) (*orderDomain.Order, error) {
	return u.orderRepo.FindByID(ctx, id)
}

func (u *orderUsecase) FindAll(ctx context.Context, params request.PaginationParam, statuses []string, paymentStatuses []string, paymentMethods []string, designerID *uuid.UUID, cashierID *uuid.UUID, search string, startDate *time.Time, endDate *time.Time, customerType string) ([]orderDomain.Order, int64, error) {
	return u.orderRepo.FindAll(ctx, params, statuses, paymentStatuses, paymentMethods, designerID, cashierID, search, startDate, endDate, customerType)
}

func (u *orderUsecase) ProcessPayment(
	ctx context.Context,
	orderID uuid.UUID,
	cashierID uuid.UUID,
	resellerID *uuid.UUID,
	customerName string,
	customerPhone string,
	payments []orderDomain.PaymentItem,
) (*orderDomain.Order, error) {
	order, err := u.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	if order.Status != orderDomain.StatusPendingPayment {
		return nil, fmt.Errorf("order is not in PENDING_PAYMENT status, current: %s", order.Status)
	}

	var amountPaid float64
	var hasTempo bool
	for _, p := range payments {
		if p.AmountPaid < 0 {
			return nil, errors.New("payment amount must be greater than or equal to 0")
		}
		amountPaid += p.AmountPaid
		if p.PaymentMethod == "tempo" {
			hasTempo = true
		}
	}

	// Set Cashier ID
	order.CashierID = &cashierID

	// Determine Customer Level
	var customerLevelID uuid.UUID
	var isReseller bool
	if resellerID != nil && *resellerID != uuid.Nil {
		reseller, err := u.resellerRepo.FindByID(ctx, *resellerID)
		if err != nil {
			return nil, fmt.Errorf("failed to find reseller: %w", err)
		}
		if reseller.CustomerLevelID != nil {
			customerLevelID = *reseller.CustomerLevelID
		} else {
			// Fallback default reseller level UUID
			customerLevelID = uuid.MustParse("d2c67ef8-82e4-4d8b-968b-5a1e2f5b6154")
		}
		order.ResellerID = resellerID
		isReseller = true
	} else {
		// Default End User level UUID
		customerLevelID = uuid.MustParse("b3c8f3a3-b26a-4638-b7f2-841a54774844")
		order.ResellerID = nil
	}

	// Update denormalized customer info
	order.CustomerName = &customerName
	order.CustomerPhone = &customerPhone

	// Recalculate Pricing
	var totalProductPrice float64
	var totalAdditionalCost float64

	for i := range order.OrderItems {
		item := &order.OrderItems[i]

		calcQty, err := u.validateUOMAndGetQty(item)
		if err != nil {
			return nil, fmt.Errorf("item[%d] validation failed: %w", i, err)
		}

		// Calculate total finishing cost
		var finishingCost float64
		for _, f := range item.Finishings {
			finishingCost += f.Price
		}

		// Check unit price from price tiers
		qtyInt := int(calcQty)
		if qtyInt < 1 {
			qtyInt = 1
		}
		priceRes, err := u.productRepo.CheckPrice(ctx, item.ProductVariantID, customerLevelID, qtyInt)
		if err != nil {
			return nil, fmt.Errorf("failed to check price tier for variant %s: %w", item.ProductVariantID, err)
		}

		item.PricePerUnit = priceRes.PricePerUnit
		item.AdditionalCost = finishingCost
		item.Subtotal = (priceRes.PricePerUnit * calcQty) + (finishingCost * float64(item.Quantity))

		totalProductPrice += priceRes.PricePerUnit * calcQty
		totalAdditionalCost += finishingCost * float64(item.Quantity)
	}

	grandTotal := totalProductPrice + totalAdditionalCost
	order.TotalProductPrice = totalProductPrice
	order.TotalAdditionalCost = totalAdditionalCost
	order.GrandTotal = grandTotal
	order.AmountPaid = amountPaid

	// Calculate Payment Status dynamically
	if amountPaid >= grandTotal {
		order.PaymentStatus = orderDomain.PaymentStatusPaid
	} else if amountPaid > 0 {
		order.PaymentStatus = orderDomain.PaymentStatusPartialPaid
	} else {
		order.PaymentStatus = orderDomain.PaymentStatusUnpaid
	}

	// Check Credit Limit if this is Tempo/Hutang (Unpaid or Partial Paid)
	isTempo := hasTempo || order.PaymentStatus == orderDomain.PaymentStatusUnpaid || order.PaymentStatus == orderDomain.PaymentStatusPartialPaid
	if isTempo && isReseller {
		// Fetch Reseller for Credit Limit validation
		reseller, err := u.resellerRepo.FindByID(ctx, *resellerID)
		if err != nil {
			return nil, fmt.Errorf("failed to find reseller for credit limit validation: %w", err)
		}

		// 1. Get all other outstanding orders for this reseller (UNPAID or PARTIAL_PAID status)
		// Custom query params to pull all outstanding reseller orders
		var params request.PaginationParam
		params.Limit = 1000 // Pull a large amount to sum safely
		params.Page = 1

		orders, _, err := u.orderRepo.FindAll(ctx, params, []string{
			orderDomain.StatusInProduction,
			orderDomain.StatusReadyForPickup,
		}, nil, nil, nil, nil, "", nil, nil, "")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch outstanding orders for credit limit validation: %w", err)
		}

		var outstandingDebt float64
		for _, o := range orders {
			if o.ResellerID != nil && *o.ResellerID == *resellerID {
				if o.PaymentStatus == orderDomain.PaymentStatusUnpaid || o.PaymentStatus == orderDomain.PaymentStatusPartialPaid {
					outstandingDebt += (o.GrandTotal - o.AmountPaid)
				}
			}
		}

		// Calculate potential new debt including this order
		newPotentialDebt := outstandingDebt + (grandTotal - amountPaid)
		if newPotentialDebt > reseller.CreditLimit {
			return nil, fmt.Errorf("Credit limit exceeded. Limit: %.2f, Outstanding: %.2f, New Order Debt: %.2f", 
				reseller.CreditLimit, outstandingDebt, grandTotal-amountPaid)
		}
	}

	// Transition to production
	order.Status = orderDomain.StatusInProduction

	// Generate INV Number
	dateStr := time.Now().Format("20060102")
	seq, err := u.orderRepo.GetNextInvSeq(ctx, dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}
	invNo := fmt.Sprintf("INV/%s/%04d", dateStr, seq)
	order.InvoiceNumber = &invNo

	err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if amountPaid > 0 {
			var paymentType string
			if amountPaid >= grandTotal {
				paymentType = "FULL_PAYMENT"
			} else {
				paymentType = "DOWN_PAYMENT"
			}

			for _, p := range payments {
				if p.AmountPaid > 0 {
					initialPayment := &orderDomain.OrderPayment{
						OrderID:       order.ID,
						CashierID:     cashierID,
						Amount:        p.AmountPaid,
						PaymentMethod: p.PaymentMethod,
						PaymentType:   paymentType,
					}
					if err := tx.Create(initialPayment).Error; err != nil {
						return fmt.Errorf("failed to create payment log: %w", err)
					}

					// Lock CashAccount and update balance
					acc, err := u.cashFlowRepo.FindAccountByNameWithLock(ctx, tx, p.PaymentMethod)
					if err != nil {
						return fmt.Errorf("failed to find cash account %s: %w", p.PaymentMethod, err)
					}
					acc.Balance += p.AmountPaid
					if err := u.cashFlowRepo.UpdateAccount(ctx, tx, acc); err != nil {
						return fmt.Errorf("failed to update cash account balance: %w", err)
					}

					desc := fmt.Sprintf("Pembayaran Order %s (%s)", order.JobNumber, paymentType)
					if order.InvoiceNumber != nil {
						desc = fmt.Sprintf("Pembayaran Invoice %s (%s)", *order.InvoiceNumber, paymentType)
					}
					cf := &cashFlowDomain.CashFlow{
						ID:              uuid.New(),
						TransactionDate: time.Now(),
						ReferenceType:   cashFlowDomain.RefOrderPayment,
						ReferenceID:     &initialPayment.ID,
						Type:            cashFlowDomain.TypeDebit,
						Amount:          p.AmountPaid,
						PaymentMethod:   p.PaymentMethod,
						Description:     &desc,
						CustomerName:    order.CustomerName,
						InvoiceNumber:   order.InvoiceNumber,
						CashierID:       cashierID,
					}
					if err := u.cashFlowRepo.CreateTx(ctx, tx, cf); err != nil {
						return fmt.Errorf("failed to create cash flow record: %w", err)
					}
				}
			}
		}

		if err := tx.Save(order).Error; err != nil {
			return fmt.Errorf("failed to save payment: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Fetch updated order to preload relation like Cashier
	updatedOrder, err := u.orderRepo.FindByID(ctx, order.ID)
	if err == nil {
		return updatedOrder, nil
	}

	return order, nil
}

func (u *orderUsecase) CreateFinishing(ctx context.Context, finishing *orderDomain.Finishing) error {
	return u.orderRepo.CreateFinishing(ctx, finishing)
}

func (u *orderUsecase) FindAllFinishings(ctx context.Context) ([]orderDomain.Finishing, error) {
	return u.orderRepo.FindAllFinishings(ctx)
}

func (u *orderUsecase) GetSPKByID(ctx context.Context, id uuid.UUID) (*orderDomain.Order, error) {
	return u.orderRepo.FindByIDWithCategoryPreload(ctx, id)
}

func (u *orderUsecase) Repay(
	ctx context.Context,
	orderID uuid.UUID,
	cashierID uuid.UUID,
	payments []orderDomain.PaymentItem,
) (*orderDomain.Order, error) {
	order, err := u.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Order must not be in DRAFT or PENDING_PAYMENT
	if order.Status == orderDomain.StatusDraft || order.Status == orderDomain.StatusPendingPayment {
		return nil, fmt.Errorf("cannot process repayment for order in %s status", order.Status)
	}

	var amountPaid float64
	for _, p := range payments {
		if p.AmountPaid <= 0 {
			return nil, errors.New("payment amount must be greater than 0")
		}
		amountPaid += p.AmountPaid
	}

	if amountPaid <= 0 {
		return nil, errors.New("total payment amount must be greater than 0")
	}

	remainingDebt := order.GrandTotal - order.AmountPaid
	if amountPaid > remainingDebt {
		return nil, fmt.Errorf("payment amount (%.2f) exceeds remaining debt (%.2f)", amountPaid, remainingDebt)
	}

	// Update order amount paid
	order.AmountPaid += amountPaid

	// Recalculate Payment Status
	if order.AmountPaid >= order.GrandTotal {
		order.PaymentStatus = orderDomain.PaymentStatusPaid
	} else {
		order.PaymentStatus = orderDomain.PaymentStatusPartialPaid
	}

	err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create payment logs
		for _, p := range payments {
			if p.AmountPaid > 0 {
				payment := &orderDomain.OrderPayment{
					OrderID:       order.ID,
					CashierID:     cashierID,
					Amount:        p.AmountPaid,
					PaymentMethod: p.PaymentMethod,
					PaymentType:   "REPAYMENT",
				}
				if err := tx.Create(payment).Error; err != nil {
					return fmt.Errorf("failed to create payment log: %w", err)
				}

				// Lock CashAccount and update balance
				acc, err := u.cashFlowRepo.FindAccountByNameWithLock(ctx, tx, p.PaymentMethod)
				if err != nil {
					return fmt.Errorf("failed to find cash account %s: %w", p.PaymentMethod, err)
				}
				acc.Balance += p.AmountPaid
				if err := u.cashFlowRepo.UpdateAccount(ctx, tx, acc); err != nil {
					return fmt.Errorf("failed to update cash account balance: %w", err)
				}

				// Create cash flow entry (General Ledger)
				desc := fmt.Sprintf("Pelunasan Order %s", order.JobNumber)
				if order.InvoiceNumber != nil {
					desc = fmt.Sprintf("Pelunasan Invoice %s", *order.InvoiceNumber)
				}
				cf := &cashFlowDomain.CashFlow{
					ID:              uuid.New(),
					TransactionDate: time.Now(),
					ReferenceType:   cashFlowDomain.RefOrderPayment,
					ReferenceID:     &payment.ID,
					Type:            cashFlowDomain.TypeDebit,
					Amount:          p.AmountPaid,
					PaymentMethod:   p.PaymentMethod,
					Description:     &desc,
					CustomerName:    order.CustomerName,
					InvoiceNumber:   order.InvoiceNumber,
					CashierID:       cashierID,
				}
				if err := u.cashFlowRepo.CreateTx(ctx, tx, cf); err != nil {
					return fmt.Errorf("failed to create cash flow record: %w", err)
				}
			}
		}

		if err := tx.Save(order).Error; err != nil {
			return fmt.Errorf("failed to save repayment: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Fetch updated order to preload relation like Cashier & Payments
	updatedOrder, err := u.orderRepo.FindByID(ctx, order.ID)
	if err == nil {
		return updatedOrder, nil
	}

	return order, nil
}

func (u *orderUsecase) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus string) (*orderDomain.Order, error) {
	order, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	currentStatus := order.Status
	if currentStatus == newStatus {
		return order, nil
	}

	allowed := false
	switch currentStatus {
	case orderDomain.StatusDraft:
		if newStatus == orderDomain.StatusPendingPayment || newStatus == orderDomain.StatusCancelled {
			allowed = true
		}
	case orderDomain.StatusPendingPayment:
		if newStatus == orderDomain.StatusDraft || newStatus == orderDomain.StatusCancelled {
			allowed = true
		}
	case orderDomain.StatusInProduction:
		if newStatus == orderDomain.StatusReadyForPickup || newStatus == orderDomain.StatusCancelled {
			allowed = true
		}
	case orderDomain.StatusReadyForPickup:
		if newStatus == orderDomain.StatusCompleted || newStatus == orderDomain.StatusCancelled {
			allowed = true
		}
	}

	if !allowed {
		return nil, fmt.Errorf("invalid status transition from %s to %s", currentStatus, newStatus)
	}

	if newStatus == orderDomain.StatusCompleted {
		if order.PaymentStatus != orderDomain.PaymentStatusPaid {
			return nil, fmt.Errorf("cannot complete order: payment status is %s, must be PAID", order.PaymentStatus)
		}
	}

	order.Status = newStatus
	if err := u.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	updatedOrder, err := u.orderRepo.FindByID(ctx, order.ID)
	if err == nil {
		return updatedOrder, nil
	}
	return order, nil
}

func (u *orderUsecase) UpdateDraft(ctx context.Context, id uuid.UUID, orderReq *orderDomain.Order) (*orderDomain.Order, error) {
	order, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	if order.Status != orderDomain.StatusDraft {
		return nil, fmt.Errorf("cannot update order in %s status, only DRAFT status allowed", order.Status)
	}

	order.ResellerID = orderReq.ResellerID
	order.CustomerName = orderReq.CustomerName
	order.CustomerPhone = orderReq.CustomerPhone
	order.Notes = orderReq.Notes

	var customerLevelID uuid.UUID
	if order.ResellerID != nil && *order.ResellerID != uuid.Nil {
		reseller, err := u.resellerRepo.FindByID(ctx, *order.ResellerID)
		if err != nil {
			return nil, fmt.Errorf("failed to find reseller: %w", err)
		}
		if reseller.CustomerLevelID != nil {
			customerLevelID = *reseller.CustomerLevelID
		} else {
			customerLevelID = uuid.MustParse("d2c67ef8-82e4-4d8b-968b-5a1e2f5b6154")
		}
	} else {
		customerLevelID = uuid.MustParse("b3c8f3a3-b26a-4638-b7f2-841a54774844")
	}

	var totalProductPrice float64
	var totalAdditionalCost float64

	for i := range orderReq.OrderItems {
		item := &orderReq.OrderItems[i]

		calcQty, err := u.validateUOMAndGetQty(item)
		if err != nil {
			return nil, fmt.Errorf("item[%d]: %w", i, err)
		}

		var finishingIDs []uuid.UUID
		for _, f := range item.Finishings {
			finishingIDs = append(finishingIDs, f.ID)
		}

		var finishings []orderDomain.Finishing
		if len(finishingIDs) > 0 {
			var err error
			finishings, err = u.orderRepo.FindFinishingsByIDs(ctx, finishingIDs)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch finishings for item %d: %w", i, err)
			}
		}
		item.Finishings = finishings

		var finishingCost float64
		for _, f := range finishings {
			finishingCost += f.Price
		}

		qtyInt := int(calcQty)
		if qtyInt < 1 {
			qtyInt = 1
		}

		priceRes, err := u.productRepo.CheckPrice(ctx, item.ProductVariantID, customerLevelID, qtyInt)
		if err != nil {
			return nil, fmt.Errorf("failed to check price tier for variant %s: %w", item.ProductVariantID, err)
		}

		item.PricePerUnit = priceRes.PricePerUnit
		item.AdditionalCost = finishingCost
		item.Subtotal = (priceRes.PricePerUnit * calcQty) + (finishingCost * float64(item.Quantity))

		totalProductPrice += priceRes.PricePerUnit * calcQty
		totalAdditionalCost += finishingCost * float64(item.Quantity)
	}

	order.TotalProductPrice = totalProductPrice
	order.TotalAdditionalCost = totalAdditionalCost
	order.GrandTotal = totalProductPrice + totalAdditionalCost
	order.OrderItems = orderReq.OrderItems

	if err := u.orderRepo.ReplaceItems(ctx, order.ID, order.OrderItems); err != nil {
		return nil, fmt.Errorf("failed to update order items: %w", err)
	}

	if err := u.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	updatedOrder, err := u.orderRepo.FindByID(ctx, order.ID)
	if err == nil {
		return updatedOrder, nil
	}
	return order, nil
}

func (u *orderUsecase) GetReportsWidgets(ctx context.Context, statuses []string, paymentStatuses []string, paymentMethods []string, designerID *uuid.UUID, cashierID *uuid.UUID, search string, startDate *time.Time, endDate *time.Time, customerType string) (*orderDomain.OrderReportsWidgetsRes, error) {
	return u.orderRepo.GetReportsWidgets(ctx, statuses, paymentStatuses, paymentMethods, designerID, cashierID, search, startDate, endDate, customerType)
}

func (u *orderUsecase) GetSalesReportWidgets(ctx context.Context, statuses []string, paymentStatuses []string, paymentMethods []string, designerID *uuid.UUID, cashierID *uuid.UUID, search string, startDate *time.Time, endDate *time.Time, customerType string) (*orderDomain.SalesReportWidgetsRes, error) {
	return u.orderRepo.GetSalesReportWidgets(ctx, statuses, paymentStatuses, paymentMethods, designerID, cashierID, search, startDate, endDate, customerType)
}

func (u *orderUsecase) Refund(
	ctx context.Context,
	id uuid.UUID,
	cashierID uuid.UUID,
	paymentMethod string,
	amount float64,
	reason string,
) (*orderDomain.Order, error) {
	order, err := u.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	if amount <= 0 {
		return nil, errors.New("refund amount must be greater than 0")
	}

	if amount > order.AmountPaid {
		return nil, fmt.Errorf("refund amount (%.2f) exceeds amount paid (%.2f)", amount, order.AmountPaid)
	}

	order.AmountPaid -= amount
	order.Status = orderDomain.StatusRefund

	if order.AmountPaid <= 0 {
		order.PaymentStatus = orderDomain.PaymentStatusUnpaid
	} else {
		order.PaymentStatus = orderDomain.PaymentStatusPartialPaid
	}

	err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		payment := &orderDomain.OrderPayment{
			OrderID:       order.ID,
			CashierID:     cashierID,
			Amount:        amount,
			PaymentMethod: paymentMethod,
			PaymentType:   "REFUND",
		}
		if err := tx.Create(payment).Error; err != nil {
			return fmt.Errorf("failed to create payment log for refund: %w", err)
		}

		acc, err := u.cashFlowRepo.FindAccountByNameWithLock(ctx, tx, paymentMethod)
		if err != nil {
			return fmt.Errorf("failed to find cash account %s: %w", paymentMethod, err)
		}
		acc.Balance -= amount
		if err := u.cashFlowRepo.UpdateAccount(ctx, tx, acc); err != nil {
			return fmt.Errorf("failed to update cash account balance: %w", err)
		}

		desc := fmt.Sprintf("Refund Order %s", order.JobNumber)
		if order.InvoiceNumber != nil {
			desc = fmt.Sprintf("Refund Invoice %s", *order.InvoiceNumber)
		}
		if reason != "" {
			desc = fmt.Sprintf("%s (%s)", desc, reason)
		}

		cf := &cashFlowDomain.CashFlow{
			ID:              uuid.New(),
			TransactionDate: time.Now(),
			ReferenceType:   cashFlowDomain.RefRefund,
			ReferenceID:     &payment.ID,
			Type:            cashFlowDomain.TypeCredit,
			Amount:          amount,
			PaymentMethod:   paymentMethod,
			Description:     &desc,
			CustomerName:    order.CustomerName,
			InvoiceNumber:   order.InvoiceNumber,
			CashierID:       cashierID,
		}
		if err := u.cashFlowRepo.CreateTx(ctx, tx, cf); err != nil {
			return fmt.Errorf("failed to create cash flow record: %w", err)
		}

		if err := tx.Save(order).Error; err != nil {
			return fmt.Errorf("failed to save refunded order: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	updatedOrder, err := u.orderRepo.FindByID(ctx, order.ID)
	if err == nil {
		return updatedOrder, nil
	}
	return order, nil
}


