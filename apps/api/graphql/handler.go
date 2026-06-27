package graphql

import (
	"context"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"

	"project-abyssoftime-cms-v2/api/graphql/generated"
	"project-abyssoftime-cms-v2/api/graphql/resolver"
	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"
)

type AccessTokenValidator interface {
	Validate(ctx context.Context, rawToken string) (*entity.AccessToken, error)
}

func NewHandler(gqlResolver *resolver.Resolver, tokenValidator AccessTokenValidator) http.Handler {
	config := generated.Config{Resolvers: gqlResolver}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(config))

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()

		authHeader := request.Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" || token == authHeader {
			http.Error(writer, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		if claims, err := pkgjwt.ValidateToken(token); err == nil {
			ctx = context.WithValue(ctx, middleware.ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, middleware.ContextKeyRole, claims.Role)
		} else if tokenValidator != nil {
			if _, err := tokenValidator.Validate(ctx, token); err != nil {
				http.Error(writer, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
		} else {
			http.Error(writer, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		srv.ServeHTTP(writer, request.WithContext(ctx))
	})
}
