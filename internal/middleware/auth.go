package middleware

import (
	"context"
	"github.com/pervukhinpm/gophermart/internal/jwt"
	"net/http"
	"strings"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		userID, err := jwt.GetUserID(tokenString)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := setUserID(r.Context(), userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
