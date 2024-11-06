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

type UserService struct {
	repo *repository.DatabaseRepository
}

func NewUserService(repo *repository.DatabaseRepository) *UserService {
	return &UserService{repo: repo}
}

func (u *UserService) RegisterUser(ctx context.Context, registerUser *model.RegisterUser) (string, error) {
	user := &model.User{
		ID:       uuid.NewString(),
		Login:    registerUser.Login,
		Password: registerUser.Password,
		Balance:  0,
	}

	err := u.repo.CreateUser(ctx, user)

	if err != nil {
		return "", err
	}

	return jwt.BuildJWTString()
}

func (u *UserService) LoginUser(ctx context.Context, loginUser *model.LoginUser) (string, error) {
	dbUser, err := u.repo.GetUser(ctx, loginUser.Login)

	if err != nil {
		if errors.Is(err, repository.ErrNoAuthUser) {
			return "", ErrInvalidLoginAndPassword
		}
		return "", err
	}

	if loginUser.Login == dbUser.Login && loginUser.Password == dbUser.Password {
		return jwt.BuildJWTString()
	}

	return "", ErrInvalidLoginAndPassword
}
