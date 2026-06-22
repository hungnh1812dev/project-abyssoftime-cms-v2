package middleware

import (
	"context"
	"net/http"
	"strings"

	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"
)

type contextKey string

const (
	ContextKeyUserID contextKey = "userID"
	ContextKeyRole   contextKey = "role"
)

// UserID extracts the authenticated user's ID from the context.
func UserID(ctx context.Context) string {
	if v, ok := ctx.Value(ContextKeyUserID).(string); ok {
		return v
	}
	return ""
}

// Role extracts the authenticated user's role from the context.
func Role(ctx context.Context) string {
	if v, ok := ctx.Value(ContextKeyRole).(string); ok {
		return v
	}
	return ""
}

// WithRole returns a context with the given role injected (useful in tests).
func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, ContextKeyRole, role)
}

// Auth validates the Bearer token from the Authorization header and
// injects userID + role into the request context.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		header := request.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(writer, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		claims, err := pkgjwt.ValidateToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			http.Error(writer, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(request.Context(), ContextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, ContextKeyRole, claims.Role)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

// RequireRole returns 403 if the role in context does not match required.
func RequireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if Role(request.Context()) != role {
			http.Error(writer, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
