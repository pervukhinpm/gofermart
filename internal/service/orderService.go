package service

import (
	"context"
	"errors"
	"github.com/pervukhinpm/gophermart/internal/luhn"
	"github.com/pervukhinpm/gophermart/internal/model"
	"github.com/pervukhinpm/gophermart/internal/repository"
	"time"
)

var ErrOrderNumberInvalid = errors.New("order number is invalid")

type OrderService struct {
	repo *repository.DatabaseRepository
}

func NewOrderService(repo *repository.DatabaseRepository) *OrderService {
	return &OrderService{
		repo: repo,
	}
}

func (o *OrderService) CreateOrder(ctx context.Context, orderNumber string, userID string) error {
	order := &model.Order{
		OrderNumber: orderNumber,
		UserID:      userID,
		Status:      model.OrderStatusNew,
		ProcessedAt: time.Now(),
	}
	return o.repo.AddOrder(ctx, order)
}

func (o *OrderService) GetOrdersByUserID(ctx context.Context, userID string) (*[]model.Order, error) {
	return o.repo.GetOrders(ctx, userID)
}

func (o *OrderService) ValidateOrderNumber(orderNumber string) bool {
	return luhn.Validate(orderNumber)
}
