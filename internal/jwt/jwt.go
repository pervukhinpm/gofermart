package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pervukhinpm/gophermart/internal/config"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func GenerateJWT(userID string, config config.Config) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.TokenExpiration)),
		},
		UserID: userID,
	})
	tokenString, err := token.SignedString([]byte(config.TokenSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetUserID(tokenString string, config config.Config) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signed method: %v", t.Header["alg"])
		}
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
