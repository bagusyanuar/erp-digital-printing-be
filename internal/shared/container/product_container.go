package container

import (
	productHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/product/delivery/http"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/product/repository"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/product/usecase"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newAttributeHandler(db *gorm.DB, logger *zap.Logger) *productHttp.AttributeHandler {
	attributeRepo := repository.NewAttributeRepository(db)
	attributeUsecase := usecase.NewAttributeUsecase(attributeRepo, logger)
	return productHttp.NewAttributeHandler(attributeUsecase)
}
