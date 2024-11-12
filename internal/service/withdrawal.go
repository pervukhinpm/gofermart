package service

import (
	"context"
	"errors"
	"github.com/pervukhinpm/gophermart/internal/model"
	"time"
)

var ErrLowBalance = errors.New("low balance")

type WithdrawalService interface {
	GetWithdrawals(ctx context.Context, userID string) (*[]model.Withdrawal, error)
	CreateWithdraw(ctx context.Context, withdrawal *model.Withdrawal) error
}

func (g *GophermartService) GetWithdrawals(ctx context.Context, userID string) (*[]model.Withdrawal, error) {
	withdrawals, err := g.repo.GetWithdrawals(ctx, userID)
	return withdrawals, err
}

func (g *GophermartService) CreateWithdraw(ctx context.Context, withdrawal *model.Withdrawal) error {
	user, err := g.repo.GetUser(ctx, withdrawal.UserID)
	if err != nil {
		return err
	}

	withdrawalAmount := int(withdrawal.Amount * 100)
	if withdrawalAmount > user.Balance {
		return ErrLowBalance
	}
	user.Balance -= withdrawalAmount
	user.Withdrawn += withdrawalAmount

	withdrawal.ProcessedAt = time.Now()
	err = g.repo.CreateWithdrawal(ctx, user, withdrawal)
	return nil
}
