package usecase

import (
	"context"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/product/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type productUsecase struct {
	productRepo domain.ProductRepository
	logger      *zap.Logger
}

func NewProductUsecase(productRepo domain.ProductRepository, logger *zap.Logger) domain.ProductUsecase {
	return &productUsecase{
		productRepo: productRepo,
		logger:      logger,
	}
}

func (u *productUsecase) Create(ctx context.Context, product *domain.Product) error {
	// Auto create default variant
	product.Variants = []domain.ProductVariant{
		{
			VariantName:    "Default",
			AdditionalCost: 0,
			IsDefault:      true,
		},
	}

	return u.productRepo.Create(ctx, product)
}

func (u *productUsecase) FindByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	return u.productRepo.FindByID(ctx, id)
}

func (u *productUsecase) FindAll(ctx context.Context, params request.PaginationParam, search string, categoryID *uuid.UUID) ([]domain.Product, int64, error) {
	return u.productRepo.FindAll(ctx, params, search, categoryID)
}

func (u *productUsecase) Update(ctx context.Context, product *domain.Product) error {
	return u.productRepo.Update(ctx, product)
}

func (u *productUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.productRepo.Delete(ctx, id)
}

func (u *productUsecase) CreateVariant(ctx context.Context, variant *domain.ProductVariant) error {
	return u.productRepo.CreateVariant(ctx, variant)
}
