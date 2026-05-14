package usecase

import (
	"context"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/reseller/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type resellerUsecase struct {
	resellerRepo domain.ResellerRepository
	logger       *zap.Logger
}

func NewResellerUsecase(resellerRepo domain.ResellerRepository, logger *zap.Logger) domain.ResellerUsecase {
	return &resellerUsecase{
		resellerRepo: resellerRepo,
		logger:       logger,
	}
}

func (u *resellerUsecase) Create(ctx context.Context, reseller *domain.Reseller) error {
	return u.resellerRepo.Create(ctx, reseller)
}

func (u *resellerUsecase) FindByID(ctx context.Context, id uuid.UUID) (*domain.Reseller, error) {
	return u.resellerRepo.FindByID(ctx, id)
}

func (u *resellerUsecase) FindAll(ctx context.Context, params request.PaginationParam, search string) ([]domain.Reseller, int64, error) {
	return u.resellerRepo.FindAll(ctx, params, search)
}

func (u *resellerUsecase) Update(ctx context.Context, reseller *domain.Reseller) error {
	return u.resellerRepo.Update(ctx, reseller)
}

func (u *resellerUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.resellerRepo.Delete(ctx, id)
}
