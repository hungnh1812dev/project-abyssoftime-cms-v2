package handler_test

import (
	"bytes"
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

type mockDocumentUC struct {
	saveFn                func(ctx context.Context, contentTypeSlug string, doc *entity.Document, userID string) (*entity.Document, error)
	getForEditFn          func(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, string, error)
	getPublishedFn        func(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error)
	publishFn             func(ctx context.Context, contentTypeSlug, documentID, locale, userID string) error
	unpublishFn           func(ctx context.Context, contentTypeSlug, documentID, locale string) error
	deleteFn              func(ctx context.Context, contentTypeSlug, documentID string) error
	getSingleTypeFn       func(ctx context.Context, contentTypeSlug, locale string) (*entity.Document, string, error)
	saveSingleTypeFn      func(ctx context.Context, contentTypeSlug string, data map[string]any, locale, userID string) (*entity.Document, error)
	publishSingleTypeFn   func(ctx context.Context, contentTypeSlug, locale, userID string) error
	unpublishSingleTypeFn func(ctx context.Context, contentTypeSlug, locale string) error
	getAllPaginatedFn      func(ctx context.Context, contentTypeSlug string, start, size int, locale string) ([]*entity.Document, []string, int64, error)
}

func (m *mockDocumentUC) Save(ctx context.Context, s string, d *entity.Document, _ []entity.FieldDefinition, u string) (*entity.Document, error) {
	return m.saveFn(ctx, s, d, u)
}
func (m *mockDocumentUC) GetForEdit(ctx context.Context, s, d, l string, _ []entity.FieldDefinition) (*entity.Document, string, error) {
	return m.getForEditFn(ctx, s, d, l)
}
func (m *mockDocumentUC) GetPublished(ctx context.Context, s, d, l string, _ []entity.FieldDefinition) (*entity.Document, error) {
	return m.getPublishedFn(ctx, s, d, l)
}
func (m *mockDocumentUC) Publish(ctx context.Context, s, d, l string, _ []entity.FieldDefinition, u string) error {
	return m.publishFn(ctx, s, d, l, u)
}
func (m *mockDocumentUC) Unpublish(ctx context.Context, s, d, l string) error {
	return m.unpublishFn(ctx, s, d, l)
}
func (m *mockDocumentUC) Delete(ctx context.Context, s, d string, _ []entity.FieldDefinition) error {
	return m.deleteFn(ctx, s, d)
}
func (m *mockDocumentUC) GetSingleType(ctx context.Context, s, l string, _ []entity.FieldDefinition) (*entity.Document, string, error) {
	return m.getSingleTypeFn(ctx, s, l)
}
func (m *mockDocumentUC) SaveSingleType(ctx context.Context, s string, data map[string]any, l string, _ []entity.FieldDefinition, u string) (*entity.Document, error) {
	return m.saveSingleTypeFn(ctx, s, data, l, u)
}
func (m *mockDocumentUC) PublishSingleType(ctx context.Context, s, l string, _ []entity.FieldDefinition, u string) error {
	return m.publishSingleTypeFn(ctx, s, l, u)
}
func (m *mockDocumentUC) UnpublishSingleType(ctx context.Context, s, l string) error {
	return m.unpublishSingleTypeFn(ctx, s, l)
}
func (m *mockDocumentUC) GetAllPaginated(ctx context.Context, s string, start, size int, l string, _ []entity.FieldDefinition) ([]*entity.Document, []string, int64, error) {
	return m.getAllPaginatedFn(ctx, s, start, size, l)
}

type mockCTUC struct {
	findBySlugFn func(ctx context.Context, slug string) (*entity.ContentType, error)
}

func (m *mockCTUC) FindBySlug(ctx context.Context, slug string) (*entity.ContentType, error) {
	return m.findBySlugFn(ctx, slug)
}

func newHandler(uc *mockDocumentUC) *handler.DocumentHandler {
	ctUC := &mockCTUC{findBySlugFn: func(_ context.Context, _ string) (*entity.ContentType, error) {
		return &entity.ContentType{ListFields: []string{"title"}}, nil
	}}
	return handler.NewDocumentHandler(uc, ctUC)
}

func TestDocumentHandler_GetSingleType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 found", func(t *testing.T) {
		uc := &mockDocumentUC{}
		uc.getSingleTypeFn = func(_ context.Context, _, _ string) (*entity.Document, string, error) {
			return &entity.Document{DocumentID: "e1"}, "draft", nil
		}
		h := newHandler(uc)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.GET("/api/document-manager/single-type/:slug", h.GetSingleType)
		req := httptest.NewRequest(http.MethodGet, "/api/document-manager/single-type/about?locale=en", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		uc := &mockDocumentUC{}
		uc.getSingleTypeFn = func(_ context.Context, _, _ string) (*entity.Document, string, error) {
			return nil, "", pkgerrors.ErrNotFound
		}
		h := newHandler(uc)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.GET("/api/document-manager/single-type/:slug", h.GetSingleType)
		req := httptest.NewRequest(http.MethodGet, "/api/document-manager/single-type/about?locale=en", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", w.Code)
		}
	})
}

