package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/pervukhinpm/gophermart/internal/jwt"
	"github.com/pervukhinpm/gophermart/internal/model"
	"github.com/pervukhinpm/gophermart/internal/repository"
)

var ErrInvalidLoginAndPassword = errors.New("invalid login and password")

type UserService interface {
	RegisterUser(ctx context.Context, registerUser *model.RegisterUser) (string, error)
	LoginUser(ctx context.Context, loginUser *model.LoginUser) (string, error)
	GetUserBalance(ctx context.Context, userID string) (*model.UserBalance, error)
}

func (g *GophermartService) RegisterUser(ctx context.Context, registerUser *model.RegisterUser) (string, error) {
	userID := uuid.NewString()

	user := &model.User{
		ID:       userID,
		Login:    registerUser.Login,
		Password: registerUser.Password,
	}

	err := g.repo.CreateUser(ctx, user)

	if err != nil {
		return "", err
	}

	return jwt.GenerateJWT(userID, g.appConfig)
}

func (g *GophermartService) LoginUser(ctx context.Context, loginUser *model.LoginUser) (string, error) {
	dbUser, err := g.repo.GetUserByLogin(ctx, loginUser.Login)

	if err != nil {
		if errors.Is(err, repository.ErrNoAuthUser) {
			return "", ErrInvalidLoginAndPassword
		}
		return "", err
	}

	if loginUser.Login == dbUser.Login && loginUser.Password == dbUser.Password {
		return jwt.GenerateJWT(dbUser.ID, g.appConfig)
	}

	return "", ErrInvalidLoginAndPassword
}

func (g *GophermartService) GetUserBalance(ctx context.Context, userID string) (*model.UserBalance, error) {
	user, err := g.repo.GetUserByID(ctx, userID)
	userBalance := &model.UserBalance{}

	if err != nil {
		return userBalance, err
	}
	userBalance.Withdrawn = float64(user.Withdrawn / 100)
	userBalance.Current = float64(user.Balance / 100)
	return userBalance, nil
}
