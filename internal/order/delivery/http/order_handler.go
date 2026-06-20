package http

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/order/delivery/http/dto"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/order/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type OrderHandler struct {
	orderUsecase domain.OrderUsecase
}

func NewOrderHandler(orderUsecase domain.OrderUsecase) *OrderHandler {
	return &OrderHandler{orderUsecase: orderUsecase}
}

// buildOrderFromReq maps a CreateOrderReq DTO into a domain Order entity.
// Extracted to eliminate duplication between SaveDraft and SubmitToCashier handlers.
func buildOrderFromReq(req *dto.CreateOrderReq) *domain.Order {
	order := &domain.Order{
		DesignerID:    req.DesignerID,
		ResellerID:    req.ResellerID,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		Notes:         req.Notes,
	}

	order.OrderItems = make([]domain.OrderItem, len(req.Items))
	for i, itemReq := range req.Items {
		item := domain.OrderItem{
			ProductVariantID: itemReq.ProductVariantID,
			UOM:              itemReq.UOM,
			LengthCM:         itemReq.LengthCM,
			WidthCM:          itemReq.WidthCM,
			Quantity:         itemReq.Quantity,
			DesignFileURL:    itemReq.DesignFileURL,
			ProductionNotes:  itemReq.ProductionNotes,
		}

		if len(itemReq.FinishingIDs) > 0 {
			item.Finishings = make([]domain.Finishing, len(itemReq.FinishingIDs))
			for j, fid := range itemReq.FinishingIDs {
				item.Finishings[j] = domain.Finishing{ID: fid}
			}
		}

		order.OrderItems[i] = item
	}

	return order
}

func (h *OrderHandler) SaveDraft(c fiber.Ctx) error {
	designerID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized: Designer ID not found in context", nil)
	}

	var req dto.CreateOrderReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	order := buildOrderFromReq(&req)
	order.DesignerID = designerID

	if err := h.orderUsecase.SaveDraft(c.Context(), order); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save draft order", err.Error())
	}

	orderFull, err := h.orderUsecase.FindByID(c.Context(), order.ID)
	if err != nil {
		return response.Created(c, "Draft order saved successfully", mapOrderToRes(order))
	}

	return response.Created(c, "Draft order saved successfully", mapOrderToRes(orderFull))
}

func (h *OrderHandler) SubmitToCashier(c fiber.Ctx) error {
	designerID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized: Designer ID not found in context", nil)
	}

	var req dto.CreateOrderReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	order := buildOrderFromReq(&req)
	order.DesignerID = designerID

	if err := h.orderUsecase.SubmitToCashier(c.Context(), order); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to submit order to cashier", err.Error())
	}

	orderFull, err := h.orderUsecase.FindByID(c.Context(), order.ID)
	if err != nil {
		return response.Created(c, "Order submitted to cashier successfully", mapOrderToRes(order))
	}

	return response.Created(c, "Order submitted to cashier successfully", mapOrderToRes(orderFull))
}

func (h *OrderHandler) SubmitExistingToCashier(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid order ID", err.Error())
	}

	if err := h.orderUsecase.SubmitExistingToCashier(c.Context(), id); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to submit order", err.Error())
	}

	order, err := h.orderUsecase.FindByID(c.Context(), id)
	if err != nil {
		return response.Success[any](c, "Order submitted successfully", nil, nil)
	}

	return response.Success(c, "Order submitted successfully", mapOrderToRes(order), nil)
}

func (h *OrderHandler) ProcessPayment(c fiber.Ctx) error {
	cashierID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized: Cashier ID not found in context", nil)
	}

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid order ID", err.Error())
	}

	var req dto.PaymentProcessReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	var payments []domain.PaymentItem
	for _, p := range req.Payments {
		payments = append(payments, domain.PaymentItem{
			PaymentMethod: p.PaymentMethod,
			AmountPaid:    p.AmountPaid,
		})
	}

	order, err := h.orderUsecase.ProcessPayment(
		c.Context(),
		id,
		cashierID,
		req.ResellerID,
		req.CustomerName,
		req.CustomerPhone,
		payments,
	)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to process payment", err.Error())
	}

	return response.Success(c, "Payment processed and order sent to production", mapOrderToRes(order), nil)
}

