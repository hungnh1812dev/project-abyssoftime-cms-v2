package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/handler"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

type mockContentTypeUC struct {
	findByIDFn   func(ctx context.Context, id string) (*entity.ContentType, error)
	findBySlugFn func(ctx context.Context, slug string) (*entity.ContentType, error)
	findAllFn    func(ctx context.Context) ([]*entity.ContentType, error)
}

func (m *mockContentTypeUC) FindByID(ctx context.Context, id string) (*entity.ContentType, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockContentTypeUC) FindBySlug(ctx context.Context, slug string) (*entity.ContentType, error) {
	return m.findBySlugFn(ctx, slug)
}
func (m *mockContentTypeUC) FindAll(ctx context.Context) ([]*entity.ContentType, error) {
	return m.findAllFn(ctx)
}

func TestContentTypeHandler_ListSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockContentTypeUC{}
	uc.findAllFn = func(_ context.Context) ([]*entity.ContentType, error) {
		return []*entity.ContentType{
			{DocumentID: "1", Name: "Blog", Slug: "blog", Kind: "collection"},
			{DocumentID: "2", Name: "Homepage", Slug: "homepage", Kind: "single"},
		}, nil
	}
	h := handler.NewContentTypeHandler(uc)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/api/content-types", h.ListSummary)

	req := httptest.NewRequest(http.MethodGet, "/api/content-types", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ListSummary() status = %d, want 200", w.Code)
	}
	var out []map[string]any
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("ListSummary() count = %d, want 2", len(out))
	}
}

func TestContentTypeHandler_ListSummary_ExcludesFieldsAndTimestamps(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockContentTypeUC{}
	uc.findAllFn = func(_ context.Context) ([]*entity.ContentType, error) {
		return []*entity.ContentType{
			{
				DocumentID: "1",
				Name: "Blog",
				Slug: "blog",
				Kind: "collection",
				Fields: []entity.FieldDefinition{
					{Name: "title", Type: "text"},
				},
			},
		}, nil
	}
	h := handler.NewContentTypeHandler(uc)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/api/content-types", h.ListSummary)

	req := httptest.NewRequest(http.MethodGet, "/api/content-types", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ListSummary() status = %d, want 200", w.Code)
	}
	var out []map[string]any
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("ListSummary() count = %d, want 1", len(out))
	}
	for _, key := range []string{"Fields", "CreatedAt", "UpdatedAt"} {
		if _, ok := out[0][key]; ok {
			t.Errorf("ListSummary() should not contain %q", key)
		}
	}
	for _, key := range []string{"ID", "Name", "Slug", "Kind"} {
		if _, ok := out[0][key]; !ok {
			t.Errorf("ListSummary() missing expected key %q", key)
		}
	}
}

func TestContentTypeHandler_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		identifier string
		setupUC    func(*mockContentTypeUC)
		wantStatus int
	}{
		{
			name:       "200 found by ObjectID",
			identifier: "aabbccddeeff00112233aabb",
			setupUC: func(m *mockContentTypeUC) {
				m.findByIDFn = func(_ context.Context, id string) (*entity.ContentType, error) {
					return &entity.ContentType{DocumentID: id, Slug: "blog"}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "200 found by slug",
			identifier: "blog",
			setupUC: func(m *mockContentTypeUC) {
				m.findBySlugFn = func(_ context.Context, slug string) (*entity.ContentType, error) {
					return &entity.ContentType{DocumentID: "abc", Slug: slug}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "404 not found by ObjectID",
			identifier: "aabbccddeeff00112233aabb",
			setupUC: func(m *mockContentTypeUC) {
				m.findByIDFn = func(_ context.Context, _ string) (*entity.ContentType, error) {
					return nil, pkgerrors.ErrNotFound
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "404 not found by slug",
			identifier: "missing",
			setupUC: func(m *mockContentTypeUC) {
				m.findBySlugFn = func(_ context.Context, _ string) (*entity.ContentType, error) {
					return nil, pkgerrors.ErrNotFound
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockContentTypeUC{}
			tt.setupUC(uc)
			h := handler.NewContentTypeHandler(uc)

			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.GET("/api/content-types/:identifier", h.Get)

			req := httptest.NewRequest(http.MethodGet, "/api/content-types/"+tt.identifier, nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Get() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
