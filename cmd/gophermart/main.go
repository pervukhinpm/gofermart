package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pervukhinpm/gophermart/internal/api"
	"github.com/pervukhinpm/gophermart/internal/config"
	"github.com/pervukhinpm/gophermart/internal/middleware"
	"github.com/pervukhinpm/gophermart/internal/repository"
	"github.com/pervukhinpm/gophermart/internal/service"
	"log"
)

func main() {
	middleware.Initialize()

	appConfig, err := config.Load(".env")
	if err != nil {
		log.Fatal("failed to initialize config")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500)
	defer cancel()

	pool, err := pgxpool.New(ctx, appConfig.DataBaseURI)
	if err != nil {
		log.Fatal("failed to initialize pool")
	}
	defer pool.Close()

	appRepository := repository.NewDatabaseRepository(pool)
	err = appRepository.Migrate()
	if err != nil {
		log.Fatal("failed to migrate")
	}

	gophermartService := service.NewGophermartService(appRepository, appConfig)
	gophermartService.StartWorkers()

	handler := api.NewGophermartHandler(gophermartService)

	router := api.Router(handler, appConfig)

	server := api.NewServer(appConfig.RunAddress, router)

	log.Fatal(server.Start())
}
