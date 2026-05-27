package usecase

import (
	"context"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/category/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type categoryUsecase struct {
	categoryRepo domain.CategoryRepository
	logger       *zap.Logger
}

func NewCategoryUsecase(categoryRepo domain.CategoryRepository, logger *zap.Logger) domain.CategoryUsecase {
	return &categoryUsecase{
		categoryRepo: categoryRepo,
		logger:       logger,
	}
}

func (u *categoryUsecase) Create(ctx context.Context, category *domain.Category) error {
	return u.categoryRepo.Create(ctx, category)
}

func (u *categoryUsecase) FindByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	return u.categoryRepo.FindByID(ctx, id)
}

func (u *categoryUsecase) FindAll(ctx context.Context, params request.PaginationParam, search string) ([]domain.Category, int64, error) {
	return u.categoryRepo.FindAll(ctx, params, search)
}

func (u *categoryUsecase) Update(ctx context.Context, category *domain.Category) error {
	return u.categoryRepo.Update(ctx, category)
}

func (u *categoryUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.categoryRepo.Delete(ctx, id)
}
