package container

import (
	cHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/capital/delivery/http"
	cRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/capital/repository"
	cUseCase "github.com/bagusyanuar/erp-digital-printing-be/internal/capital/usecase"
	cfRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newCapitalHandler(db *gorm.DB, logger *zap.Logger) *cHttp.CapitalHandler {
	capitalRepo := cRepo.NewCapitalRepository(db)
	cashFlowRepo := cfRepo.NewCashFlowRepository(db)
	usecase := cUseCase.NewCapitalUsecase(capitalRepo, cashFlowRepo, db)
	return cHttp.NewCapitalHandler(usecase)
}
