package container

import (
	cfRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/repository"
	expenseHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/expense/delivery/http"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/expense/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/expense/repository"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/expense/usecase"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newExpenseUsecase(db *gorm.DB, logger *zap.Logger) domain.ExpenseUsecase {
	repo := repository.NewExpenseRepository(db)
	cfRepository := cfRepo.NewCashFlowRepository(db)
	return usecase.NewExpenseUsecase(repo, cfRepository, db, logger)
}

func newExpenseHandler(expenseUsecase domain.ExpenseUsecase) *expenseHttp.ExpenseHandler {
	return expenseHttp.NewExpenseHandler(expenseUsecase)
}