func (h *OrderHandler) FindByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid order ID", err.Error())
	}

	order, err := h.orderUsecase.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Order not found", err.Error())
	}

	return response.Success(c, "Order fetched successfully", mapOrderToRes(order), nil)
}

func (h *OrderHandler) FindAll(c fiber.Ctx) error {
	var params request.PaginationParam
	if err := c.Bind().Query(&params); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid query parameters", err.Error())
	}

	var statuses []string
	if statusQuery := c.Query("status", ""); statusQuery != "" {
		parts := strings.Split(statusQuery, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				statuses = append(statuses, trimmed)
			}
		}
	}

	var paymentStatuses []string
	if paymentStatusQuery := c.Query("payment_status", ""); paymentStatusQuery != "" {
		parts := strings.Split(paymentStatusQuery, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				paymentStatuses = append(paymentStatuses, trimmed)
			}
		}
	}

	var designerID *uuid.UUID
	if designerQuery := c.Query("designer_id", ""); designerQuery != "" {
		if did, err := uuid.Parse(designerQuery); err == nil {
			designerID = &did
		}
	}

	var cashierID *uuid.UUID
	if cashierQuery := c.Query("cashier_id", ""); cashierQuery != "" {
		if cid, err := uuid.Parse(cashierQuery); err == nil {
			cashierID = &cid
		}
	}

	search := c.Query("search", "")
	startDateStr := c.Query("start_date", "")
	endDateStr := c.Query("end_date", "")
	customerType := c.Query("customer_type", "")

	var startDate, endDate *time.Time
	if (startDateStr == "" && endDateStr != "") || (startDateStr != "" && endDateStr == "") {
		return response.Error(c, fiber.StatusBadRequest, "Invalid date range", "Both start_date and end_date must be provided together")
	}

	if startDateStr != "" && endDateStr != "" {
		t1, err1 := time.Parse("2006-01-02", startDateStr)
		t2, err2 := time.Parse("2006-01-02", endDateStr)
		if err1 != nil || err2 != nil {
			return response.Error(c, fiber.StatusBadRequest, "Invalid date format", "Dates must be in YYYY-MM-DD format")
		}
		t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
		t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 23, 59, 59, 999999999, t2.Location())
		startDate = &t1
		endDate = &t2
	}

	orders, total, err := h.orderUsecase.FindAll(c.Context(), params, statuses, paymentStatuses, designerID, cashierID, search, startDate, endDate, customerType)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch orders", err.Error())
	}

	resList := make([]dto.OrderRes, len(orders))
	for i, o := range orders {
		resList[i] = mapOrderToRes(&o)
	}

	meta := response.Meta{
		Pagination: &response.Pagination{
			TotalItems:  total,
			TotalPages:  int(math.Ceil(float64(total) / float64(params.GetLimit()))),
			CurrentPage: params.GetPage(),
			Limit:       params.GetLimit(),
		},
	}

	return response.Success(c, "Orders fetched successfully", resList, meta)
}

func (h *OrderHandler) CreateFinishing(c fiber.Ctx) error {
	var req dto.CreateFinishingReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	finishing := &domain.Finishing{
		Name:  req.Name,
		Price: req.Price,
	}

	if err := h.orderUsecase.CreateFinishing(c.Context(), finishing); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create finishing", err.Error())
	}

	res := dto.FinishingRes{
		ID:    finishing.ID,
		Name:  finishing.Name,
		Price: finishing.Price,
	}

	return response.Created(c, "Finishing created successfully", res)
}

func (h *OrderHandler) FindAllFinishings(c fiber.Ctx) error {
	finishings, err := h.orderUsecase.FindAllFinishings(c.Context())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch finishings", err.Error())
	}

	resList := make([]dto.FinishingRes, len(finishings))
	for i, f := range finishings {
		resList[i] = dto.FinishingRes{
			ID:    f.ID,
			Name:  f.Name,
			Price: f.Price,
		}
	}

	return response.Success(c, "Finishings fetched successfully", resList, nil)
}

