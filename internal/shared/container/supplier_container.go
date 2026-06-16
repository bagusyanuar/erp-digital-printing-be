package container

import (
	supplierHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/supplier/delivery/http"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/supplier/repository"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/supplier/usecase"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newSupplierHandler(db *gorm.DB, logger *zap.Logger) *supplierHttp.SupplierHandler {
	supplierRepo := repository.NewSupplierRepository(db)
	supplierUsecase := usecase.NewSupplierUsecase(supplierRepo, logger)
	return supplierHttp.NewSupplierHandler(supplierUsecase)
}
