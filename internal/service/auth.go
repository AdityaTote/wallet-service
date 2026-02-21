package service

import (
	"context"
	"fmt"

	"github.com/AdityaTote/wallet-service/internal/config"
	"github.com/AdityaTote/wallet-service/internal/lib/utils"
	"github.com/AdityaTote/wallet-service/internal/models"
	"github.com/AdityaTote/wallet-service/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type AuthService interface {
	Signup(models.UserParams) (*models.UserResponse ,error)
	Signin(models.UserParams) (*models.UserResponse ,error)
}

type authService struct {
	ctx context.Context
	cfg *config.Config
	log zerolog.Logger
	repo *repository.Queries
}

func (a *authService) Signup(input models.UserParams) (*models.UserResponse ,error) {
	_, err := a.repo.GetUserByUsername(a.ctx, input.Username)
	if err == nil {
		a.log.Debug().Msg("user with username already exist")
		return nil, fmt.Errorf("authentication failed")
	}
	
	hash_pass, err := utils.HashPassword(input.Password)
	if err != nil {
		a.log.Error().Err(err).Msg("failed to hash password")
		return nil, fmt.Errorf("authentication failed")
	}

	user, err := a.repo.CreateUser(a.ctx, repository.CreateUserParams{
		Username: input.Username,
		Password: hash_pass,
	})

	if err != nil {
		a.log.Error().Err(err).Msg("failed to create user")
		return nil, fmt.Errorf("authentication failed")
	}

	assets, err := a.repo.GetAssetByCode(a.ctx, config.AssetCodeUC)
	if err != nil {
		a.log.Error().Err(err).Msg("failed to get asset")
		return nil, fmt.Errorf("authentication failed")
	}

	wallet, err := a.repo.CreateWallet(a.ctx, repository.CreateWalletParams{
		OwnerType: repository.WalletOwnerTypeUSER,
		OwnerID: user.ID,
		AssetID: assets.ID,
	})
	if err != nil {
		a.log.Error().Err(err).Msg("failed to create wallet")
		return nil, fmt.Errorf("authentication failed")
	}

	// initialize bonus transaction for user
	tnx, err := a.repo.CreateTxn(a.ctx, repository.CreateTxnParams{
		ID: uuid.New(),
		Type: repository.TransactionTypeBONUS,
	})
	if err != nil {
		a.log.Error().Err(err).Msg("failed to create transaction")
		return nil, fmt.Errorf("authentication failed")
	}

	// create ledger entry for bonus transaction
	_, err = a.repo.CreateLedger(a.ctx, repository.CreateLedgerParams{
		TransactionID: tnx.ID,
		WalletID: wallet.ID,
		Amount: config.InitialBonusAmount,
	})
	if err != nil {
		a.log.Error().Err(err).Msg("failed to create ledger entry")
		return nil, fmt.Errorf("authentication failed")
	}

	// get balance for wallet
	balanceInterface, err := a.repo.GetBalance(a.ctx, wallet.ID)
	if err != nil {
		a.log.Error().Err(err).Msg("failed to get wallet balance")
		return nil, fmt.Errorf("authentication failed")
	}

	balance, ok := balanceInterface.(int64)
	if !ok {
		a.log.Error().Msg("failed to assert balance type")
		return nil, fmt.Errorf("authentication failed")
	}
	
	// generate access token
	access_token, err := utils.GenerateAccessToken([]byte(a.cfg.JWTSecret), user.ID, wallet.ID)
	if err != nil {
		a.log.Error().Err(err).Msg("failed to generate access token")
		return nil, fmt.Errorf("authentication failed")
	}

	return &models.UserResponse{
		Id:       user.ID,
		Username: user.Username,
		WalletId: &wallet.ID,
		Balance: &balance,
		AccessToken: access_token,
	}, nil
}

func (a *authService) Signin(input models.UserParams) (*models.UserResponse ,error) {
	user, err := a.repo.GetUserByUsername(a.ctx, input.Username)
	if err != nil {
		a.log.Debug().Msg("user with username does not exist")
		return nil, fmt.Errorf("authentication failed")
	}

	if !utils.VerifyPassword(user.Password, input.Password) {
		a.log.Debug().Msg("password does not match")
		return nil, fmt.Errorf("authentication failed")
	}

	wallet, err := a.repo.GetWalletByOwner(a.ctx, user.ID)
	if err != nil {
		a.log.Error().Err(err).Msg("failed to get wallet for user")
		return nil, fmt.Errorf("authentication failed")
	}

	access_token, err := utils.GenerateAccessToken([]byte(a.cfg.JWTSecret), user.ID, wallet.ID)
	if err != nil {
		a.log.Error().Err(err).Msg("failed to generate access token")
		return nil, fmt.Errorf("authentication failed")
	}

	return &models.UserResponse{
		Id:       user.ID,
		Username: user.Username,
		AccessToken: access_token,
	}, nil
}