// mapOrderToRes converts a domain Order into an OrderRes DTO for API response.
func mapOrderToRes(o *domain.Order) dto.OrderRes {
	res := dto.OrderRes{
		ID:                  o.ID,
		JobNumber:            o.JobNumber,
		InvoiceNumber:        o.InvoiceNumber,
		ResellerID:           o.ResellerID,
		DesignerID:           o.DesignerID,
		CashierID:            o.CashierID,
		CustomerName:         o.CustomerName,
		CustomerPhone:        o.CustomerPhone,
		Status:               o.Status,
		PaymentStatus:        o.PaymentStatus,
		Notes:                o.Notes,
		TotalAdditionalCost:  o.TotalAdditionalCost,
		TotalProductPrice:    o.TotalProductPrice,
		GrandTotal:           o.GrandTotal,
		AmountPaid:           o.AmountPaid,
		CreatedAt:            o.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:            o.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if o.Reseller != nil {
		res.ResellerName = &o.Reseller.Name
		res.Reseller = &dto.ResellerRes{
			ID:          o.Reseller.ID,
			Name:        o.Reseller.Name,
			Phone:       o.Reseller.Phone,
			Email:       o.Reseller.Email,
			Address:     o.Reseller.Address,
			CreditLimit: o.Reseller.CreditLimit,
		}
	}

	if o.Designer != nil {
		res.DesignerName = o.Designer.Username
	}

	if o.Cashier != nil {
		res.CashierName = &o.Cashier.Username
	}

	if len(o.OrderItems) > 0 {
		res.OrderItems = make([]dto.OrderItemRes, len(o.OrderItems))
		for i, item := range o.OrderItems {
			itemRes := dto.OrderItemRes{
				ID:               item.ID,
				ProductVariantID: item.ProductVariantID,
				UOM:              item.UOM,
				LengthCM:         item.LengthCM,
				WidthCM:          item.WidthCM,
				Quantity:         item.Quantity,
				DesignFileURL:    item.DesignFileURL,
				ProductionNotes:  item.ProductionNotes,
				PricePerUnit:     item.PricePerUnit,
				AdditionalCost:   item.AdditionalCost,
				Subtotal:         item.Subtotal,
			}

			// Safe nil checks to prevent panic on unloaded relations
			if item.ProductVariant != nil {
				itemRes.VariantName = item.ProductVariant.VariantName
				if item.ProductVariant.Product.Name != "" {
					itemRes.ProductName = item.ProductVariant.Product.Name
				}
			}

			if len(item.Finishings) > 0 {
				itemRes.Finishings = make([]dto.FinishingRes, len(item.Finishings))
				for j, f := range item.Finishings {
					itemRes.Finishings[j] = dto.FinishingRes{
						ID:    f.ID,
						Name:  f.Name,
						Price: f.Price,
					}
				}
			}

			res.OrderItems[i] = itemRes
		}
	}

	if len(o.OrderPayments) > 0 {
		// Sort by CreatedAt ascending
		sort.Slice(o.OrderPayments, func(i, j int) bool {
			return o.OrderPayments[i].CreatedAt.Before(o.OrderPayments[j].CreatedAt)
		})

		res.OrderPayments = make([]dto.OrderPaymentRes, len(o.OrderPayments))
		batchMap := make(map[string]int)
		currentBatchNum := 0

		for i, op := range o.OrderPayments {
			timeKey := op.CreatedAt.Format("2006-01-02 15:04:05")
			batchNum, exists := batchMap[timeKey]
			if !exists {
				currentBatchNum++
				batchMap[timeKey] = currentBatchNum
				batchNum = currentBatchNum
			}

			opRes := dto.OrderPaymentRes{
				ID:            op.ID,
				CashierID:     op.CashierID,
				Amount:        op.Amount,
				PaymentMethod: op.PaymentMethod,
				PaymentType:   op.PaymentType,
				PaymentNumber: batchNum,
				CreatedAt:     op.CreatedAt.Format("2006-01-02 15:04:05"),
			}
			if op.Cashier != nil {
				opRes.CashierName = op.Cashier.Username
			}
			res.OrderPayments[i] = opRes
		}
	}

	return res
}

func (h *OrderHandler) GetSPKByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid order ID", err.Error())
	}

	order, err := h.orderUsecase.GetSPKByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Order not found", err.Error())
	}

	// Grouping logic by Category
	categoryMap := make(map[uuid.UUID]*dto.SPKByCategoryRes)
	var categoryOrder []uuid.UUID

	for _, item := range order.OrderItems {
		var catID uuid.UUID
		catName := "Uncategorized"

		if item.ProductVariant != nil && item.ProductVariant.Product.CategoryID != uuid.Nil {
			catID = item.ProductVariant.Product.CategoryID
			if item.ProductVariant.Product.Category.Name != "" {
				catName = item.ProductVariant.Product.Category.Name
			}
		} else {
			catID = uuid.Nil
		}

		spkItem := dto.SPKItemRes{
			ID:              item.ID,
			UOM:              item.UOM,
			LengthCM:         item.LengthCM,
			WidthCM:          item.WidthCM,
			Quantity:         item.Quantity,
			DesignFileURL:    item.DesignFileURL,
			ProductionNotes:  item.ProductionNotes,
		}

		if item.ProductVariant != nil {
			spkItem.VariantName = item.ProductVariant.VariantName
			spkItem.ProductName = item.ProductVariant.Product.Name
		}

		if len(item.Finishings) > 0 {
			spkItem.Finishings = make([]dto.FinishingRes, len(item.Finishings))
			for j, f := range item.Finishings {
				spkItem.Finishings[j] = dto.FinishingRes{
					ID:    f.ID,
					Name:  f.Name,
					Price: f.Price,
				}
			}
		}

		if group, exists := categoryMap[catID]; exists {
			group.Items = append(group.Items, spkItem)
		} else {
			categoryMap[catID] = &dto.SPKByCategoryRes{
				CategoryID:   catID,
				CategoryName: catName,
				Items:        []dto.SPKItemRes{spkItem},
			}
			categoryOrder = append(categoryOrder, catID)
		}
	}

	spkByCategoryList := make([]dto.SPKByCategoryRes, len(categoryOrder))
	for i, catID := range categoryOrder {
		spkByCategoryList[i] = *categoryMap[catID]
	}

	res := dto.OrderSPKRes{
		OrderID:       order.ID,
		JobNumber:     order.JobNumber,
		InvoiceNumber: order.InvoiceNumber,
		CustomerName:  order.CustomerName,
		CustomerPhone: order.CustomerPhone,
		Status:        order.Status,
		SPKByCategory: spkByCategoryList,
	}

	return response.Success(c, "SPK fetched successfully", res, nil)
}

