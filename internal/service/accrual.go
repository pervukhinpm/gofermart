package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pervukhinpm/gophermart/internal/model"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

var ErrUnknownError = errors.New("unknown error")
var ErrOrderNotRegistered = errors.New("order not registered")

type ErrTooManyRequests struct {
	retryAfter int
}

func NewErrTooManyRequests(retryAfter int) *ErrTooManyRequests {
	return &ErrTooManyRequests{retryAfter: retryAfter}
}

func (e *ErrTooManyRequests) Error() string {
	return fmt.Sprintf("too many requests, retry after: %d", e.retryAfter)
}

type AccrualService interface {
	StartWorkers()
	CreateAccrual(orderNumber string, userID string)
}

// StartWorkers запускает воркеры для обработки задач в очереди
func (g *GophermartService) StartWorkers() {
	for i := 0; i < g.workerCount; i++ {
		go g.worker(i)
	}
}

func (g *GophermartService) CreateAccrual(orderNumber string, userID string) {
	order := &model.Order{OrderNumber: orderNumber, UserID: userID, Status: model.OrderStatusNew}
	// Блокировка до тех пор, пока не появится место в очереди
	for {
		select {
		case g.taskQueue <- order:
			log.Printf("Order %s added to queue for accrual processing", orderNumber)
			return
		default:
			log.Printf("Task queue is full, waiting to process order %s", orderNumber)
			time.Sleep(1 * time.Second)
		}
	}
}

// worker обрабатывает заказы из очереди
func (g *GophermartService) worker(id int) {
	for order := range g.taskQueue {
		log.Printf("Worker %d processing order %s", id, order.OrderNumber)
		g.processAccrual(order)
	}
}

func (g *GophermartService) processAccrual(order *model.Order) {
	accrual, err := g.getAccrual(order)

	retryDelay := time.Second * 1

	if err != nil {
		log.Printf("Error getting accrual for order %s: %s", order.OrderNumber, err)

		// Обработка кода 429 (Too Many Requests)
		if tooManyRequestErr := new(ErrTooManyRequests); errors.As(err, &tooManyRequestErr) {
			retryDelay = time.Second * time.Duration(tooManyRequestErr.retryAfter)
		}
	}

	time.Sleep(retryDelay)

	switch accrual.Status {
	case model.OrderStatusRegistered, model.OrderStatusProcessing:
		g.processAccrual(order)
	case model.OrderStatusProcessed, model.OrderStatusInvalid:
		updatedOrder := &model.Order{
			OrderNumber: order.OrderNumber,
			UserID:      order.UserID,
			Status:      accrual.Status,
			ProcessedAt: order.ProcessedAt,
			Accrual:     int(accrual.Accrual * 100),
		}
		err = g.repo.UpdateOrder(context.Background(), updatedOrder)
		if err != nil {
			log.Printf("Error updating order status for order %s: %s", order.OrderNumber, err)
		}
	default:
		g.processAccrual(order)
	}
}

func (g *GophermartService) getAccrual(order *model.Order) (*model.Accrual, error) {
	accrual := &model.Accrual{}

	url := fmt.Sprintf("%s/api/orders/%s", g.appConfig.AccrualSystemAddress, order.OrderNumber)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
		return accrual, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	log.Printf("creating request: %v\n response: %v", req, resp)

	if err != nil {
		return accrual, err
	}
	bytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return accrual, err
	}
	err = json.Unmarshal(bytes, &accrual)
	if err != nil {
		return accrual, err
	}

	switch resp.StatusCode {
	case 200:
		return accrual, nil
	case 204:
		return accrual, ErrOrderNotRegistered
	case 429:
		retryAfterStr := resp.Header.Get("Retry-After")
		retryAfter := 60
		if retryAfterStr != "" {
			retryAfter, err = strconv.Atoi(retryAfterStr)
			if err != nil {
				return nil, NewErrTooManyRequests(retryAfter)
			}
		}
		return nil, NewErrTooManyRequests(retryAfter)
	}

	return nil, ErrUnknownError
}
