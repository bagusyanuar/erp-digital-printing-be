package container

import (
	orderRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/order/repository"
	reportHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/report/delivery"
	reportRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/report/repository"
	reportUsecase "github.com/bagusyanuar/erp-digital-printing-be/internal/report/usecase"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newReportHandler(db *gorm.DB, logger *zap.Logger) *reportHttp.ReportHandler {
	oRepo := orderRepo.NewOrderRepository(db)
	rRepo := reportRepo.NewReportRepository(db)
	rUsecase := reportUsecase.NewReportUsecase(oRepo, rRepo, logger)
	return reportHttp.NewReportHandler(rUsecase, logger)
}
