package container

import (
	"github.com/bagusyanuar/erp-digital-printing-be/internal/reseller/delivery/http"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/reseller/repository"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/reseller/usecase"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newResellerHandler(db *gorm.DB, logger *zap.Logger) *http.ResellerHandler {
	resellerRepo := repository.NewResellerRepository(db)
	resellerUsecase := usecase.NewResellerUsecase(resellerRepo, logger)
	return http.NewResellerHandler(resellerUsecase)
}
