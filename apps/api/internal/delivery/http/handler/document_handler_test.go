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
	saveFn         func(ctx context.Context, contentTypeSlug string, doc *entity.Document, userID string) (*entity.Document, error)
	getForEditFn   func(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, string, error)
	getPublishedFn func(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error)
	getAllFn       func(ctx context.Context, contentTypeSlug string) ([]*entity.Document, error)
	deleteFn       func(ctx context.Context, contentTypeSlug, documentID string) error
	publishFn      func(ctx context.Context, contentTypeSlug, documentID, locale, userID string) error
	unpublishFn    func(ctx context.Context, contentTypeSlug, documentID, locale string) error
}

func (m *mockDocumentUC) Save(ctx context.Context, contentTypeSlug string, doc *entity.Document, userID string) (*entity.Document, error) {
	return m.saveFn(ctx, contentTypeSlug, doc, userID)
}
func (m *mockDocumentUC) GetForEdit(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, string, error) {
	return m.getForEditFn(ctx, contentTypeSlug, documentID, locale)
}
func (m *mockDocumentUC) GetPublished(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error) {
	return m.getPublishedFn(ctx, contentTypeSlug, documentID, locale)
}
func (m *mockDocumentUC) GetAll(ctx context.Context, contentTypeSlug string) ([]*entity.Document, error) {
	return m.getAllFn(ctx, contentTypeSlug)
}
func (m *mockDocumentUC) Delete(ctx context.Context, contentTypeSlug, documentID string) error {
	return m.deleteFn(ctx, contentTypeSlug, documentID)
}
func (m *mockDocumentUC) Publish(ctx context.Context, contentTypeSlug, documentID, locale, userID string) error {
	return m.publishFn(ctx, contentTypeSlug, documentID, locale, userID)
}
func (m *mockDocumentUC) Unpublish(ctx context.Context, contentTypeSlug, documentID, locale string) error {
	return m.unpublishFn(ctx, contentTypeSlug, documentID, locale)
}

// ---- List ------------------------------------------------------------------