func (h *OrderHandler) Repay(c fiber.Ctx) error {
	cashierID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized: Cashier ID not found in context", nil)
	}

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid order ID", err.Error())
	}

	var req dto.OrderRepayReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	var payments []domain.PaymentItem
	for _, p := range req.Payments {
		payments = append(payments, domain.PaymentItem{
			PaymentMethod: p.PaymentMethod,
			AmountPaid:    p.AmountPaid,
		})
	}

	order, err := h.orderUsecase.Repay(
		c.Context(),
		id,
		cashierID,
		payments,
	)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to process repayment", err.Error())
	}

	return response.Success(c, "Repayment processed successfully", mapOrderToRes(order), nil)
}

func (h *OrderHandler) GetPaymentsByOrderID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid order ID", err.Error())
	}

	order, err := h.orderUsecase.FindByID(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "Order not found", err.Error())
	}

	if len(order.OrderPayments) > 0 {
		sort.Slice(order.OrderPayments, func(i, j int) bool {
			return order.OrderPayments[i].CreatedAt.Before(order.OrderPayments[j].CreatedAt)
		})
	}

	resList := make([]dto.OrderPaymentRes, len(order.OrderPayments))
	batchMap := make(map[string]int)
	currentBatchNum := 0

	for i, op := range order.OrderPayments {
		timeKey := op.CreatedAt.Format("2006-01-02 15:04:05")
		batchNum, exists := batchMap[timeKey]
		if !exists {
			currentBatchNum++
			batchMap[timeKey] = currentBatchNum
			batchNum = currentBatchNum
		}

		opRes := dto.OrderPaymentRes{
			ID:            op.ID,
			CashierID:     op.CashierID,
			Amount:        op.Amount,
			PaymentMethod: op.PaymentMethod,
			PaymentType:   op.PaymentType,
			PaymentNumber: batchNum,
			CreatedAt:     op.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if op.Cashier != nil {
			opRes.CashierName = op.Cashier.Username
		}
		resList[i] = opRes
	}

	return response.Success(c, "Order payments fetched successfully", resList, nil)
}

