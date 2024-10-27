package service

import (
	"github.com/pervukhinpm/gophermart/internal/config"
	"github.com/pervukhinpm/gophermart/internal/repository"
)

type AccrualService struct {
	repo      repository.DatabaseRepository
	appConfig config.Config
}

func (*AccrualService) CreateAccrual(orderNumber string, userID string) {

}