func TestDocumentHandler_SaveSingleType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockDocumentUC{}
	uc.saveSingleTypeFn = func(_ context.Context, _ string, _ map[string]any, _ string, _ string) (*entity.Document, error) {
		return &entity.Document{DocumentID: "e1", Locale: "en"}, nil
	}
	uc.getSingleTypeFn = func(_ context.Context, _, _ string) (*entity.Document, string, error) {
		return &entity.Document{DocumentID: "e1"}, "draft", nil
	}
	h := newHandler(uc)

	body, _ := json.Marshal(map[string]any{"data": map[string]any{"title": "Hello"}})
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.PUT("/api/document-manager/single-type/:slug", h.SaveSingleType)
	req := httptest.NewRequest(http.MethodPut, "/api/document-manager/single-type/about", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestDocumentHandler_PublishSingleType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockDocumentUC{}
	uc.publishSingleTypeFn = func(_ context.Context, _, _, _ string) error { return nil }
	h := newHandler(uc)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/api/document-manager/single-type/:slug/publish", h.PublishSingleType)
	req := httptest.NewRequest(http.MethodPost, "/api/document-manager/single-type/about/publish", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestDocumentHandler_UnpublishSingleType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockDocumentUC{}
	uc.unpublishSingleTypeFn = func(_ context.Context, _, _ string) error { return nil }
	h := newHandler(uc)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/api/document-manager/single-type/:slug/unpublish", h.UnpublishSingleType)
	req := httptest.NewRequest(http.MethodPost, "/api/document-manager/single-type/about/unpublish", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestDocumentHandler_ListCollection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns paginated response with projected data", func(t *testing.T) {
		uc := &mockDocumentUC{}
		uc.getAllPaginatedFn = func(_ context.Context, _ string, _, _ int, _ string) ([]*entity.Document, []string, int64, error) {
			return []*entity.Document{
				{DocumentID: "d1", Data: map[string]any{"title": "Post 1", "body": "full body"}, Locale: "en"},
			}, []string{"draft"}, 5, nil
		}
		ctUC := &mockCTUC{findBySlugFn: func(_ context.Context, _ string) (*entity.ContentType, error) {
			return &entity.ContentType{
				ListFields: []string{"title"},
				Fields:     []entity.FieldDefinition{{Name: "title", Type: "text"}, {Name: "body", Type: "richtext"}},
			}, nil
		}}
		h := handler.NewDocumentHandler(uc, ctUC)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.GET("/api/document-manager/collection-type/:slug", h.ListCollection)
		req := httptest.NewRequest(http.MethodGet, "/api/document-manager/collection-type/articles?start=0&size=20&locale=en", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		var resp map[string]any
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp["total"].(float64) != 5 {
			t.Errorf("total = %v, want 5", resp["total"])
		}
		items := resp["items"].([]any)
		if len(items) != 1 {
			t.Fatalf("items count = %d, want 1", len(items))
		}
		item := items[0].(map[string]any)
		data := item["data"].(map[string]any)
		if _, ok := data["body"]; ok {
			t.Error("should not include 'body' in projected data")
		}
		if data["title"] != "Post 1" {
			t.Errorf("data.title = %v, want Post 1", data["title"])
		}
	})

	t.Run("defaults pagination params", func(t *testing.T) {
		var gotStart, gotSize int
		uc := &mockDocumentUC{}
		uc.getAllPaginatedFn = func(_ context.Context, _ string, start, size int, _ string) ([]*entity.Document, []string, int64, error) {
			gotStart = start
			gotSize = size
			return nil, nil, 0, nil
		}
		h := newHandler(uc)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.GET("/api/document-manager/collection-type/:slug", h.ListCollection)
		req := httptest.NewRequest(http.MethodGet, "/api/document-manager/collection-type/articles", nil)
		r.ServeHTTP(w, req)

		if gotStart != 0 {
			t.Errorf("default start = %d, want 0", gotStart)
		}
		if gotSize != 20 {
			t.Errorf("default size = %d, want 20", gotSize)
		}
	})

	t.Run("caps size at 100", func(t *testing.T) {
		var gotSize int
		uc := &mockDocumentUC{}
		uc.getAllPaginatedFn = func(_ context.Context, _ string, _, size int, _ string) ([]*entity.Document, []string, int64, error) {
			gotSize = size
			return nil, nil, 0, nil
		}
		h := newHandler(uc)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.GET("/api/document-manager/collection-type/:slug", h.ListCollection)
		req := httptest.NewRequest(http.MethodGet, "/api/document-manager/collection-type/articles?size=500", nil)
		r.ServeHTTP(w, req)

		if gotSize != 100 {
			t.Errorf("capped size = %d, want 100", gotSize)
		}
	})
}

func TestDocumentHandler_GetCollection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 found", func(t *testing.T) {
		uc := &mockDocumentUC{}
		uc.getForEditFn = func(_ context.Context, _, documentID, _ string) (*entity.Document, string, error) {
			return &entity.Document{DocumentID: documentID}, "draft", nil
		}
		h := newHandler(uc)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.GET("/api/document-manager/collection-type/:slug/:documentId", h.GetCollection)
		req := httptest.NewRequest(http.MethodGet, "/api/document-manager/collection-type/articles/abc", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", w.Code)
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		uc := &mockDocumentUC{}
		uc.getForEditFn = func(_ context.Context, _, _, _ string) (*entity.Document, string, error) {
			return nil, "", pkgerrors.ErrNotFound
		}
		h := newHandler(uc)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.GET("/api/document-manager/collection-type/:slug/:documentId", h.GetCollection)
		req := httptest.NewRequest(http.MethodGet, "/api/document-manager/collection-type/articles/missing", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", w.Code)
		}
	})
}

