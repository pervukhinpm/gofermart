package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/pervukhinpm/gophermart/internal/config"
	"github.com/pervukhinpm/gophermart/internal/middleware"
)

func Router(g *GophermartHandler, config config.Config) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	// Публичные маршруты (без аутентификации)
	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", g.RegisterUser)
		r.Post("/api/user/login", g.LoginUser)
	})

	// Маршруты, требующие аутентификации
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(config))

		r.Post("/api/user/orders", g.CreateOrder)
		r.Get("/api/user/orders", g.GetOrders)

		r.Get("/api/user/balance", g.GetUserBalance)
		r.Post("/api/user/balance/withdraw", g.CreateWithdraw)

		r.Get("/api/user/withdrawals", g.GetWithdrawals)
	})

	return r
}
