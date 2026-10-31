package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/adexcell/crm-ai/internal/infra/http/response"
	"github.com/adexcell/crm-ai/internal/infra/jwt"
	"github.com/google/uuid"
)

type userIDKeyType string

const userIDKey userIDKeyType = "user_id"

func Auth(jwtManager *jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			tokenStr := strings.TrimPrefix(header, "Bearer ")
			if tokenStr == header || tokenStr == "" {
				response.Error(w,
					http.StatusUnauthorized,
					response.CodeUnauthorized,
					"missing bearer token",
					nil,
				)
				return
			}

			claims, err := jwtManager.Validate(tokenStr)
			if err != nil {
				response.Error(
					w,
					http.StatusUnauthorized,
					response.CodeUnauthorized,
					"invalid token",
					nil,
				)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext возвращает полный объект клеймов из JWT
func ClaimsFromContext(ctx context.Context) (*jwt.Claims, bool) {
	claims, ok := ctx.Value(userIDKey).(*jwt.Claims)
	return claims, ok
}

// UserIDFromContext возвращает конкретный ID пользователя
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	if claims, ok := ClaimsFromContext(ctx); ok {
		return claims.UserID, true
	}
	return uuid.Nil, false
}