func (h *OrderHandler) UpdateStatus(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid order ID", err.Error())
	}

	var req dto.UpdateOrderStatusReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	order, err := h.orderUsecase.UpdateStatus(c.Context(), id, req.Status)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update order status", err.Error())
	}

	return response.Success(c, "Order status updated successfully", mapOrderToRes(order), nil)
}

func (h *OrderHandler) UpdateDraft(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid order ID", err.Error())
	}

	var req dto.CreateOrderReq
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	order := buildOrderFromReq(&req)

	updatedOrder, err := h.orderUsecase.UpdateDraft(c.Context(), id, order)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update order draft", err.Error())
	}

	return response.Success(c, "Draft order updated successfully", mapOrderToRes(updatedOrder), nil)
}

func (h *OrderHandler) GetReportsWidgets(c fiber.Ctx) error {
	var statuses []string
	if statusQuery := c.Query("status", ""); statusQuery != "" {
		parts := strings.Split(statusQuery, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				statuses = append(statuses, trimmed)
			}
		}
	}

	var paymentStatuses []string
	if paymentStatusQuery := c.Query("payment_status", ""); paymentStatusQuery != "" {
		parts := strings.Split(paymentStatusQuery, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				paymentStatuses = append(paymentStatuses, trimmed)
			}
		}
	}

	var designerID *uuid.UUID
	if designerQuery := c.Query("designer_id", ""); designerQuery != "" {
		if did, err := uuid.Parse(designerQuery); err == nil {
			designerID = &did
		}
	}

	var cashierID *uuid.UUID
	if cashierQuery := c.Query("cashier_id", ""); cashierQuery != "" {
		if cid, err := uuid.Parse(cashierQuery); err == nil {
			cashierID = &cid
		}
	}

	search := c.Query("search", "")
	startDateStr := c.Query("start_date", "")
	endDateStr := c.Query("end_date", "")
	customerType := c.Query("customer_type", "")

	var startDate, endDate *time.Time
	if (startDateStr == "" && endDateStr != "") || (startDateStr != "" && endDateStr == "") {
		return response.Error(c, fiber.StatusBadRequest, "Invalid date range", "Both start_date and end_date must be provided together")
	}

	if startDateStr != "" && endDateStr != "" {
		t1, err1 := time.Parse("2006-01-02", startDateStr)
		t2, err2 := time.Parse("2006-01-02", endDateStr)
		if err1 != nil || err2 != nil {
			return response.Error(c, fiber.StatusBadRequest, "Invalid date format", "Dates must be in YYYY-MM-DD format")
		}
		t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
		t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 23, 59, 59, 999999999, t2.Location())
		startDate = &t1
		endDate = &t2
	}

	data, err := h.orderUsecase.GetReportsWidgets(c.Context(), statuses, paymentStatuses, designerID, cashierID, search, startDate, endDate, customerType)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch reports widgets", err.Error())
	}

	res := dto.OrderReportsWidgetsRes{
		OmsetPenjualan:     data.OmsetPenjualan,
		VolumeTransaksi:    data.VolumeTransaksi,
		TotalProdukTerjual: data.TotalProdukTerjual,
		StatusNota: dto.OrderReportsStatusNotaRes{
			Lunas:      data.LunasCount,
			BelumLunas: data.BelumLunasCount,
		},
	}

	return response.Success(c, "Reports widgets fetched successfully", res, nil)
}


