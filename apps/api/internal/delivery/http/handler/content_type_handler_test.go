package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/handler"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

// ---- mock usecase ----------------------------------------------------------

type mockContentTypeUC struct {
	findByIDFn func(ctx context.Context, id string) (*entity.ContentType, error)
	findAllFn  func(ctx context.Context) ([]*entity.ContentType, error)
}

func (m *mockContentTypeUC) FindByID(ctx context.Context, id string) (*entity.ContentType, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockContentTypeUC) FindAll(ctx context.Context) ([]*entity.ContentType, error) {
	return m.findAllFn(ctx)
}

// ---- List ------------------------------------------------------------------

func TestContentTypeHandler_List(t *testing.T) {
	uc := &mockContentTypeUC{}
	uc.findAllFn = func(_ context.Context) ([]*entity.ContentType, error) {
		return []*entity.ContentType{
			{ID: "1", Slug: "blog"},
			{ID: "2", Slug: "homepage"},
		}, nil
	}
	h := handler.NewContentTypeHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/api/content-types", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("List() status = %d, want 200", w.Code)
	}
	var out []map[string]any
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("List() count = %d, want 2", len(out))
	}
}

// ---- GetByID ---------------------------------------------------------------

func TestContentTypeHandler_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		setupUC    func(*mockContentTypeUC)
		wantStatus int
	}{
		{
			name: "200 found",
			id:   "abc",
			setupUC: func(m *mockContentTypeUC) {
				m.findByIDFn = func(_ context.Context, id string) (*entity.ContentType, error) {
					return &entity.ContentType{ID: id, Slug: "blog"}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "404 not found",
			id:   "missing",
			setupUC: func(m *mockContentTypeUC) {
				m.findByIDFn = func(_ context.Context, _ string) (*entity.ContentType, error) {
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

			req := httptest.NewRequest(http.MethodGet, "/api/content-types/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			w := httptest.NewRecorder()
			h.GetByID(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetByID() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
