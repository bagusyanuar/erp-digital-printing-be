package container

import (
	expenseDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/expense/domain"
	payableHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/payable/delivery/http"
	payableRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/payable/repository"
	payableUseCase "github.com/bagusyanuar/erp-digital-printing-be/internal/payable/usecase"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newPayableHandler(db *gorm.DB, expenseUsecase expenseDomain.ExpenseUsecase, logger *zap.Logger) *payableHttp.PayableHandler {
	repo := payableRepo.NewPayableRepository(db)
	usecase := payableUseCase.NewPayableUsecase(repo, expenseUsecase, logger)
	return payableHttp.NewPayableHandler(usecase)
}

