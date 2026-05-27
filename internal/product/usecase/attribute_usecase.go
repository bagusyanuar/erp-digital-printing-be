package usecase

import (
	"context"
	"regexp"
	"strings"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/product/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type attributeUsecase struct {
	attributeRepo domain.AttributeRepository
	logger        *zap.Logger
}

func NewAttributeUsecase(attributeRepo domain.AttributeRepository, logger *zap.Logger) domain.AttributeUsecase {
	return &attributeUsecase{
		attributeRepo: attributeRepo,
		logger:        logger,
	}
}

func (u *attributeUsecase) Create(ctx context.Context, attribute *domain.Attribute) error {
	attribute.Code = slugify(attribute.Name)
	return u.attributeRepo.Create(ctx, attribute)
}

func (u *attributeUsecase) FindByID(ctx context.Context, id uuid.UUID) (*domain.Attribute, error) {
	return u.attributeRepo.FindByID(ctx, id)
}

func (u *attributeUsecase) FindAll(ctx context.Context, params request.PaginationParam, search string) ([]domain.Attribute, int64, error) {
	return u.attributeRepo.FindAll(ctx, params, search)
}

func (u *attributeUsecase) Update(ctx context.Context, attribute *domain.Attribute) error {
	attribute.Code = slugify(attribute.Name)
	return u.attributeRepo.Update(ctx, attribute)
}

func (u *attributeUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.attributeRepo.Delete(ctx, id)
}

func slugify(str string) string {
	str = strings.ToLower(str)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	str = reg.ReplaceAllString(str, "_")
	str = strings.Trim(str, "_")
	return str
}
