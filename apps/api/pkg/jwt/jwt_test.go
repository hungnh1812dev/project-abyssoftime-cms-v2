package jwt_test

import (
	"testing"

	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"
)

func TestAccessTokenRoundTrip(t *testing.T) {
	token, err := pkgjwt.GenerateAccessToken("user-123", "admin")
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	claims, err := pkgjwt.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-123")
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %q, want %q", claims.Role, "admin")
	}
}

func TestRefreshTokenRoundTrip(t *testing.T) {
	token, err := pkgjwt.GenerateRefreshToken("user-456")
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}

	claims, err := pkgjwt.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UserID != "user-456" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-456")
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	_, err := pkgjwt.ValidateToken("not.a.valid.token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
