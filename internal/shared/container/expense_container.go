package container

import (
	cfRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/repository"
	expenseHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/expense/delivery/http"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/expense/repository"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/expense/usecase"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newExpenseHandler(db *gorm.DB, logger *zap.Logger) *expenseHttp.ExpenseHandler {
	repo := repository.NewExpenseRepository(db)
	cfRepository := cfRepo.NewCashFlowRepository(db)
	expenseUsecase := usecase.NewExpenseUsecase(repo, cfRepository, db, logger)
	return expenseHttp.NewExpenseHandler(expenseUsecase)
}
