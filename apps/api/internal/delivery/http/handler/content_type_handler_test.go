package handler_test

import (
	"bytes"
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
	createFn   func(ctx context.Context, ct *entity.ContentType) error
	findByIDFn func(ctx context.Context, id string) (*entity.ContentType, error)
	findAllFn  func(ctx context.Context) ([]*entity.ContentType, error)
	updateFn   func(ctx context.Context, ct *entity.ContentType) error
	deleteFn   func(ctx context.Context, id string) error
}

func (m *mockContentTypeUC) Create(ctx context.Context, ct *entity.ContentType) error {
	return m.createFn(ctx, ct)
}
func (m *mockContentTypeUC) FindByID(ctx context.Context, id string) (*entity.ContentType, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockContentTypeUC) FindAll(ctx context.Context) ([]*entity.ContentType, error) {
	return m.findAllFn(ctx)
}
func (m *mockContentTypeUC) Update(ctx context.Context, ct *entity.ContentType) error {
	return m.updateFn(ctx, ct)
}
func (m *mockContentTypeUC) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

// ---- Create ----------------------------------------------------------------

func TestContentTypeHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		setupUC    func(*mockContentTypeUC)
		wantStatus int
	}{
		{
			name: "201 on valid create",
			body: map[string]any{"name": "Blog", "slug": "blog", "kind": "collection"},
			setupUC: func(m *mockContentTypeUC) {
				m.createFn = func(_ context.Context, ct *entity.ContentType) error {
					ct.ID = "new-id"
					return nil
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "409 on duplicate slug",
			body: map[string]any{"name": "Blog", "slug": "blog", "kind": "collection"},
			setupUC: func(m *mockContentTypeUC) {
				m.createFn = func(_ context.Context, _ *entity.ContentType) error {
					return pkgerrors.ErrConflict
				}
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "400 on invalid kind",
			body: map[string]any{"name": "Bad", "slug": "bad", "kind": "unknown"},
			setupUC: func(m *mockContentTypeUC) {
				m.createFn = func(_ context.Context, _ *entity.ContentType) error {
					return pkgerrors.ErrBadRequest
				}
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 on malformed JSON",
			body:       "not json",
			setupUC:    func(m *mockContentTypeUC) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockContentTypeUC{}
			tt.setupUC(uc)
			h := handler.NewContentTypeHandler(uc)

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/content-types", &buf)
			w := httptest.NewRecorder()
			h.Create(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Create() status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
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

// ---- Update ----------------------------------------------------------------

func TestContentTypeHandler_Update(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       any
		setupUC    func(*mockContentTypeUC)
		wantStatus int
	}{
		{
			name: "200 on valid update",
			id:   "abc",
			body: map[string]any{"name": "Updated", "slug": "blog", "kind": "collection"},
			setupUC: func(m *mockContentTypeUC) {
				m.updateFn = func(_ context.Context, _ *entity.ContentType) error { return nil }
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "404 not found",
			id:   "missing",
			body: map[string]any{"name": "X", "slug": "x", "kind": "single"},
			setupUC: func(m *mockContentTypeUC) {
				m.updateFn = func(_ context.Context, _ *entity.ContentType) error { return pkgerrors.ErrNotFound }
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "400 invalid kind",
			id:   "abc",
			body: map[string]any{"name": "X", "slug": "x", "kind": "bad"},
			setupUC: func(m *mockContentTypeUC) {
				m.updateFn = func(_ context.Context, _ *entity.ContentType) error { return pkgerrors.ErrBadRequest }
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockContentTypeUC{}
			tt.setupUC(uc)
			h := handler.NewContentTypeHandler(uc)

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/api/content-types/"+tt.id, &buf)
			req.SetPathValue("id", tt.id)
			w := httptest.NewRecorder()
			h.Update(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Update() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// ---- Delete ----------------------------------------------------------------

func TestContentTypeHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		setupUC    func(*mockContentTypeUC)
		wantStatus int
	}{
		{
			name: "204 on delete",
			id:   "abc",
			setupUC: func(m *mockContentTypeUC) {
				m.deleteFn = func(_ context.Context, _ string) error { return nil }
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "404 not found",
			id:   "missing",
			setupUC: func(m *mockContentTypeUC) {
				m.deleteFn = func(_ context.Context, _ string) error { return pkgerrors.ErrNotFound }
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockContentTypeUC{}
			tt.setupUC(uc)
			h := handler.NewContentTypeHandler(uc)

			req := httptest.NewRequest(http.MethodDelete, "/api/content-types/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			w := httptest.NewRecorder()
			h.Delete(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Delete() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
