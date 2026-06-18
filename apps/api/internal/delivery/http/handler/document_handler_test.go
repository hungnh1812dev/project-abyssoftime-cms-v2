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
	saveFn              func(ctx context.Context, contentTypeSlug string, doc *entity.Document, userID string) (*entity.Document, error)
	getForEditFn        func(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, string, error)
	getPublishedFn      func(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error)
	publishFn           func(ctx context.Context, contentTypeSlug, documentID, locale, userID string) error
	unpublishFn         func(ctx context.Context, contentTypeSlug, documentID, locale string) error
	deleteFn            func(ctx context.Context, contentTypeSlug, documentID string) error
	getSingleTypeFn     func(ctx context.Context, contentTypeSlug, locale string) (*entity.Document, string, error)
	saveSingleTypeFn    func(ctx context.Context, contentTypeSlug string, data map[string]any, locale, userID string) (*entity.Document, error)
	publishSingleTypeFn func(ctx context.Context, contentTypeSlug, locale, userID string) error
	unpublishSingleTypeFn func(ctx context.Context, contentTypeSlug, locale string) error
	getAllPaginatedFn    func(ctx context.Context, contentTypeSlug string, start, size int, locale string) ([]*entity.Document, []string, int64, error)
}

func (m *mockDocumentUC) Save(ctx context.Context, s string, d *entity.Document, u string) (*entity.Document, error) {
	return m.saveFn(ctx, s, d, u)
}
func (m *mockDocumentUC) GetForEdit(ctx context.Context, s, d, l string) (*entity.Document, string, error) {
	return m.getForEditFn(ctx, s, d, l)
}
func (m *mockDocumentUC) GetPublished(ctx context.Context, s, d, l string) (*entity.Document, error) {
	return m.getPublishedFn(ctx, s, d, l)
}
func (m *mockDocumentUC) Publish(ctx context.Context, s, d, l, u string) error {
	return m.publishFn(ctx, s, d, l, u)
}
func (m *mockDocumentUC) Unpublish(ctx context.Context, s, d, l string) error {
	return m.unpublishFn(ctx, s, d, l)
}
func (m *mockDocumentUC) Delete(ctx context.Context, s, d string) error {
	return m.deleteFn(ctx, s, d)
}
func (m *mockDocumentUC) GetSingleType(ctx context.Context, s, l string) (*entity.Document, string, error) {
	return m.getSingleTypeFn(ctx, s, l)
}
func (m *mockDocumentUC) SaveSingleType(ctx context.Context, s string, data map[string]any, l, u string) (*entity.Document, error) {
	return m.saveSingleTypeFn(ctx, s, data, l, u)
}
func (m *mockDocumentUC) PublishSingleType(ctx context.Context, s, l, u string) error {
	return m.publishSingleTypeFn(ctx, s, l, u)
}
func (m *mockDocumentUC) UnpublishSingleType(ctx context.Context, s, l string) error {
	return m.unpublishSingleTypeFn(ctx, s, l)
}
func (m *mockDocumentUC) GetAllPaginated(ctx context.Context, s string, start, size int, l string) ([]*entity.Document, []string, int64, error) {
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

// ---- Single-type: GetSingleType -------------------------------------------

func TestDocumentHandler_GetSingleType(t *testing.T) {
	t.Run("200 found", func(t *testing.T) {
		uc := &mockDocumentUC{}
		uc.getSingleTypeFn = func(_ context.Context, _, _ string) (*entity.Document, string, error) {
			return &entity.Document{DocumentID: "e1"}, "draft", nil
		}
		h := newHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/api/document-manager/single-type/about?locale=en", nil)
		req.SetPathValue("slug", "about")
		w := httptest.NewRecorder()
		h.GetSingleType(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("GetSingleType() status = %d, want 200", w.Code)
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		uc := &mockDocumentUC{}
		uc.getSingleTypeFn = func(_ context.Context, _, _ string) (*entity.Document, string, error) {
			return nil, "", pkgerrors.ErrNotFound
		}
		h := newHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/api/document-manager/single-type/about?locale=en", nil)
		req.SetPathValue("slug", "about")
		w := httptest.NewRecorder()
		h.GetSingleType(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("GetSingleType() status = %d, want 404", w.Code)
		}
	})
}

// ---- Single-type: SaveSingleType ------------------------------------------

func TestDocumentHandler_SaveSingleType(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.saveSingleTypeFn = func(_ context.Context, _ string, _ map[string]any, _ string, _ string) (*entity.Document, error) {
		return &entity.Document{DocumentID: "e1", Locale: "en"}, nil
	}
	uc.getSingleTypeFn = func(_ context.Context, _, _ string) (*entity.Document, string, error) {
		return &entity.Document{DocumentID: "e1"}, "draft", nil
	}
	h := newHandler(uc)

	body := map[string]any{"data": map[string]any{"title": "Hello"}}
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(body)
	req := httptest.NewRequest(http.MethodPut, "/api/document-manager/single-type/about", &buf)
	req.SetPathValue("slug", "about")
	w := httptest.NewRecorder()
	h.SaveSingleType(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("SaveSingleType() status = %d, want 200", w.Code)
	}
}

// ---- Single-type: Publish/Unpublish ----------------------------------------

func TestDocumentHandler_PublishSingleType(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.publishSingleTypeFn = func(_ context.Context, _, _, _ string) error { return nil }
	h := newHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/document-manager/single-type/about/publish", nil)
	req.SetPathValue("slug", "about")
	w := httptest.NewRecorder()
	h.PublishSingleType(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("PublishSingleType() status = %d, want 200", w.Code)
	}
}

func TestDocumentHandler_UnpublishSingleType(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.unpublishSingleTypeFn = func(_ context.Context, _, _ string) error { return nil }
	h := newHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/document-manager/single-type/about/unpublish", nil)
	req.SetPathValue("slug", "about")
	w := httptest.NewRecorder()
	h.UnpublishSingleType(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UnpublishSingleType() status = %d, want 200", w.Code)
	}
}

// ---- Collection-type: ListCollection ---------------------------------------

func TestDocumentHandler_ListCollection(t *testing.T) {
	t.Run("returns paginated response with projected data", func(t *testing.T) {
		uc := &mockDocumentUC{}
		uc.getAllPaginatedFn = func(_ context.Context, _ string, start, size int, _ string) ([]*entity.Document, []string, int64, error) {
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

		req := httptest.NewRequest(http.MethodGet, "/api/document-manager/collection-type/articles?start=0&size=20&locale=en", nil)
		req.SetPathValue("slug", "articles")
		w := httptest.NewRecorder()
		h.ListCollection(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("ListCollection() status = %d, want 200", w.Code)
		}
		var resp map[string]any
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp["total"].(float64) != 5 {
			t.Errorf("ListCollection() total = %v, want 5", resp["total"])
		}
		items := resp["items"].([]any)
		if len(items) != 1 {
			t.Fatalf("ListCollection() items count = %d, want 1", len(items))
		}
		item := items[0].(map[string]any)
		data := item["data"].(map[string]any)
		if _, ok := data["body"]; ok {
			t.Error("ListCollection() should not include 'body' in projected data")
		}
		if data["title"] != "Post 1" {
			t.Errorf("ListCollection() data.title = %v, want Post 1", data["title"])
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

		req := httptest.NewRequest(http.MethodGet, "/api/document-manager/collection-type/articles", nil)
		req.SetPathValue("slug", "articles")
		w := httptest.NewRecorder()
		h.ListCollection(w, req)

		if gotStart != 0 {
			t.Errorf("ListCollection() default start = %d, want 0", gotStart)
		}
		if gotSize != 20 {
			t.Errorf("ListCollection() default size = %d, want 20", gotSize)
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

		req := httptest.NewRequest(http.MethodGet, "/api/document-manager/collection-type/articles?size=500", nil)
		req.SetPathValue("slug", "articles")
		w := httptest.NewRecorder()
		h.ListCollection(w, req)

		if gotSize != 100 {
			t.Errorf("ListCollection() capped size = %d, want 100", gotSize)
		}
	})
}

// ---- Collection-type: GetCollection ----------------------------------------

func TestDocumentHandler_GetCollection(t *testing.T) {
	t.Run("200 found", func(t *testing.T) {
		uc := &mockDocumentUC{}
		uc.getForEditFn = func(_ context.Context, _, documentID, _ string) (*entity.Document, string, error) {
			return &entity.Document{DocumentID: documentID}, "draft", nil
		}
		h := newHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/api/document-manager/collection-type/articles/abc", nil)
		req.SetPathValue("slug", "articles")
		req.SetPathValue("documentId", "abc")
		w := httptest.NewRecorder()
		h.GetCollection(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetCollection() status = %d, want 200", w.Code)
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		uc := &mockDocumentUC{}
		uc.getForEditFn = func(_ context.Context, _, _, _ string) (*entity.Document, string, error) {
			return nil, "", pkgerrors.ErrNotFound
		}
		h := newHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/api/document-manager/collection-type/articles/missing", nil)
		req.SetPathValue("slug", "articles")
		req.SetPathValue("documentId", "missing")
		w := httptest.NewRecorder()
		h.GetCollection(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("GetCollection() status = %d, want 404", w.Code)
		}
	})
}

// ---- Collection-type: CreateCollection ------------------------------------

func TestDocumentHandler_CreateCollection(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.saveFn = func(_ context.Context, _ string, doc *entity.Document, _ string) (*entity.Document, error) {
		doc.DocumentID = "new-entry"
		return doc, nil
	}
	h := newHandler(uc)

	body := map[string]any{"data": map[string]any{"title": "Hello"}}
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(body)
	req := httptest.NewRequest(http.MethodPost, "/api/document-manager/collection-type/articles", &buf)
	req.SetPathValue("slug", "articles")
	w := httptest.NewRecorder()
	h.CreateCollection(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("CreateCollection() status = %d, want 201", w.Code)
	}
}

// ---- Collection-type: UpdateCollection ------------------------------------

func TestDocumentHandler_UpdateCollection(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.saveFn = func(_ context.Context, _ string, doc *entity.Document, _ string) (*entity.Document, error) {
		return doc, nil
	}
	uc.getForEditFn = func(_ context.Context, _, documentID, _ string) (*entity.Document, string, error) {
		return &entity.Document{DocumentID: documentID}, "modified", nil
	}
	h := newHandler(uc)

	body := map[string]any{"data": map[string]any{"title": "Updated"}}
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(body)
	req := httptest.NewRequest(http.MethodPut, "/api/document-manager/collection-type/articles/abc", &buf)
	req.SetPathValue("slug", "articles")
	req.SetPathValue("documentId", "abc")
	w := httptest.NewRecorder()
	h.UpdateCollection(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UpdateCollection() status = %d, want 200", w.Code)
	}
}

// ---- Collection-type: DeleteCollection ------------------------------------

func TestDocumentHandler_DeleteCollection(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.deleteFn = func(_ context.Context, _, _ string) error { return nil }
	h := newHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/api/document-manager/collection-type/articles/abc", nil)
	req.SetPathValue("slug", "articles")
	req.SetPathValue("documentId", "abc")
	w := httptest.NewRecorder()
	h.DeleteCollection(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("DeleteCollection() status = %d, want 204", w.Code)
	}
}

// ---- Collection-type: Publish/Unpublish -----------------------------------

func TestDocumentHandler_PublishCollection(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.publishFn = func(_ context.Context, _, _, _, _ string) error { return nil }
	h := newHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/document-manager/collection-type/articles/abc/publish", nil)
	req.SetPathValue("slug", "articles")
	req.SetPathValue("documentId", "abc")
	w := httptest.NewRecorder()
	h.PublishCollection(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("PublishCollection() status = %d, want 200", w.Code)
	}
}

func TestDocumentHandler_UnpublishCollection(t *testing.T) {
	uc := &mockDocumentUC{}
	uc.unpublishFn = func(_ context.Context, _, _, _ string) error { return nil }
	h := newHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/document-manager/collection-type/articles/abc/unpublish", nil)
	req.SetPathValue("slug", "articles")
	req.SetPathValue("documentId", "abc")
	w := httptest.NewRecorder()
	h.UnpublishCollection(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UnpublishCollection() status = %d, want 200", w.Code)
	}
}

// ---- Public (unchanged) ---------------------------------------------------

func TestDocumentHandler_GetPublic(t *testing.T) {
	t.Run("200 when published", func(t *testing.T) {
		uc := &mockDocumentUC{}
		uc.getPublishedFn = func(_ context.Context, _, documentID, _ string) (*entity.Document, error) {
			return &entity.Document{DocumentID: documentID}, nil
		}
		h := newHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/api/public/document-manager/articles/abc", nil)
		req.SetPathValue("slug", "articles")
		req.SetPathValue("documentId", "abc")
		w := httptest.NewRecorder()
		h.GetPublic(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetPublic() status = %d, want 200", w.Code)
		}
	})

	t.Run("404 when not published", func(t *testing.T) {
		uc := &mockDocumentUC{}
		uc.getPublishedFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
			return nil, pkgerrors.ErrNotFound
		}
		h := newHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/api/public/document-manager/articles/abc", nil)
		req.SetPathValue("slug", "articles")
		req.SetPathValue("documentId", "abc")
		w := httptest.NewRecorder()
		h.GetPublic(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("GetPublic() status = %d, want 404", w.Code)
		}
	})
}
