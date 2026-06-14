package usecase

import (
	"context"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type cashFlowUsecase struct {
	repo domain.CashFlowRepository
	db   *gorm.DB
}

func NewCashFlowUsecase(repo domain.CashFlowRepository, db *gorm.DB) domain.CashFlowUsecase {
	return &cashFlowUsecase{repo: repo, db: db}
}

func (u *cashFlowUsecase) GetReport(ctx context.Context, filter domain.CashFlowFilter) ([]domain.CashFlowTransactionRes, int64, error) {
	cashFlows, total, err := u.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	transactions := make([]domain.CashFlowTransactionRes, 0, len(cashFlows))

	for _, cf := range cashFlows {
		cashierName := "System"
		if cf.Cashier != nil {
			cashierName = cf.Cashier.Username
		}

		transactions = append(transactions, domain.CashFlowTransactionRes{
			ID:              cf.ID,
			TransactionDate: cf.TransactionDate,
			ReferenceType:   cf.ReferenceType,
			ReferenceID:     cf.ReferenceID,
			Type:            cf.Type,
			Amount:          cf.Amount,
			PaymentMethod:   cf.PaymentMethod,
			Description:     cf.Description,
			CustomerName:    cf.CustomerName,
			CashierName:     cashierName,
		})
	}

	return transactions, total, nil
}

func (u *cashFlowUsecase) CreateAdjustment(ctx context.Context, cashierID uuid.UUID, amount float64, flowType string, paymentMethod string, description string) (*domain.CashFlow, error) {
	desc := description
	cf := &domain.CashFlow{
		ID:              uuid.New(),
		TransactionDate: time.Now(),
		ReferenceType:   domain.RefAdjustment,
		ReferenceID:     nil,
		Type:            flowType,
		Amount:          amount,
		PaymentMethod:   paymentMethod,
		Description:     &desc,
		CashierID:       cashierID,
	}

	err := u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := u.repo.FindAccountByNameWithLock(ctx, tx, paymentMethod)
		if err != nil {
			return err
		}

		if flowType == domain.TypeDebit {
			acc.Balance += amount
		} else {
			acc.Balance -= amount
		}

		if err := u.repo.UpdateAccount(ctx, tx, acc); err != nil {
			return err
		}

		if err := u.repo.CreateTx(ctx, tx, cf); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return cf, nil
}

func (u *cashFlowUsecase) FindAllAccounts(ctx context.Context) ([]domain.CashAccount, error) {
	return u.repo.FindAllAccounts(ctx)
}

func (u *cashFlowUsecase) GetSummary(ctx context.Context, filter domain.CashFlowFilter) (*domain.CashFlowSummaryRes, error) {
	return u.repo.GetSummary(ctx, filter)
}

