package service

import (
	"context"
	"errors"
	"github.com/pervukhinpm/gophermart/internal/luhn"
	"github.com/pervukhinpm/gophermart/internal/model"
	"time"
)

var ErrOrderNumberInvalid = errors.New("order number is invalid")

type OrderService interface {
	CreateOrder(ctx context.Context, orderNumber string, userID string) error
	GetOrdersByUserID(ctx context.Context, userID string) (*[]model.Order, error)
	ValidateOrderNumber(orderNumber string) bool
}

func (g *GophermartService) CreateOrder(ctx context.Context, orderNumber string, userID string) error {
	order := &model.Order{
		OrderNumber: orderNumber,
		UserID:      userID,
		Status:      model.OrderStatusNew,
		ProcessedAt: time.Now(),
	}
	return g.repo.AddOrder(ctx, order)
}

func (g *GophermartService) GetOrdersByUserID(ctx context.Context, userID string) (*[]model.Order, error) {
	return g.repo.GetOrders(ctx, userID)
}

func (g *GophermartService) ValidateOrderNumber(orderNumber string) bool {
	return luhn.Validate(orderNumber)
}
