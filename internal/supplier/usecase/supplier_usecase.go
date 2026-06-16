package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/supplier/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/request"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type supplierUsecase struct {
	supplierRepo domain.SupplierRepository
	logger       *zap.Logger
}

func NewSupplierUsecase(supplierRepo domain.SupplierRepository, logger *zap.Logger) domain.SupplierUsecase {
	return &supplierUsecase{
		supplierRepo: supplierRepo,
		logger:       logger,
	}
}

func (u *supplierUsecase) Create(ctx context.Context, supplier *domain.Supplier) error {
	if supplier.Name == "" {
		return errors.New("supplier name is required")
	}

	existing, err := u.supplierRepo.FindByName(ctx, supplier.Name)
	if err != nil {
		return fmt.Errorf("failed to check existing supplier: %w", err)
	}
	if existing != nil {
		return errors.New("supplier name already exists")
	}

	return u.supplierRepo.Create(ctx, supplier)
}

func (u *supplierUsecase) FindByID(ctx context.Context, id uuid.UUID) (*domain.Supplier, error) {
	return u.supplierRepo.FindByID(ctx, id)
}

func (u *supplierUsecase) FindAll(ctx context.Context, params request.PaginationParam, search string) ([]domain.Supplier, int64, error) {
	return u.supplierRepo.FindAll(ctx, params, search)
}

func (u *supplierUsecase) Update(ctx context.Context, supplier *domain.Supplier) error {
	if supplier.Name == "" {
		return errors.New("supplier name is required")
	}

	existing, err := u.supplierRepo.FindByName(ctx, supplier.Name)
	if err != nil {
		return fmt.Errorf("failed to check existing supplier: %w", err)
	}
	if existing != nil && existing.ID != supplier.ID {
		return errors.New("supplier name already exists")
	}

	return u.supplierRepo.Update(ctx, supplier)
}

func (u *supplierUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	// Let's first make sure the supplier exists.
	_, err := u.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("supplier not found: %w", err)
	}

	return u.supplierRepo.Delete(ctx, id)
}
