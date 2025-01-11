package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pervukhinpm/gophermart/internal/middleware"
	"github.com/pervukhinpm/gophermart/internal/model"
	"github.com/pervukhinpm/gophermart/internal/repository"
	"github.com/pervukhinpm/gophermart/internal/service"
	"io"
	"net/http"
	"time"
)

type GophermartHandler struct {
	service service.Service
}

func NewGophermartHandler(service service.Service) *GophermartHandler {
	return &GophermartHandler{service: service}
}

func (g *GophermartHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var user model.RegisterUser
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(bodyBytes, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	tokenString, err := g.service.RegisterUser(ctx, &user)
	if err != nil {
		if errors.Is(err, repository.ErrUserDuplicated) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	w.WriteHeader(http.StatusOK)
}

func (g *GophermartHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var user model.LoginUser
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(bodyBytes, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	tokenString, err := g.service.LoginUser(ctx, &user)
	if err != nil {
		if errors.Is(err, service.ErrInvalidLoginAndPassword) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	w.WriteHeader(http.StatusOK)
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

	userID := middleware.GetUserID(r.Context())

	orderNumber := string(body)

	isValidOrderNumber := g.service.ValidateOrderNumber(orderNumber)

	if !isValidOrderNumber {
		middleware.Log.Error("Invalid order number!", orderNumber)
		http.Error(w, "Invalid order number!", http.StatusUnprocessableEntity)
		return
	}

	err = g.service.CreateOrder(r.Context(), orderNumber, userID)

	if err != nil {
		if errors.Is(err, repository.ErrOrderAlreadyCreatedByUser) {
			middleware.Log.Error("Order already exists by user %s", userID)
			http.Error(w, "Order already exists", http.StatusOK)
			return
		}
		if errors.Is(err, repository.ErrOrderAlreadyExist) {
			middleware.Log.Error("Order already exists")
			http.Error(w, "Order already exists", http.StatusConflict)
			return
		}
		if errors.Is(err, service.ErrOrderNumberInvalid) {
			middleware.Log.Error("Invalid order number!", err)
			http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	g.service.CreateAccrual(orderNumber, userID)

	w.WriteHeader(http.StatusAccepted)
}

func (g *GophermartHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	ctx := r.Context()

	orders, err := g.service.GetOrdersByUserID(ctx, userID)

	if err != nil {
		if errors.Is(err, repository.ErrNoOrders) {
			middleware.Log.Error("Error no orders by user id %s", userID)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		middleware.Log.Error("Error getting orders by user id %s", userID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(orders)
	if err != nil {
		middleware.Log.Error("Error marshalling response")
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
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())

	balance, err := g.service.GetUserBalance(r.Context(), userID)

	if err != nil {
		middleware.Log.Error("Error getting user balance %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response, err := json.Marshal(balance)
	if err != nil {
		middleware.Log.Error("Error marshaling response %s", err)
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

func (g *GophermartHandler) CreateWithdraw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.Log.Error("Error reading body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var withdrawal model.Withdrawal
	err = json.Unmarshal(bodyBytes, &withdrawal)
	if err != nil {
		middleware.Log.Error("Error unmarshalling body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	valid := g.service.ValidateOrderNumber(withdrawal.OrderID)
	if !valid {
		middleware.Log.Error("Invalid order number")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	userID := middleware.GetUserID(r.Context())

	withdrawal.UserID = userID

	err = g.service.CreateWithdraw(r.Context(), &withdrawal)
	if err != nil {
		if errors.Is(err, service.ErrLowBalance) {
			middleware.Log.Warn("Error creating withdrawal, low balance")
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		middleware.Log.Error("Error creating withdrawal")
		w.WriteHeader(http.StatusInternalServerError)
		return

	}
	w.WriteHeader(http.StatusOK)
}

func (g *GophermartHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())

	withdrawals, err := g.service.GetWithdrawals(r.Context(), userID)

	if err != nil {
		if errors.Is(err, repository.ErrNoWithdrawals) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response, err := json.Marshal(withdrawals)
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
