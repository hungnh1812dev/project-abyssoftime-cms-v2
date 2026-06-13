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

type mockDocumentUC struct {
	createFn    func(ctx context.Context, doc *entity.Document) error
	getOneFn    func(ctx context.Context, id string) (*entity.Document, error)
	getAllFn     func(ctx context.Context, contentTypeID string) ([]*entity.Document, error)
	updateFn    func(ctx context.Context, doc *entity.Document) error
	deleteFn    func(ctx context.Context, id string) error
	publishFn   func(ctx context.Context, id string) error
	unpublishFn func(ctx context.Context, id string) error
}

func (m *mockDocumentUC) Create(ctx context.Context, doc *entity.Document) error {
	return m.createFn(ctx, doc)
}
func (m *mockDocumentUC) GetOne(ctx context.Context, id string) (*entity.Document, error) {
	return m.getOneFn(ctx, id)
}
func (m *mockDocumentUC) GetAll(ctx context.Context, contentTypeID string) ([]*entity.Document, error) {
	return m.getAllFn(ctx, contentTypeID)
}
func (m *mockDocumentUC) Update(ctx context.Context, doc *entity.Document) error {
	return m.updateFn(ctx, doc)
}
func (m *mockDocumentUC) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}
func (m *mockDocumentUC) Publish(ctx context.Context, id string) error {
	return m.publishFn(ctx, id)
}
func (m *mockDocumentUC) Unpublish(ctx context.Context, id string) error {
	return m.unpublishFn(ctx, id)
}

// ---- List ------------------------------------------------------------------

func TestDocumentHandler_List(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.getAllFn = func(_ context.Context, contentTypeID string) ([]*entity.Document, error) {
		return []*entity.Document{{ID: "1", ContentTypeID: contentTypeID}}, nil
	}
	h := handler.NewDocumentHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/api/documents?contentType=ct-1", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("List() status = %d, want 200", w.Code)
	}
	var out []map[string]any
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(out) != 1 {
		t.Errorf("List() count = %d, want 1", len(out))
	}
}

// ---- Create ----------------------------------------------------------------

func TestDocumentHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		setupUC    func(*mockDocumentUC)
		wantStatus int
	}{
		{
			name: "201 on valid create",
			body: map[string]any{"contentTypeId": "ct-1", "data": map[string]any{"title": "Hello"}},
			setupUC: func(m *mockDocumentUC) {
				m.createFn = func(_ context.Context, doc *entity.Document) error {
					doc.ID = "new-id"
					return nil
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "400 on malformed JSON",
			body:       "not json",
			setupUC:    func(m *mockDocumentUC) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockDocumentUC{}
			tt.setupUC(uc)
			h := handler.NewDocumentHandler(uc)

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/documents", &buf)
			w := httptest.NewRecorder()
			h.Create(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Create() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// ---- GetByID ---------------------------------------------------------------

func TestDocumentHandler_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		setupUC    func(*mockDocumentUC)
		wantStatus int
	}{
		{
			name: "200 found",
			id:   "abc",
			setupUC: func(m *mockDocumentUC) {
				m.getOneFn = func(_ context.Context, id string) (*entity.Document, error) {
					return &entity.Document{ID: id}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "404 not found",
			id:   "missing",
			setupUC: func(m *mockDocumentUC) {
				m.getOneFn = func(_ context.Context, _ string) (*entity.Document, error) {
					return nil, pkgerrors.ErrNotFound
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockDocumentUC{}
			tt.setupUC(uc)
			h := handler.NewDocumentHandler(uc)

			req := httptest.NewRequest(http.MethodGet, "/api/documents/"+tt.id, nil)
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

func TestDocumentHandler_Update(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.updateFn = func(_ context.Context, _ *entity.Document) error { return nil }
	h := handler.NewDocumentHandler(uc)

	body := map[string]any{"data": map[string]any{"title": "Updated"}}
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(body)
	req := httptest.NewRequest(http.MethodPut, "/api/documents/abc", &buf)
	req.SetPathValue("id", "abc")
	w := httptest.NewRecorder()
	h.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Update() status = %d, want 200", w.Code)
	}
}

// ---- Delete ----------------------------------------------------------------

func TestDocumentHandler_Delete(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.deleteFn = func(_ context.Context, _ string) error { return nil }
	h := handler.NewDocumentHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/api/documents/abc", nil)
	req.SetPathValue("id", "abc")
	w := httptest.NewRecorder()
	h.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Delete() status = %d, want 204", w.Code)
	}
}

// ---- Publish / Unpublish ---------------------------------------------------

func TestDocumentHandler_Publish(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.publishFn = func(_ context.Context, _ string) error { return nil }
	h := handler.NewDocumentHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/documents/abc/publish", nil)
	req.SetPathValue("id", "abc")
	w := httptest.NewRecorder()
	h.Publish(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Publish() status = %d, want 200", w.Code)
	}
}

func TestDocumentHandler_Unpublish(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.unpublishFn = func(_ context.Context, _ string) error { return nil }
	h := handler.NewDocumentHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/documents/abc/unpublish", nil)
	req.SetPathValue("id", "abc")
	w := httptest.NewRecorder()
	h.Unpublish(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Unpublish() status = %d, want 200", w.Code)
	}
}
