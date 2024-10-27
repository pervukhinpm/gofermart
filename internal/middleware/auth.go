package middleware

import (
	"context"
	"github.com/pervukhinpm/gophermart/internal/jwt"
	"net/http"
)

const CookieName = "jwt"

func Auth(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			var tokenString string

			cookie, err := r.Cookie(CookieName)

			if err != nil {
				tokenString, err = jwt.BuildJWTString()

				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:  CookieName,
					Value: tokenString,
				})
			} else {
				tokenString = cookie.Value
			}

			userID, err := jwt.GetUserID(tokenString)

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := setUserID(r.Context(), userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}(next)
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
