package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

func newTestCache() *middleware.RoleCache {
	cache := middleware.NewRoleCache()
	cache.Load([]*entity.RoleEntity{
		{Slug: "admin", Permissions: []string{"content:read", "content:create", "users:manage"}, Level: 80},
		{Slug: "editor", Permissions: []string{"content:read", "content:create"}, Level: 60},
		{Slug: "guest", Permissions: []string{"content:read"}, Level: 20},
	})
	return cache
}

func TestRoleCache_HasPermission(t *testing.T) {
	cache := newTestCache()

	tests := []struct {
		role       string
		perm       string
		wantResult bool
	}{
		{"admin", "content:read", true},
		{"admin", "users:manage", true},
		{"editor", "content:create", true},
		{"editor", "users:manage", false},
		{"guest", "content:read", true},
		{"guest", "content:create", false},
		{"unknown", "content:read", false},
	}
	for _, tt := range tests {
		t.Run(tt.role+"_"+tt.perm, func(t *testing.T) {
			got := cache.HasPermission(tt.role, tt.perm)
			if got != tt.wantResult {
				t.Errorf("HasPermission(%q, %q) = %v, want %v", tt.role, tt.perm, got, tt.wantResult)
			}
		})
	}
}

func TestRoleCache_GetLevel(t *testing.T) {
	cache := newTestCache()

	if got := cache.GetLevel("admin"); got != 80 {
		t.Errorf("GetLevel(admin) = %d, want 80", got)
	}
	if got := cache.GetLevel("unknown"); got != 0 {
		t.Errorf("GetLevel(unknown) = %d, want 0", got)
	}
}

func TestRoleCache_Load_Replaces(t *testing.T) {
	cache := newTestCache()

	if !cache.HasPermission("admin", "users:manage") {
		t.Fatal("expected admin to have users:manage before reload")
	}

	cache.Load([]*entity.RoleEntity{
		{Slug: "admin", Permissions: []string{"content:read"}, Level: 80},
	})

	if cache.HasPermission("admin", "users:manage") {
		t.Error("expected admin to lose users:manage after reload")
	}
	if cache.HasPermission("editor", "content:read") {
		t.Error("expected editor to be gone after reload")
	}
}

func TestGinRequirePermission_Allowed(t *testing.T) {
	cache := newTestCache()
	tok := ginValidToken(t, "user-1", "admin")

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	called := false
	r.GET("/test", middleware.GinAuth(), middleware.GinRequirePermission(cache, "users:manage"), func(c *gin.Context) {
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
		t.Error("handler must be called when permission is present")
	}
}

func TestGinRequirePermission_Forbidden(t *testing.T) {
	cache := newTestCache()
	tok := ginValidToken(t, "user-1", "guest")

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	called := false
	r.GET("/test", middleware.GinAuth(), middleware.GinRequirePermission(cache, "users:manage"), func(c *gin.Context) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
	if called {
		t.Error("handler must not be called when permission is missing")
	}
}

func TestGinRequirePermission_UnknownRole(t *testing.T) {
	cache := newTestCache()
	tok := ginValidToken(t, "user-1", "nonexistent_role")

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.GET("/test", middleware.GinAuth(), middleware.GinRequirePermission(cache, "content:read"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403 for unknown role", w.Code)
	}
}
