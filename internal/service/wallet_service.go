package service

import (
	"context"
	"errors"

	"wallet-service/internal/repository"

	"github.com/google/uuid"
)

type WalletService struct {
	repo repository.WalletRepositoryInterface
}

func NewWalletService(repo repository.WalletRepositoryInterface) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) Deposit(ctx context.Context, walletID uuid.UUID, amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	return s.repo.UpdateBalance(ctx, walletID, amount)
}

func (s *WalletService) Withdraw(ctx context.Context, walletID uuid.UUID, amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	return s.repo.UpdateBalance(ctx, walletID, -amount)
}

func (s *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	return s.repo.GetBalance(ctx, walletID)
}