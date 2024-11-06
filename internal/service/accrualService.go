package service

import (
	"github.com/pervukhinpm/gophermart/internal/config"
	"github.com/pervukhinpm/gophermart/internal/repository"
)

type AccrualService struct {
	repo      *repository.DatabaseRepository
	appConfig config.Config
}

func NewAccrualService(repo *repository.DatabaseRepository, appConfig config.Config) *AccrualService {
	return &AccrualService{repo: repo, appConfig: appConfig}
}

func (*AccrualService) CreateAccrual(orderNumber string, userID string) {

}
