package container

import (
	cfHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/delivery/http"
	cfRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/repository"
	cfUseCase "github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/usecase"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newCashFlowHandler(db *gorm.DB, logger *zap.Logger) *cfHttp.CashFlowHandler {
	repo := cfRepo.NewCashFlowRepository(db)
	usecase := cfUseCase.NewCashFlowUsecase(repo, db)
	
	fundTransferRepo := cfRepo.NewFundTransferRepository(db)
	fundTransferUsecase := cfUseCase.NewFundTransferUsecase(fundTransferRepo, repo, db)
	
	return cfHttp.NewCashFlowHandler(usecase, fundTransferUsecase)
}
