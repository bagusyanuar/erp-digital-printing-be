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

func (u *cashFlowUsecase) GetReport(ctx context.Context, startDate time.Time, endDate time.Time) (*domain.CashFlowReportRes, error) {
	cashFlows, err := u.repo.FindAll(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var totalDebit float64
	var totalCredit float64

	detailsByMethod := map[string]domain.CashFlowMethodDetail{
		"cash":     {Debit: 0, Credit: 0, Balance: 0},
		"transfer": {Debit: 0, Credit: 0, Balance: 0},
		"qris":     {Debit: 0, Credit: 0, Balance: 0},
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

		// Calculate total summary
		if cf.Type == domain.TypeDebit {
			totalDebit += cf.Amount
		} else if cf.Type == domain.TypeCredit {
			totalCredit += cf.Amount
		}

		// Calculate by method
		method := cf.PaymentMethod
		detail, exists := detailsByMethod[method]
		if !exists {
			detail = domain.CashFlowMethodDetail{Debit: 0, Credit: 0, Balance: 0}
		}

		if cf.Type == domain.TypeDebit {
			detail.Debit += cf.Amount
		} else if cf.Type == domain.TypeCredit {
			detail.Credit += cf.Amount
		}
		detail.Balance = detail.Debit - detail.Credit
		detailsByMethod[method] = detail
	}

	return &domain.CashFlowReportRes{
		Summary: domain.CashFlowSummary{
			TotalDebit:  totalDebit,
			TotalCredit: totalCredit,
			NetBalance:  totalDebit - totalCredit,
		},
		DetailsByMethod: detailsByMethod,
		Transactions:     transactions,
	}, nil
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

