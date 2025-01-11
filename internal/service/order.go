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

type OrderService interface {
	CreateOrder(ctx context.Context, orderNumber string, userID string) error
	GetOrdersByUserID(ctx context.Context, userID string) (*[]model.OrderResponse, error)
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

func (g *GophermartService) GetOrdersByUserID(ctx context.Context, userID string) (*[]model.OrderResponse, error) {
	orders, err := g.repo.GetOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(*orders) == 0 {
		return nil, repository.ErrNoOrders
	}
	ordersResponse := make([]model.OrderResponse, 0)
	for _, order := range *orders {
		var orderResponse model.OrderResponse
		orderResponse.Status = order.Status
		orderResponse.Accrual = order.Accrual
		orderResponse.Number = order.OrderNumber
		orderResponse.UploadedAt = order.ProcessedAt
		ordersResponse = append(ordersResponse, orderResponse)
	}
	return &ordersResponse, nil
}

func (g *GophermartService) ValidateOrderNumber(orderNumber string) bool {
	return luhn.Validate(orderNumber)
}
