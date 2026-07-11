package container

import (
	reportHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/report/delivery"
	reportUsecase "github.com/bagusyanuar/erp-digital-printing-be/internal/report/usecase"
	orderRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/order/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newReportHandler(db *gorm.DB, logger *zap.Logger) *reportHttp.ReportHandler {
	oRepo := orderRepo.NewOrderRepository(db)
	rUsecase := reportUsecase.NewReportUsecase(oRepo, logger)
	return reportHttp.NewReportHandler(rUsecase, logger)
}
