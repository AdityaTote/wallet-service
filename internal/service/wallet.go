package service

import (
	"context"
	"fmt"

	"github.com/AdityaTote/wallet-service/internal/models"
	"github.com/AdityaTote/wallet-service/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type WalletService interface {
	TopUp(*models.WalletServiceParams) (*models.WalletResponse, error)
	Spend(*models.WalletServiceParams) (*models.WalletResponse, error)
	Balance(userId uuid.UUID, walletId uuid.UUID) (int64, error)
	// Bonus()
}

type walletService struct {
	ctx context.Context
	log zerolog.Logger
	repo repository.Repository
}

func (w *walletService) TopUp(input *models.WalletServiceParams) (*models.WalletResponse, error) {
	query := w.repo.Queries()
	
	// check the tnx id
	_, err := query.GetTransactionById(w.ctx, input.TxnId)
	if err == nil {
		w.log.Debug().Msg("transaction with id already exist")

		// get balance for wallet
		balanceInterface, err := query.GetBalance(w.ctx, input.WalletId)
		if err != nil {
			w.log.Error().Err(err).Msg("failed to get balance for wallet")
			return nil, models.ErrBalanceRetrievalFailed
		}
		fmt.Println("balanceInterface: ", balanceInterface)
		balance, ok := balanceInterface.(int64)
		if !ok {
			w.log.Error().Msg("invalid balance type")
			return nil, models.ErrInvalidBalance
		}

		fmt.Println("balance: ", balance)

		return &models.WalletResponse{
			Message: "transaction with id already exist",
			Balance: balance,
		}, nil
	}

	var balance int64

	err = w.repo.WithTransaction(w.ctx, func(q *repository.Queries) error {
		var ok bool
		// lock wallet row
		walletId, err := q.LockWallet(w.ctx, input.UserId)
		if err != nil {
			return err
		}

		// create tnx
		tnx, err := q.CreateTxn(w.ctx, repository.CreateTxnParams{
			ID: input.TxnId,
			Type: repository.TransactionTypeTOPUP,
		})
		if err != nil {
			return err
		}

		// add ledger entry for topup for user account
		_, err = q.CreateLedger(w.ctx, repository.CreateLedgerParams{
			Amount: int32(input.Amount),
			WalletID: walletId,
			TransactionID: tnx.ID,
		})
		if err != nil {
			return err
		}

		systemWalletId, err := q.GetSystemWallet(w.ctx)
		if err != nil {
			return err
		}

		// add legder entry for spend for system account
		_, err = q.CreateLedger(w.ctx, repository.CreateLedgerParams{
			Amount: -int32(input.Amount),
			TransactionID: tnx.ID,
			WalletID: systemWalletId,
		})
		if err != nil {
			return err
		}

		balanceInterface, err := q.GetBalance(w.ctx, walletId)
		if err != nil {
			return err
		}
		balance, ok = balanceInterface.(int64)
		if !ok {
			return models.ErrInvalidBalance
		}

		return nil
	})

	if err != nil {
		w.log.Error().Err(err).Msg("transaction failed")
		return nil, models.NewAppError(err, "transaction failed", 500)
	}

	return &models.WalletResponse{
		Message: "topup successful",
		Balance: balance,
	}, nil
}

func (w *walletService) Spend(input *models.WalletServiceParams) (*models.WalletResponse, error) {
	query := w.repo.Queries()

	// check the tnx id
	_, err := query.GetTransactionById(w.ctx, input.TxnId)
	if err == nil {
		w.log.Debug().Msg("transaction with id already exist")

		// get balance for wallet
		balanceInterface, err := query.GetBalance(w.ctx, input.WalletId)
		if err != nil {
			w.log.Error().Err(err).Msg("failed to get balance for wallet")
			return nil, models.ErrBalanceRetrievalFailed
		}
		balance, ok := balanceInterface.(int64)
		if !ok {
			w.log.Error().Msg("invalid balance type")
			return nil, models.ErrInvalidBalance
		}

		return &models.WalletResponse{
			Message: "transaction with id already exist",
			Balance: balance,
		}, nil
	}

	var balance int64

	err = w.repo.WithTransaction(w.ctx, func(q *repository.Queries) error {
		var ok bool
		// lock wallet row
		walletId, err := q.LockWallet(w.ctx, input.UserId)
		if err != nil {
			return err
		}

		// check balance for wallet
		currBalanceInterface, err := q.GetBalance(w.ctx, walletId)
		if err != nil {
			return err
		}
		currBalance, ok := currBalanceInterface.(int64)
		if !ok {
			return models.ErrInvalidBalance
		}

		if currBalance < input.Amount {
			return models.ErrInsufficientBalance
		}

		// create tnx
		tnx, err := q.CreateTxn(w.ctx, repository.CreateTxnParams{
			ID: input.TxnId,
			Type: repository.TransactionTypeSPEND,
		})
		if err != nil {
			return err
		}

		// add ledger entry for spend for user account
		_, err = q.CreateLedger(w.ctx, repository.CreateLedgerParams{
			Amount: -int32(input.Amount),
			TransactionID: tnx.ID,
			WalletID: walletId,
		})
		if err != nil {
			return err
		}

		// find system account
		systemWalletId, err := q.GetSystemWallet(w.ctx)
		if err != nil {
			return err
		}

		// add ledger entry for topup for system account
		_, err = q.CreateLedger(w.ctx, repository.CreateLedgerParams{
			Amount: int32(input.Amount),
			TransactionID: tnx.ID,
			WalletID: systemWalletId,
		})
		if err != nil {
			return err
		}

		balanceInterface, err := q.GetBalance(w.ctx, walletId)
		if err != nil {
			return err
		}
		balance, ok = balanceInterface.(int64)
		if !ok {
			return models.ErrInvalidBalance
		}

		return nil
	})

	if err != nil {
		w.log.Error().Err(err).Msg("transaction failed")
		return nil, models.NewAppError(err, "transaction failed", 500)
	}

	return &models.WalletResponse{
		Message: "spend successful",
		Balance: balance,
	}, nil
}

func (w *walletService) Balance(userId uuid.UUID, walletId uuid.UUID) (int64, error) {
	// get balance for user account
	query := w.repo.Queries()

	balanceInterface, err := query.GetBalance(w.ctx, walletId)
	if err != nil {
		w.log.Error().Err(err).Msg("failed to get balance for wallet")
		return 0, models.ErrBalanceRetrievalFailed
	}

	balance, ok := balanceInterface.(int64)
	if !ok {
		w.log.Error().Msg("invalid balance type")
		return 0, models.ErrInvalidBalance
	}

	return balance, nil
}