package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/pervukhinpm/gophermart/internal/middleware"
)

func Router(g *GophermartHandler) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	// Публичные маршруты (без аутентификации)
	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", g.RegisterUser)
		r.Post("/api/user/login", g.LoginUser)
	})

	// Маршруты, требующие аутентификации
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)

		r.Post("/api/user/orders", g.CreateOrder)
		r.Get("/api/user/orders", g.GetOrders)

		r.Get("/api/user/balance", g.GetUserBalance)
		r.Post("/api/user/balance/withdraw", g.CreateWithdraw)

		r.Get("/api/user/withdrawals", g.GetWithdrawals)
	})

	return r
}
