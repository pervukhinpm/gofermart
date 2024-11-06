package main

import (
	"context"
	"github.com/pervukhinpm/gophermart/internal/api"
	"github.com/pervukhinpm/gophermart/internal/config"
	"github.com/pervukhinpm/gophermart/internal/middleware"
	"github.com/pervukhinpm/gophermart/internal/repository"
	"github.com/pervukhinpm/gophermart/internal/service"
	"log"
)

func main() {
	middleware.Initialize()
	appConfig := config.NewConfig()

	appRepository, err := repository.NewDatabaseRepository(
		context.Background(),
		*appConfig,
	)

	if err != nil {
		middleware.Log.Error("Failed to initialize repository: %v", err)
		return
	}

	defer func(appRepository repository.Repository) {
		err := appRepository.Close()
		if err != nil {
			middleware.Log.Error("Failed to close repository: %v", err)
		}
	}(appRepository)

	orderService := service.NewOrderService(appRepository)
	accrualService := service.NewAccrualService(appRepository, *appConfig)
	userService := service.NewUserService(appRepository)
	handler := api.NewGophermartHandler(orderService, accrualService, userService)

	router := api.Router(handler)

	server := api.NewServer(appConfig.RunAddress, router)

	log.Fatal(server.Start())
}
