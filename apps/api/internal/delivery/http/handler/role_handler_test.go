package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	role "project-abyssoftime-cms-v2/api/internal/usecase/role"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type mockRoleUseCase struct {
	findAllFn  func(ctx context.Context) ([]*entity.RoleEntity, error)
	findByIDFn func(ctx context.Context, id string) (*entity.RoleEntity, error)
	createFn   func(ctx context.Context, input role.CreateRoleInput, callerLevel int) (*entity.RoleEntity, error)
	updateFn   func(ctx context.Context, id string, input role.UpdateRoleInput, callerLevel int) (*entity.RoleEntity, error)
	deleteFn   func(ctx context.Context, id string) error
}

func (m *mockRoleUseCase) FindAll(ctx context.Context) ([]*entity.RoleEntity, error) {
	return m.findAllFn(ctx)
}
func (m *mockRoleUseCase) FindByID(ctx context.Context, id string) (*entity.RoleEntity, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockRoleUseCase) Create(ctx context.Context, input role.CreateRoleInput, callerLevel int) (*entity.RoleEntity, error) {
	return m.createFn(ctx, input, callerLevel)
}
func (m *mockRoleUseCase) Update(ctx context.Context, id string, input role.UpdateRoleInput, callerLevel int) (*entity.RoleEntity, error) {
	return m.updateFn(ctx, id, input, callerLevel)
}
func (m *mockRoleUseCase) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

func setupRoleHandler() (*RoleHandler, *middleware.RoleCache) {
	cache := middleware.NewRoleCache()
	cache.Load([]*entity.RoleEntity{
		{Slug: "super_admin", Permissions: entity.AllPermissionStrings(), Level: 100},
	})
	uc := &mockRoleUseCase{
		findAllFn: func(_ context.Context) ([]*entity.RoleEntity, error) {
			return []*entity.RoleEntity{
				{DocumentID: "d1", Name: "Admin", Slug: "admin", Level: 80},
				{DocumentID: "d2", Name: "Guest", Slug: "guest", Level: 20},
			}, nil
		},
		findByIDFn: func(_ context.Context, id string) (*entity.RoleEntity, error) {
			if id == "d1" {
				return &entity.RoleEntity{DocumentID: "d1", Name: "Admin", Slug: "admin", Level: 80}, nil
			}
			return nil, pkgerrors.ErrNotFound
		},
		createFn: func(_ context.Context, input role.CreateRoleInput, _ int) (*entity.RoleEntity, error) {
			return &entity.RoleEntity{DocumentID: "d-new", Name: input.Name, Slug: input.Slug, Level: input.Level, Permissions: input.Permissions}, nil
		},
		updateFn: func(_ context.Context, id string, input role.UpdateRoleInput, _ int) (*entity.RoleEntity, error) {
			r := &entity.RoleEntity{DocumentID: id, Name: "Admin", Slug: "admin", Level: 80}
			if input.Name != nil {
				r.Name = *input.Name
			}
			return r, nil
		},
		deleteFn: func(_ context.Context, id string) error {
			if id == "default-role" {
				return pkgerrors.ErrValidation
			}
			return nil
		},
	}
	h := NewRoleHandler(uc, cache)
	return h, cache
}

func TestRoleHandler_List(t *testing.T) {
	h, _ := setupRoleHandler()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/api/roles", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/roles", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var roles []entity.RoleEntity
	if err := json.NewDecoder(w.Body).Decode(&roles); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("len(roles) = %d, want 2", len(roles))
	}
}

func TestRoleHandler_Get(t *testing.T) {
	h, _ := setupRoleHandler()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/api/roles/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/roles/d1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRoleHandler_Get_NotFound(t *testing.T) {
	h, _ := setupRoleHandler()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/api/roles/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/roles/nonexistent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestRoleHandler_Create(t *testing.T) {
	h, _ := setupRoleHandler()

	body, _ := json.Marshal(map[string]any{
		"name":        "Content Manager",
		"slug":        "content-manager",
		"permissions": []string{"content:read"},
		"level":       50,
	})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/api/roles", func(c *gin.Context) {
		c.Set("role", "super_admin")
		h.Create(c)
	})

	c.Request = httptest.NewRequest(http.MethodPost, "/api/roles", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201; body = %s", w.Code, w.Body.String())
	}
}

func TestRoleHandler_Update(t *testing.T) {
	h, _ := setupRoleHandler()

	newName := "Admin Updated"
	body, _ := json.Marshal(map[string]any{"name": newName})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.PUT("/api/roles/:id", func(c *gin.Context) {
		c.Set("role", "super_admin")
		h.Update(c)
	})

	c.Request = httptest.NewRequest(http.MethodPut, "/api/roles/d1", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}
}

func TestRoleHandler_Delete(t *testing.T) {
	h, _ := setupRoleHandler()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.DELETE("/api/roles/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/roles/d1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

func TestRoleHandler_Delete_DefaultRole(t *testing.T) {
	h, _ := setupRoleHandler()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.DELETE("/api/roles/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/roles/default-role", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}
