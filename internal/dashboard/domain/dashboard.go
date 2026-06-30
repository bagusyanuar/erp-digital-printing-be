package domain

import (
	"context"
	"time"
)

type DashboardWidgetsRes struct {
	TotalOmzet      float64 `json:"total_omzet"`
	TotalPendapatan float64 `json:"total_pendapatan"`
	TotalPengeluaran float64 `json:"total_pengeluaran"`
	VolumeTransaksi int64   `json:"volume_transaksi"`
}

type DashboardRepository interface {
	GetWidgets(ctx context.Context, startDate *time.Time, endDate *time.Time) (*DashboardWidgetsRes, error)
}

type DashboardUsecase interface {
	GetWidgets(ctx context.Context, startDate *time.Time, endDate *time.Time) (*DashboardWidgetsRes, error)
}
