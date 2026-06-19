package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func ginValidToken(t *testing.T, userID, role string) string {
	t.Helper()
	tok, err := pkgjwt.GenerateAccessToken(userID, role)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	return tok
}

func TestGinAuth_ValidToken(t *testing.T) {
	tok := ginValidToken(t, "user-1", "admin")

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	var capturedUserID, capturedRole string
	var contextUserID, contextRole string

	r.GET("/test", middleware.GinAuth(), func(c *gin.Context) {
		capturedUserID = c.GetString("userID")
		capturedRole = c.GetString("role")
		contextUserID = middleware.UserID(c.Request.Context())
		contextRole = middleware.Role(c.Request.Context())
		c.Status(http.StatusOK)
	})

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if capturedUserID != "user-1" {
		t.Errorf("gin userID = %q, want %q", capturedUserID, "user-1")
	}
	if capturedRole != "admin" {
		t.Errorf("gin role = %q, want %q", capturedRole, "admin")
	}
	if contextUserID != "user-1" {
		t.Errorf("context userID = %q, want %q", contextUserID, "user-1")
	}
	if contextRole != "admin" {
		t.Errorf("context role = %q, want %q", contextRole, "admin")
	}
}

func TestGinAuth_MissingHeader(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	called := false
	r.GET("/test", middleware.GinAuth(), func(c *gin.Context) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
	if called {
		t.Error("next handler must not be called when token is missing")
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "unauthorized" {
		t.Errorf("error = %q, want %q", body["error"], "unauthorized")
	}
}

func TestGinAuth_InvalidToken(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.GET("/test", middleware.GinAuth(), func(c *gin.Context) {})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestGinRequireRole_Allowed(t *testing.T) {
	tok := ginValidToken(t, "user-1", "admin")

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	called := false
	r.GET("/test", middleware.GinAuth(), middleware.GinRequireRole("admin"), func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !called {
		t.Error("next handler must be called when role matches")
	}
}

func TestGinRequireRole_Forbidden(t *testing.T) {
	tok := ginValidToken(t, "user-1", "guest")

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	called := false
	r.GET("/test", middleware.GinAuth(), middleware.GinRequireRole("admin"), func(c *gin.Context) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
	if called {
		t.Error("next handler must not be called when role is wrong")
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "forbidden" {
		t.Errorf("error = %q, want %q", body["error"], "forbidden")
	}
}
