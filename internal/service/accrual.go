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
	const maxRetries = 10
	const retryBaseDelay = time.Second

	var attempt int
	for attempt = 0; attempt < maxRetries; attempt++ {
		log.Printf("Attempt %d: Processing order %s", attempt+1, order.OrderNumber)

		accrual, err := g.getAccrual(order)
		if err != nil {
			log.Printf("Error getting accrual for order %s: %s", order.OrderNumber, err)

			// Если ошибка - 429 (Too Many Requests), ждём указанное время
			if tooManyRequestErr := new(ErrTooManyRequests); errors.As(err, &tooManyRequestErr) {
				delay := time.Second * time.Duration(tooManyRequestErr.retryAfter)
				log.Printf("Too many requests for order %s, retrying after %s", order.OrderNumber, delay)
				time.Sleep(delay)
				continue
			}

			time.Sleep(retryBaseDelay)
			continue
		}

		switch accrual.Status {
		case model.OrderStatusRegistered, model.OrderStatusProcessing:
			log.Printf("Order %s is still processing. Retrying...", order.OrderNumber)
			time.Sleep(retryBaseDelay)
			continue
		case model.OrderStatusProcessed, model.OrderStatusInvalid:
			updatedOrder := &model.Order{
				OrderNumber: order.OrderNumber,
				UserID:      order.UserID,
				Status:      accrual.Status,
				ProcessedAt: order.ProcessedAt,
				Accrual:     accrual.Accrual * 100,
			}
			err := g.repo.UpdateOrder(context.Background(), updatedOrder)
			if err != nil {
				log.Printf("Error updating order status for order %s: %s", order.OrderNumber, err)
			} else {
				log.Printf("Order %s successfully processed", order.OrderNumber)
			}
			return
		default:
			log.Printf("Unknown status for order %s: %s. Retrying...", order.OrderNumber, accrual.Status)
			time.Sleep(retryBaseDelay)
		}
	}

	log.Printf("Max retries reached for order %s. Giving up.", order.OrderNumber)
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
	defer resp.Body.Close()
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
			parsedRetryAfter, err := strconv.Atoi(retryAfterStr)
			if err == nil {
				retryAfter = parsedRetryAfter
			}
		}
		return nil, NewErrTooManyRequests(retryAfter)
	}

	return nil, ErrUnknownError
}
