// internal/auth/middleware.go
package auth

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey string

const (
	ContextUserIDKey    ctxKey = "user_id"
	ContextCompanyIDKey ctxKey = "company_id"
	ContextRolesKey     ctxKey = "roles"
	ContextEmailKey     ctxKey = "email"
)

func JWTMiddleware(validator *JWTValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, claims, err := validator.ParseToken(tokenStr)
			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			// Use our UserClaims type if you like; or map claims.

			userID, _ := claims["sub"].(string)
			companyID, _ := claims["company_id"].(string)
			email, _ := claims["email"].(string)

			var roles []string
			if rRaw, ok := claims["roles"].([]any); ok {
				for _, rv := range rRaw {
					if s, ok := rv.(string); ok {
						roles = append(roles, s)
					}
				}
			}

			ctx := context.WithValue(r.Context(), ContextUserIDKey, userID)
			ctx = context.WithValue(ctx, ContextCompanyIDKey, companyID)
			ctx = context.WithValue(ctx, ContextEmailKey, email)
			ctx = context.WithValue(ctx, ContextRolesKey, roles)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper accessors
func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ContextUserIDKey).(string); ok {
		return v
	}
	return ""
}

func CompanyIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ContextCompanyIDKey).(string); ok {
		return v
	}
	return ""
}

func EmailFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ContextEmailKey).(string); ok {
		return v
	}
	return ""
}

func RolesFromContext(ctx context.Context) []string {
	if v, ok := ctx.Value(ContextRolesKey).([]string); ok {
		return v
	}
	return nil
}
