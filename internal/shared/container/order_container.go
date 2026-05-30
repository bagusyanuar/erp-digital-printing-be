package container

import (
	orderHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/order/delivery/http"
	orderRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/order/repository"
	orderUseCase "github.com/bagusyanuar/erp-digital-printing-be/internal/order/usecase"
	productRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/product/repository"
	resellerRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/reseller/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newOrderHandler(db *gorm.DB, logger *zap.Logger) *orderHttp.OrderHandler {
	oRepo := orderRepo.NewOrderRepository(db)
	pRepo := productRepo.NewProductRepository(db)
	rRepo := resellerRepo.NewResellerRepository(db)

	oUsecase := orderUseCase.NewOrderUsecase(oRepo, pRepo, rRepo, logger)
	return orderHttp.NewOrderHandler(oUsecase)
}
