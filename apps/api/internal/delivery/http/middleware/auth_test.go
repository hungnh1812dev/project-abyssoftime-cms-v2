package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"
)

func validToken(t *testing.T, userID, role string) string {
	t.Helper()
	tok, err := pkgjwt.GenerateAccessToken(userID, role)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	return tok
}

// captureHandler records the userID and role it receives in context.
type captureHandler struct {
	userID string
	role   string
	called bool
}

func (c *captureHandler) ServeHTTP(_ http.ResponseWriter, r *http.Request) {
	c.called = true
	c.userID = middleware.UserID(r.Context())
	c.role = middleware.Role(r.Context())
}

// ---- Auth middleware -------------------------------------------------------

func TestAuthMiddleware_ValidToken(t *testing.T) {
	tok := validToken(t, "user-1", "admin")
	next := &captureHandler{}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()

	middleware.Auth(next).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !next.called {
		t.Fatal("next handler was not called")
	}
	if next.userID != "user-1" {
		t.Errorf("userID = %q, want %q", next.userID, "user-1")
	}
	if next.role != "admin" {
		t.Errorf("role = %q, want %q", next.role, "admin")
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	next := &captureHandler{}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	middleware.Auth(next).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
	if next.called {
		t.Error("next handler must not be called when token is missing")
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	next := &captureHandler{}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer not.a.valid.token")
	w := httptest.NewRecorder()

	middleware.Auth(next).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// ---- RequireRole -----------------------------------------------------------

func TestRequireRole_Allowed(t *testing.T) {
	next := &captureHandler{}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := middleware.WithRole(r.Context(), "admin")
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	middleware.RequireRole("admin", next).ServeHTTP(w, r)

	if !next.called {
		t.Error("next handler must be called when role matches")
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	next := &captureHandler{}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := middleware.WithRole(r.Context(), "guest")
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	middleware.RequireRole("admin", next).ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
	if next.called {
		t.Error("next handler must not be called when role is wrong")
	}
}
