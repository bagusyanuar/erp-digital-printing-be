package container

import (
	dashboardHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/dashboard/delivery/http"
	dashboardRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/dashboard/repository"
	dashboardUseCase "github.com/bagusyanuar/erp-digital-printing-be/internal/dashboard/usecase"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newDashboardHandler(db *gorm.DB, logger *zap.Logger) *dashboardHttp.DashboardHandler {
	repo := dashboardRepo.NewDashboardRepository(db)
	usecase := dashboardUseCase.NewDashboardUsecase(repo, logger)
	return dashboardHttp.NewDashboardHandler(usecase)
}
