package resolver

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"
)

// AuthDirective validates the Bearer JWT from the Authorization header and
// injects userID + role into ctx using the same keys as the REST middleware,
// so resolver methods can call middleware.UserID(ctx) identically to handlers.
func AuthDirective(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	r, ok := ctx.Value(RequestCtxKey).(*http.Request)
	if !ok || r == nil {
		return nil, errors.New("unauthorized")
	}
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return nil, errors.New("unauthorized")
	}
	claims, err := pkgjwt.ValidateToken(strings.TrimPrefix(header, "Bearer "))
	if err != nil {
		return nil, errors.New("unauthorized")
	}
	ctx = context.WithValue(ctx, middleware.ContextKeyUserID, claims.UserID)
	ctx = context.WithValue(ctx, middleware.ContextKeyRole, claims.Role)
	return next(ctx)
}
