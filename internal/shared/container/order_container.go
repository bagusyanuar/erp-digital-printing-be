package container

import (
	cfRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/repository"
	orderHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/order/delivery/http"
	orderDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/order/domain"
	orderRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/order/repository"
	orderUseCase "github.com/bagusyanuar/erp-digital-printing-be/internal/order/usecase"
	productRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/product/repository"
	resellerRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/reseller/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newOrderUsecase(db *gorm.DB, logger *zap.Logger) orderDomain.OrderUsecase {
	oRepo := orderRepo.NewOrderRepository(db)
	pRepo := productRepo.NewProductRepository(db)
	rRepo := resellerRepo.NewResellerRepository(db)
	cRepo := cfRepo.NewCashFlowRepository(db)

	return orderUseCase.NewOrderUsecase(oRepo, pRepo, rRepo, cRepo, db, logger)
}

func newOrderHandler(orderUsecase orderDomain.OrderUsecase) *orderHttp.OrderHandler {
	return orderHttp.NewOrderHandler(orderUsecase)
}
