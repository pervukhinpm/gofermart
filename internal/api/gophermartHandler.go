package api

import (
	"encoding/json"
	"errors"
	"github.com/pervukhinpm/gophermart/internal/repository"
	"github.com/pervukhinpm/gophermart/internal/service"
	"io"
	"net/http"
)

type GophermartHandler struct {
	orderService   *service.OrderService
	accrualService *service.AccrualService
}

func (g *GophermartHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {

}

func (g *GophermartHandler) LoginUser(w http.ResponseWriter, r *http.Request) {

}

func (g *GophermartHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(w, "Empty body!", http.StatusBadRequest)
		return
	}

	userID := "string"

	orderNumber := string(body)

	isValidOrderNumber := g.orderService.ValidateOrderNumber(orderNumber)

	if !isValidOrderNumber {
		http.Error(w, "Invalid order number!", http.StatusUnprocessableEntity)
		return
	}

	err = g.orderService.CreateOrder(r.Context(), orderNumber, userID)

	if err != nil {
		if errors.Is(err, repository.ErrOrderAlreadyCreatedByUser) {
			http.Error(w, "Order already exists", http.StatusOK)
			return
		}
		if errors.Is(err, repository.ErrOrderAlreadyExist) {
			http.Error(w, "Order already exists", http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	g.accrualService.CreateAccrual(orderNumber, userID)

	w.WriteHeader(http.StatusAccepted)
}

func (g *GophermartHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := ""

	ctx := r.Context()

	orders, err := g.orderService.GetOrdersByUserID(ctx, userID)

	if err != nil {
		if errors.Is(err, repository.ErrNoOrders) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(orders)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		return
	}
}

func (g *GophermartHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {

}

func (g *GophermartHandler) CreateWithdraw(w http.ResponseWriter, r *http.Request) {

}

func (g *GophermartHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {

}