func TestDocumentHandler_List(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.getAllFn = func(_ context.Context, slug string) ([]*entity.Document, error) {
		return []*entity.Document{{DocumentID: "1", ContentTypeID: "ct-1"}}, nil
	}
	uc.getForEditFn = func(_ context.Context, _, documentID, _ string) (*entity.Document, string, error) {
		return &entity.Document{DocumentID: documentID}, "draft", nil
	}
	h := handler.NewDocumentHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/api/content-types/articles/documents", nil)
	req.SetPathValue("slug", "articles")
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
	if out[0]["Status"] != "draft" {
		t.Errorf("List() Status field = %v, want draft", out[0]["Status"])
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
			body: map[string]any{"data": map[string]any{"title": "Hello"}},
			setupUC: func(m *mockDocumentUC) {
				m.saveFn = func(_ context.Context, _ string, doc *entity.Document, _ string) (*entity.Document, error) {
					doc.DocumentID = "new-entry"
					return doc, nil
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
			req := httptest.NewRequest(http.MethodPost, "/api/content-types/articles/documents", &buf)
			req.SetPathValue("slug", "articles")
			w := httptest.NewRecorder()
			h.Create(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Create() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// ---- GetByID (admin, draft + status) ---------------------------------------

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
				m.getForEditFn = func(_ context.Context, _, documentID, _ string) (*entity.Document, string, error) {
					return &entity.Document{DocumentID: documentID}, "draft", nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "404 not found",
			id:   "missing",
			setupUC: func(m *mockDocumentUC) {
				m.getForEditFn = func(_ context.Context, _, _, _ string) (*entity.Document, string, error) {
					return nil, "", pkgerrors.ErrNotFound
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

			req := httptest.NewRequest(http.MethodGet, "/api/content-types/articles/documents/"+tt.id, nil)
			req.SetPathValue("slug", "articles")
			req.SetPathValue("documentId", tt.id)
			w := httptest.NewRecorder()
			h.GetByID(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetByID() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestDocumentHandler_GetByID_ForwardsLocaleQueryParam(t *testing.T) {
	var gotLocale string
	uc := &mockDocumentUC{}
	uc.getForEditFn = func(_ context.Context, _, documentID, locale string) (*entity.Document, string, error) {
		gotLocale = locale
		return &entity.Document{DocumentID: documentID}, "draft", nil
	}
	h := handler.NewDocumentHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/api/content-types/articles/documents/abc?locale=vi", nil)
	req.SetPathValue("slug", "articles")
	req.SetPathValue("documentId", "abc")
	w := httptest.NewRecorder()
	h.GetByID(w, req)

	if gotLocale != "vi" {
		t.Errorf("GetByID() forwarded locale = %q, want vi", gotLocale)
	}
}

// ---- GetPublic (public/content read) ----------------------------------------

func TestDocumentHandler_GetPublic(t *testing.T) {
	tests := []struct {
		name       string
		setupUC    func(*mockDocumentUC)
		wantStatus int
	}{
		{
			name: "200 when published",
			setupUC: func(m *mockDocumentUC) {
				m.getPublishedFn = func(_ context.Context, _, documentID, _ string) (*entity.Document, error) {
					return &entity.Document{DocumentID: documentID}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "404 when never published",
			setupUC: func(m *mockDocumentUC) {
				m.getPublishedFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
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

			req := httptest.NewRequest(http.MethodGet, "/api/public/content-types/articles/documents/abc", nil)
			req.SetPathValue("slug", "articles")
			req.SetPathValue("documentId", "abc")
			w := httptest.NewRecorder()
			h.GetPublic(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetPublic() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestDocumentHandler_GetPublic_ForwardsLocaleQueryParam(t *testing.T) {
	var gotLocale string
	uc := &mockDocumentUC{}
	uc.getPublishedFn = func(_ context.Context, _, documentID, locale string) (*entity.Document, error) {
		gotLocale = locale
		return &entity.Document{DocumentID: documentID}, nil
	}
	h := handler.NewDocumentHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/api/public/content-types/articles/documents/abc?locale=vi", nil)
	req.SetPathValue("slug", "articles")
	req.SetPathValue("documentId", "abc")
	w := httptest.NewRecorder()
	h.GetPublic(w, req)

	if gotLocale != "vi" {
		t.Errorf("GetPublic() forwarded locale = %q, want vi", gotLocale)
	}
}

// ---- Update ----------------------------------------------------------------

func TestDocumentHandler_Update(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.saveFn = func(_ context.Context, _ string, doc *entity.Document, _ string) (*entity.Document, error) {
		return doc, nil
	}
	uc.getForEditFn = func(_ context.Context, _, documentID, _ string) (*entity.Document, string, error) {
		return &entity.Document{DocumentID: documentID}, "modified", nil
	}
	h := handler.NewDocumentHandler(uc)

	body := map[string]any{"data": map[string]any{"title": "Updated"}}
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(body)
	req := httptest.NewRequest(http.MethodPut, "/api/content-types/articles/documents/abc", &buf)
	req.SetPathValue("slug", "articles")
	req.SetPathValue("documentId", "abc")
	w := httptest.NewRecorder()
	h.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Update() status = %d, want 200", w.Code)
	}
}

func TestDocumentHandler_Update_ForwardsLocaleQueryParamToSavedDoc(t *testing.T) {
	var savedLocale string
	uc := &mockDocumentUC{}
	uc.saveFn = func(_ context.Context, _ string, doc *entity.Document, _ string) (*entity.Document, error) {
		savedLocale = doc.Locale
		return doc, nil
	}
	uc.getForEditFn = func(_ context.Context, _, documentID, _ string) (*entity.Document, string, error) {
		return &entity.Document{DocumentID: documentID}, "modified", nil
	}
	h := handler.NewDocumentHandler(uc)

	body := map[string]any{"data": map[string]any{"title": "Updated"}}
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(body)
	req := httptest.NewRequest(http.MethodPut, "/api/content-types/articles/documents/abc?locale=vi", &buf)
	req.SetPathValue("slug", "articles")
	req.SetPathValue("documentId", "abc")
	w := httptest.NewRecorder()
	h.Update(w, req)

	if savedLocale != "vi" {
		t.Errorf("Update() saved doc.Locale = %q, want vi", savedLocale)
	}
}

// ---- Delete ----------------------------------------------------------------

func TestDocumentHandler_Delete(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.deleteFn = func(_ context.Context, _, _ string) error { return nil }
	h := handler.NewDocumentHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/api/content-types/articles/documents/abc", nil)
	req.SetPathValue("slug", "articles")
	req.SetPathValue("documentId", "abc")
	w := httptest.NewRecorder()
	h.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Delete() status = %d, want 204", w.Code)
	}
}

// ---- Publish / Unpublish ---------------------------------------------------

func TestDocumentHandler_Publish(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.publishFn = func(_ context.Context, _, _, _, _ string) error { return nil }
	h := handler.NewDocumentHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/content-types/articles/documents/abc/publish", nil)
	req.SetPathValue("slug", "articles")
	req.SetPathValue("documentId", "abc")
	w := httptest.NewRecorder()
	h.Publish(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Publish() status = %d, want 200", w.Code)
	}
}

func TestDocumentHandler_Publish_ForwardsLocaleQueryParam(t *testing.T) {
	var gotLocale string
	uc := &mockDocumentUC{}
	uc.publishFn = func(_ context.Context, _, _, locale, _ string) error {
		gotLocale = locale
		return nil
	}
	h := handler.NewDocumentHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/content-types/articles/documents/abc/publish?locale=vi", nil)
	req.SetPathValue("slug", "articles")
	req.SetPathValue("documentId", "abc")
	w := httptest.NewRecorder()
	h.Publish(w, req)

	if gotLocale != "vi" {
		t.Errorf("Publish() forwarded locale = %q, want vi", gotLocale)
	}
}

func TestDocumentHandler_Unpublish(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.unpublishFn = func(_ context.Context, _, _, _ string) error { return nil }
	h := handler.NewDocumentHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/content-types/articles/documents/abc/unpublish", nil)
	req.SetPathValue("slug", "articles")
	req.SetPathValue("documentId", "abc")
	w := httptest.NewRecorder()
	h.Unpublish(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Unpublish() status = %d, want 200", w.Code)
	}
}

func TestDocumentHandler_Unpublish_ForwardsLocaleQueryParam(t *testing.T) {
	var gotLocale string
	uc := &mockDocumentUC{}
	uc.unpublishFn = func(_ context.Context, _, _, locale string) error {
		gotLocale = locale
		return nil
	}
	h := handler.NewDocumentHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/content-types/articles/documents/abc/unpublish?locale=vi", nil)
	req.SetPathValue("slug", "articles")
	req.SetPathValue("documentId", "abc")
	w := httptest.NewRecorder()
	h.Unpublish(w, req)

	if gotLocale != "vi" {
		t.Errorf("Unpublish() forwarded locale = %q, want vi", gotLocale)
	}
}
