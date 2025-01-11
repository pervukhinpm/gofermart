package middleware

import (
	"context"
	"github.com/pervukhinpm/gophermart/internal/config"
	"github.com/pervukhinpm/gophermart/internal/jwt"
	"net/http"
	"strings"
)

const (
	bearerHeader        = "Bearer "
	authorizationHeader = "Authorization"
)

func Auth(config config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get(authorizationHeader)
			if authHeader == "" || !strings.HasPrefix(authHeader, bearerHeader) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, bearerHeader)

			userID, err := jwt.GetUserID(tokenString, config)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := setUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type UserID struct {
	value string
}

func setUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserID{}, userID)
}

func GetUserID(ctx context.Context) string {
	userID, ok := ctx.Value(UserID{}).(string)
	if !ok {
		return ""
	}

	return userID
}
