package usecase

import (
	"context"
	"time"

	expenseDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/expense/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/payable/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"go.uber.org/zap"
)

type payableUsecase struct {
	repo           domain.PayableRepository
	expenseUsecase expenseDomain.ExpenseUsecase
	logger         *zap.Logger
}

func NewPayableUsecase(
	repo domain.PayableRepository,
	expenseUsecase expenseDomain.ExpenseUsecase,
	logger *zap.Logger,
) domain.PayableUsecase {
	return &payableUsecase{
		repo:           repo,
		expenseUsecase: expenseUsecase,
		logger:         logger,
	}
}

func (u *payableUsecase) GetVendorPayables(
	ctx context.Context,
	params request.PaginationParam,
	search string,
	status string,
	startDate *time.Time,
	endDate *time.Time,
) ([]domain.VendorPayable, int64, error) {
	return u.repo.FindVendorPayables(ctx, params, search, status, startDate, endDate)
}

func (u *payableUsecase) GetSettledPayables(
	ctx context.Context,
	params request.PaginationParam,
	search string,
	startDate *time.Time,
	endDate *time.Time,
) ([]domain.SettledExpensePayable, int64, error) {
	return u.repo.FindSettledPayables(ctx, params, search, startDate, endDate)
}

func (u *payableUsecase) GetSummary(ctx context.Context) (*domain.VendorPayableSummary, error) {
	return u.repo.GetSummary(ctx)
}

func (u *payableUsecase) PayExpense(ctx context.Context, input domain.PayVendorExpenseInput) (*domain.VendorPayable, error) {
	payments := make([]expenseDomain.ExpensePayment, len(input.Payments))
	for i, p := range input.Payments {
		payments[i] = expenseDomain.ExpensePayment{
			Amount:        p.Amount,
			PaymentMethod: p.PaymentMethod,
			PaymentDate:   time.Now(),
			CashierID:     input.CashierID,
		}
	}

	if err := u.expenseUsecase.PayInstallment(ctx, input.ExpenseID, input.CashierID, payments); err != nil {
		return nil, err
	}

	return u.repo.FindVendorPayableByExpenseID(ctx, input.ExpenseID)
}

