package usecase

import (
	"context"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/receivable/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"go.uber.org/zap"
)

type receivableUsecase struct {
	repo   domain.ReceivableRepository
	logger *zap.Logger
}

func NewReceivableUsecase(
	repo domain.ReceivableRepository,
	logger *zap.Logger,
) domain.ReceivableUsecase {
	return &receivableUsecase{
		repo:   repo,
		logger: logger,
	}
}

func (u *receivableUsecase) GetCustomerReceivables(
	ctx context.Context,
	params request.PaginationParam,
	search string,
	status string,
	startDate *time.Time,
	endDate *time.Time,
) ([]domain.CustomerReceivable, int64, error) {
	return u.repo.FindCustomerReceivables(ctx, params, search, status, startDate, endDate)
}

func (u *receivableUsecase) GetSettledInvoices(
	ctx context.Context,
	params request.PaginationParam,
	search string,
	startDate *time.Time,
	endDate *time.Time,
) ([]domain.SettledInvoiceReceivable, int64, error) {
	return u.repo.FindSettledInvoices(ctx, params, search, startDate, endDate)
}

func (u *receivableUsecase) GetSummary(ctx context.Context) (*domain.CustomerReceivableSummary, error) {
	return u.repo.GetSummary(ctx)
}
