package interceptor

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"
)

func init() {
	pkgjwt.SetSecret("test-secret")
}

func validToken(t *testing.T) string {
	t.Helper()
	tok, err := pkgjwt.GenerateAccessToken("user-1", "admin")
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	return tok
}

func dummyHandler(_ context.Context, _ any) (any, error) {
	return "ok", nil
}

func dummyInfo(method string) *grpc.UnaryServerInfo {
	return &grpc.UnaryServerInfo{FullMethod: method}
}

func TestAuthInterceptor_ValidToken(t *testing.T) {
	interceptor := AuthUnaryInterceptor()
	tok := validToken(t)

	md := metadata.Pairs("authorization", "Bearer "+tok)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	var capturedCtx context.Context
	handler := func(ctx context.Context, req any) (any, error) {
		capturedCtx = ctx
		return "ok", nil
	}

	_, err := interceptor(ctx, nil, dummyInfo("/cms.v1.DocumentService/GetDocument"), handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	userID := middleware.UserID(capturedCtx)
	if userID != "user-1" {
		t.Errorf("userID = %q, want %q", userID, "user-1")
	}
	role := middleware.Role(capturedCtx)
	if role != "admin" {
		t.Errorf("role = %q, want %q", role, "admin")
	}
}

func TestAuthInterceptor_MissingToken(t *testing.T) {
	interceptor := AuthUnaryInterceptor()
	ctx := context.Background()

	_, err := interceptor(ctx, nil, dummyInfo("/cms.v1.DocumentService/SaveDocument"), dummyHandler)
	if err == nil {
		t.Fatal("expected error for missing token")
	}
	if s, ok := status.FromError(err); !ok || s.Code() != codes.Unauthenticated {
		t.Errorf("code = %v, want Unauthenticated", err)
	}
}

func TestAuthInterceptor_InvalidToken(t *testing.T) {
	interceptor := AuthUnaryInterceptor()

	md := metadata.Pairs("authorization", "Bearer invalid.token.here")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := interceptor(ctx, nil, dummyInfo("/cms.v1.DocumentService/SaveDocument"), dummyHandler)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
	if s, ok := status.FromError(err); !ok || s.Code() != codes.Unauthenticated {
		t.Errorf("code = %v, want Unauthenticated", err)
	}
}

func TestAuthInterceptor_PublicMethod_SkipsAuth(t *testing.T) {
	interceptor := AuthUnaryInterceptor()
	ctx := context.Background()

	resp, err := interceptor(ctx, nil, dummyInfo("/cms.v1.AuthService/SetupStatus"), dummyHandler)
	if err != nil {
		t.Fatalf("unexpected error for public method: %v", err)
	}
	if resp != "ok" {
		t.Errorf("resp = %v, want ok", resp)
	}
}
