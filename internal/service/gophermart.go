package service

import (
	"github.com/pervukhinpm/gophermart/internal/config"
	"github.com/pervukhinpm/gophermart/internal/model"
	"github.com/pervukhinpm/gophermart/internal/repository"
)

type Service interface {
	UserService
	OrderService
	AccrualService
	WithdrawalService
}

type GophermartService struct {
	repo        *repository.DatabaseRepository
	appConfig   config.Config
	taskQueue   chan *model.Order
	workerCount int
}

func NewGophermartService(repo *repository.DatabaseRepository, appConfig config.Config) *GophermartService {
	return &GophermartService{
		repo:        repo,
		appConfig:   appConfig,
		taskQueue:   make(chan *model.Order, 100),
		workerCount: 5,
	}
}
