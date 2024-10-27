package service

import (
	"context"
	"github.com/pervukhinpm/gophermart/internal/models"
	"github.com/pervukhinpm/gophermart/internal/repository"
)

type OrderService struct {
	repo repository.DatabaseRepository
}

func NewOrderService(repo repository.DatabaseRepository) *OrderService {
	return &OrderService{
		repo: repo,
	}
}

func (o *OrderService) CreateOrder(ctx context.Context, orderNumber string, userID string) error {

	return nil
}

func (o *OrderService) GetOrdersByUserID(ctx context.Context, userID string) (*[]models.Order, error) {
	return nil, nil
}

func (o *OrderService) ValidateOrderNumber(orderNumber string) bool {
	return true
}
