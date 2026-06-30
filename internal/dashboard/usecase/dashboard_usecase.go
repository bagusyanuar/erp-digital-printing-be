package usecase

import (
	"context"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/dashboard/domain"
	"go.uber.org/zap"
)

type dashboardUsecase struct {
	dashboardRepo domain.DashboardRepository
	logger        *zap.Logger
}

func NewDashboardUsecase(dashboardRepo domain.DashboardRepository, logger *zap.Logger) domain.DashboardUsecase {
	return &dashboardUsecase{
		dashboardRepo: dashboardRepo,
		logger:        logger,
	}
}

func (u *dashboardUsecase) GetWidgets(ctx context.Context, startDate *time.Time, endDate *time.Time) (*domain.DashboardWidgetsRes, error) {
	return u.dashboardRepo.GetWidgets(ctx, startDate, endDate)
}
