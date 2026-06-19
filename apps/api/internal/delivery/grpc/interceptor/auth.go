package interceptor

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"
)

var publicMethods = map[string]bool{
	"/cms.v1.AuthService/SetupStatus":             true,
	"/cms.v1.AuthService/Login":                   true,
	"/cms.v1.AuthService/Register":                true,
	"/cms.v1.AuthService/Refresh":                 true,
	"/cms.v1.DocumentService/GetDocument":         true,
	"/cms.v1.DocumentService/GetSingleType":       true,
	"/cms.v1.ContentTypeService/ListContentTypes": true,
	"/cms.v1.ContentTypeService/GetContentType":   true,
}

func AuthUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if publicMethods[info.FullMethod] {
			ctx = tryInjectAuth(ctx)
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		vals := md.Get("authorization")
		if len(vals) == 0 || !strings.HasPrefix(vals[0], "Bearer ") {
			return nil, status.Error(codes.Unauthenticated, "missing bearer token")
		}

		claims, err := pkgjwt.ValidateToken(strings.TrimPrefix(vals[0], "Bearer "))
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, middleware.ContextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, middleware.ContextKeyRole, claims.Role)
		return handler(ctx, req)
	}
}

func tryInjectAuth(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	vals := md.Get("authorization")
	if len(vals) == 0 || !strings.HasPrefix(vals[0], "Bearer ") {
		return ctx
	}
	claims, err := pkgjwt.ValidateToken(strings.TrimPrefix(vals[0], "Bearer "))
	if err != nil {
		return ctx
	}
	ctx = context.WithValue(ctx, middleware.ContextKeyUserID, claims.UserID)
	ctx = context.WithValue(ctx, middleware.ContextKeyRole, claims.Role)
	return ctx
}
