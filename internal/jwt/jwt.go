package jwt

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pervukhinpm/gophermart/internal/config"
	"time"
)

type Claims struct {
	UserID    string           `json:"user_id"`
	ExpiresAt *jwt.NumericDate `json:"exp"`
}

func (c Claims) Valid() error {
	if c.ExpiresAt != nil && c.ExpiresAt.Time.Before(time.Now()) {
		return errors.New("token has expired")
	}
	return nil
}

func GenerateJWT(userID string, config config.Config) (string, error) {
	claims := Claims{
		UserID:    userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.TokenExpiration)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(config.TokenSecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign the token: %w", err)
	}

	return tokenString, nil
}

func GetUserID(tokenString string, config config.Config) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.TokenSecretKey), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", fmt.Errorf("token is not valid")
	}
	return claims.UserID, nil
}
