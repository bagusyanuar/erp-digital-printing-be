package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	capitalDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/capital/domain"
	cashFlowDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type capitalUsecase struct {
	capitalRepo  capitalDomain.CapitalRepository
	cashFlowRepo cashFlowDomain.CashFlowRepository
	db           *gorm.DB
}

func NewCapitalUsecase(
	capitalRepo capitalDomain.CapitalRepository,
	cashFlowRepo cashFlowDomain.CashFlowRepository,
	db *gorm.DB,
) capitalDomain.CapitalUsecase {
	return &capitalUsecase{
		capitalRepo:  capitalRepo,
		cashFlowRepo: cashFlowRepo,
		db:           db,
	}
}

func (u *capitalUsecase) Create(
	ctx context.Context,
	creatorID uuid.UUID,
	txType string,
	amount float64,
	paymentMethod string,
	description string,
) (*capitalDomain.CapitalTransaction, error) {
	if txType != capitalDomain.CapitalInjection && txType != capitalDomain.CapitalWithdrawal {
		return nil, errors.New("invalid capital transaction type")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}

	var desc *string
	if description != "" {
		desc = &description
	}

	capitalTx := &capitalDomain.CapitalTransaction{
		ID:              uuid.New(),
		TransactionDate: time.Now(),
		Type:            txType,
		Amount:          amount,
		PaymentMethod:   paymentMethod,
		Description:     desc,
		CreatedBy:       creatorID,
	}

	err := u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Acquire pessimistic lock on the target account
		acc, err := u.cashFlowRepo.FindAccountByNameWithLock(ctx, tx, paymentMethod)
		if err != nil {
			return fmt.Errorf("failed to lock cash account: %w", err)
		}

		var cashFlowType string
		var refType string

		if txType == capitalDomain.CapitalInjection {
			acc.Balance += amount
			cashFlowType = cashFlowDomain.TypeDebit
			refType = "CAPITAL_INJECTION"
		} else {
			if acc.Balance < amount {
				return errors.New("insufficient cash account balance")
			}
			acc.Balance -= amount
			cashFlowType = cashFlowDomain.TypeCredit
			refType = "CAPITAL_WITHDRAWAL"
		}

		// Update target cash account balance
		if err := u.cashFlowRepo.UpdateAccount(ctx, tx, acc); err != nil {
			return fmt.Errorf("failed to update cash account balance: %w", err)
		}

		// Save capital transaction record
		if err := u.capitalRepo.CreateTx(ctx, tx, capitalTx); err != nil {
			return fmt.Errorf("failed to create capital transaction: %w", err)
		}

		// Insert into cash_flows ledger
		cf := &cashFlowDomain.CashFlow{
			ID:              uuid.New(),
			TransactionDate: capitalTx.TransactionDate,
			ReferenceType:   refType,
			ReferenceID:     &capitalTx.ID,
			Type:            cashFlowType,
			Amount:          amount,
			PaymentMethod:   paymentMethod,
			Description:     desc,
			CashierID:       creatorID,
		}

		if err := u.cashFlowRepo.CreateTx(ctx, tx, cf); err != nil {
			return fmt.Errorf("failed to create cash flow entry: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return capitalTx, nil
}

func (u *capitalUsecase) FindAll(ctx context.Context, filter capitalDomain.CapitalFilter) ([]capitalDomain.CapitalTransaction, int64, error) {
	return u.capitalRepo.FindAll(ctx, filter)
}

func (u *capitalUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		capitalTx, err := u.capitalRepo.FindByIDTx(ctx, tx, id)
		if err != nil {
			return fmt.Errorf("capital transaction not found: %w", err)
		}

		// Lock target cash account
		acc, err := u.cashFlowRepo.FindAccountByNameWithLock(ctx, tx, capitalTx.PaymentMethod)
		if err != nil {
			return fmt.Errorf("failed to lock cash account: %w", err)
		}

		// Reverse balance update
		if capitalTx.Type == capitalDomain.CapitalInjection {
			if acc.Balance < capitalTx.Amount {
				return errors.New("cannot cancel injection: reversing would result in negative cash account balance")
			}
			acc.Balance -= capitalTx.Amount
		} else {
			acc.Balance += capitalTx.Amount
		}

		// Update cash account balance
		if err := u.cashFlowRepo.UpdateAccount(ctx, tx, acc); err != nil {
			return fmt.Errorf("failed to update cash account balance: %w", err)
		}

		// Soft-delete capital transaction
		if err := u.capitalRepo.DeleteTx(ctx, tx, id); err != nil {
			return fmt.Errorf("failed to delete capital transaction: %w", err)
		}

		// Soft-delete cash flows entry referencing this transaction
		// We delete cash_flows where reference_id matches capital transaction ID
		if err := tx.WithContext(ctx).
			Where("reference_id = ?", id).
			Delete(&cashFlowDomain.CashFlow{}).Error; err != nil {
			return fmt.Errorf("failed to delete linked cash flow entry: %w", err)
		}

		return nil
	})
}
