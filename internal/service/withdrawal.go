package service

import (
	"context"
	"errors"
	"github.com/pervukhinpm/gophermart/internal/model"
	"github.com/pervukhinpm/gophermart/internal/repository"
	"time"
)

var ErrLowBalance = errors.New("low balance")

type WithdrawalService interface {
	GetWithdrawals(ctx context.Context, userID string) (*[]model.Withdrawal, error)
	CreateWithdraw(ctx context.Context, withdrawal *model.Withdrawal) error
}

func (g *GophermartService) GetWithdrawals(ctx context.Context, userID string) (*[]model.Withdrawal, error) {
	return g.repo.GetWithdrawals(ctx, userID)
}

func (g *GophermartService) CreateWithdraw(ctx context.Context, withdrawal *model.Withdrawal) error {
	user, err := g.repo.GetUserByID(ctx, withdrawal.UserID)
	if err != nil {
		return err
	}

	withdrawalAmount := withdrawal.Amount * 100
	if withdrawalAmount > user.Balance {
		return ErrLowBalance
	}
	user.Balance -= withdrawalAmount
	user.Withdrawn += withdrawalAmount

	withdrawal.ProcessedAt = time.Now()

	dbWithdrawal := repository.Withdrawal{
		OrderID:     withdrawal.OrderID,
		Amount:      int(withdrawalAmount),
		ProcessedAt: withdrawal.ProcessedAt,
	}

	return g.repo.CreateWithdrawal(ctx, user, &dbWithdrawal)
}