func TestDocumentHandler_CreateCollection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockDocumentUC{}
	uc.saveFn = func(_ context.Context, _ string, doc *entity.Document, _ string) (*entity.Document, error) {
		doc.DocumentID = "new-entry"
		return doc, nil
	}
	h := newHandler(uc)

	body, _ := json.Marshal(map[string]any{"data": map[string]any{"title": "Hello"}})
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/api/document-manager/collection-type/:slug", h.CreateCollection)
	req := httptest.NewRequest(http.MethodPost, "/api/document-manager/collection-type/articles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
}

func TestDocumentHandler_UpdateCollection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockDocumentUC{}
	uc.saveFn = func(_ context.Context, _ string, doc *entity.Document, _ string) (*entity.Document, error) {
		return doc, nil
	}
	uc.getForEditFn = func(_ context.Context, _, documentID, _ string) (*entity.Document, string, error) {
		return &entity.Document{DocumentID: documentID}, "modified", nil
	}
	h := newHandler(uc)

	body, _ := json.Marshal(map[string]any{"data": map[string]any{"title": "Updated"}})
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.PUT("/api/document-manager/collection-type/:slug/:documentId", h.UpdateCollection)
	req := httptest.NewRequest(http.MethodPut, "/api/document-manager/collection-type/articles/abc", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestDocumentHandler_DeleteCollection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockDocumentUC{}
	uc.deleteFn = func(_ context.Context, _, _ string) error { return nil }
	h := newHandler(uc)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.DELETE("/api/document-manager/collection-type/:slug/:documentId", h.DeleteCollection)
	req := httptest.NewRequest(http.MethodDelete, "/api/document-manager/collection-type/articles/abc", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

func TestDocumentHandler_PublishCollection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockDocumentUC{}
	uc.publishFn = func(_ context.Context, _, _, _, _ string) error { return nil }
	h := newHandler(uc)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/api/document-manager/collection-type/:slug/:documentId/publish", h.PublishCollection)
	req := httptest.NewRequest(http.MethodPost, "/api/document-manager/collection-type/articles/abc/publish", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestDocumentHandler_UnpublishCollection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockDocumentUC{}
	uc.unpublishFn = func(_ context.Context, _, _, _ string) error { return nil }
	h := newHandler(uc)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/api/document-manager/collection-type/:slug/:documentId/unpublish", h.UnpublishCollection)
	req := httptest.NewRequest(http.MethodPost, "/api/document-manager/collection-type/articles/abc/unpublish", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestDocumentHandler_GetPublic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 when published", func(t *testing.T) {
		uc := &mockDocumentUC{}
		uc.getPublishedFn = func(_ context.Context, _, documentID, _ string) (*entity.Document, error) {
			return &entity.Document{DocumentID: documentID}, nil
		}
		h := newHandler(uc)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.GET("/api/public/document-manager/:slug/:documentId", h.GetPublic)
		req := httptest.NewRequest(http.MethodGet, "/api/public/document-manager/articles/abc", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", w.Code)
		}
	})

	t.Run("404 when not published", func(t *testing.T) {
		uc := &mockDocumentUC{}
		uc.getPublishedFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
			return nil, pkgerrors.ErrNotFound
		}
		h := newHandler(uc)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.GET("/api/public/document-manager/:slug/:documentId", h.GetPublic)
		req := httptest.NewRequest(http.MethodGet, "/api/public/document-manager/articles/abc", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", w.Code)
		}
	})
}
