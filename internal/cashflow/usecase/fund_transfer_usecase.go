package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type fundTransferUsecase struct {
	fundTransferRepo domain.FundTransferRepository
	cashFlowRepo     domain.CashFlowRepository
	db               *gorm.DB
}

func NewFundTransferUsecase(
	fundTransferRepo domain.FundTransferRepository,
	cashFlowRepo domain.CashFlowRepository,
	db *gorm.DB,
) domain.FundTransferUsecase {
	return &fundTransferUsecase{
		fundTransferRepo: fundTransferRepo,
		cashFlowRepo:     cashFlowRepo,
		db:               db,
	}
}

func (u *fundTransferUsecase) Transfer(
	ctx context.Context,
	cashierID uuid.UUID,
	fromAccountName string,
	toAccountName string,
	amount float64,
	notes string,
	transferDate *time.Time,
) (*domain.FundTransfer, error) {
	if fromAccountName == toAccountName {
		return nil, errors.New("origin and destination accounts cannot be the same")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}

	var transfer *domain.FundTransfer

	err := u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var accFrom, accTo *domain.CashAccount
		var err error

		// Lock accounts in alphabetical order of their names to prevent deadlocks
		if fromAccountName < toAccountName {
			accFrom, err = u.cashFlowRepo.FindAccountByNameWithLock(ctx, tx, fromAccountName)
			if err != nil {
				return fmt.Errorf("origin account '%s' not found: %w", fromAccountName, err)
			}
			accTo, err = u.cashFlowRepo.FindAccountByNameWithLock(ctx, tx, toAccountName)
			if err != nil {
				return fmt.Errorf("destination account '%s' not found: %w", toAccountName, err)
			}
		} else {
			accTo, err = u.cashFlowRepo.FindAccountByNameWithLock(ctx, tx, toAccountName)
			if err != nil {
				return fmt.Errorf("destination account '%s' not found: %w", toAccountName, err)
			}
			accFrom, err = u.cashFlowRepo.FindAccountByNameWithLock(ctx, tx, fromAccountName)
			if err != nil {
				return fmt.Errorf("origin account '%s' not found: %w", fromAccountName, err)
			}
		}

		if accFrom.Balance < amount {
			return fmt.Errorf("insufficient balance in account '%s' (current: %.2f, required: %.2f)", fromAccountName, accFrom.Balance, amount)
		}

		// Update balances
		accFrom.Balance -= amount
		accTo.Balance += amount

		if err := u.cashFlowRepo.UpdateAccount(ctx, tx, accFrom); err != nil {
			return err
		}
		if err := u.cashFlowRepo.UpdateAccount(ctx, tx, accTo); err != nil {
			return err
		}

		// Create FundTransfer
		var notesPtr *string
		if notes != "" {
			notesPtr = &notes
		}
		
		var trfDate time.Time
		if transferDate != nil && !transferDate.IsZero() {
			trfDate = *transferDate
		} else {
			trfDate = time.Now()
		}

		transfer = &domain.FundTransfer{
			ID:            uuid.New(),
			TransferDate:  trfDate,
			FromAccountID: accFrom.ID,
			ToAccountID:   accTo.ID,
			Amount:        amount,
			Notes:         notesPtr,
			CashierID:     cashierID,
		}

		if err := u.fundTransferRepo.CreateTx(ctx, tx, transfer); err != nil {
			return err
		}

		// Create CashFlow ledger entries
		desc := fmt.Sprintf("Pemindahan Dana dari %s ke %s", fromAccountName, toAccountName)
		if notes != "" {
			desc = fmt.Sprintf("%s (%s)", desc, notes)
		}

		cfFrom := &domain.CashFlow{
			ID:              uuid.New(),
			TransactionDate: transfer.TransferDate,
			ReferenceType:   domain.RefFundTransfer,
			ReferenceID:     &transfer.ID,
			Type:            domain.TypeCredit,
			Amount:          amount,
			PaymentMethod:   fromAccountName,
			Description:     &desc,
			CashierID:       cashierID,
		}

		cfTo := &domain.CashFlow{
			ID:              uuid.New(),
			TransactionDate: transfer.TransferDate,
			ReferenceType:   domain.RefFundTransfer,
			ReferenceID:     &transfer.ID,
			Type:            domain.TypeDebit,
			Amount:          amount,
			PaymentMethod:   toAccountName,
			Description:     &desc,
			CashierID:       cashierID,
		}

		if err := u.cashFlowRepo.CreateTx(ctx, tx, cfFrom); err != nil {
			return err
		}
		if err := u.cashFlowRepo.CreateTx(ctx, tx, cfTo); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return transfer, nil
}

func (u *fundTransferUsecase) FindAll(ctx context.Context, filter domain.FundTransferFilter) ([]domain.FundTransfer, int64, error) {
	return u.fundTransferRepo.FindAll(ctx, filter)
}

func (u *fundTransferUsecase) Cancel(ctx context.Context, cashierID uuid.UUID, id uuid.UUID) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transfer, err := u.fundTransferRepo.FindByID(ctx, id)
		if err != nil {
			return err
		}

		fromAccountName := transfer.FromAccount.Name
		toAccountName := transfer.ToAccount.Name

		var accFrom, accTo *domain.CashAccount

		// Lock accounts in alphabetical order
		if fromAccountName < toAccountName {
			accFrom, err = u.cashFlowRepo.FindAccountByNameWithLock(ctx, tx, fromAccountName)
			if err != nil {
				return err
			}
			accTo, err = u.cashFlowRepo.FindAccountByNameWithLock(ctx, tx, toAccountName)
			if err != nil {
				return err
			}
		} else {
			accTo, err = u.cashFlowRepo.FindAccountByNameWithLock(ctx, tx, toAccountName)
			if err != nil {
				return err
			}
			accFrom, err = u.cashFlowRepo.FindAccountByNameWithLock(ctx, tx, fromAccountName)
			if err != nil {
				return err
			}
		}

		// Reverse balances
		accFrom.Balance += transfer.Amount
		accTo.Balance -= transfer.Amount

		if err := u.cashFlowRepo.UpdateAccount(ctx, tx, accFrom); err != nil {
			return err
		}
		if err := u.cashFlowRepo.UpdateAccount(ctx, tx, accTo); err != nil {
			return err
		}

		// Delete FundTransfer
		if err := u.fundTransferRepo.DeleteTx(ctx, tx, id); err != nil {
			return err
		}

		// Delete CashFlow entries
		if err := tx.Where("reference_type = ? AND reference_id = ?", domain.RefFundTransfer, id).Delete(&domain.CashFlow{}).Error; err != nil {
			return err
		}

		return nil
	})
}
