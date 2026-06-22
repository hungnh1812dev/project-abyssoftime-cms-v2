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

type mockContentTypeUC struct {
	findByIDFn   func(ctx context.Context, id string) (*entity.ContentType, error)
	findBySlugFn func(ctx context.Context, slug string) (*entity.ContentType, error)
	findAllFn    func(ctx context.Context) ([]*entity.ContentType, error)
	updateFn     func(ctx context.Context, ct *entity.ContentType) error
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
func (m *mockContentTypeUC) Update(ctx context.Context, ct *entity.ContentType) error {
	return m.updateFn(ctx, ct)
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

func TestContentTypeHandler_UpdateListFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	blogCT := &entity.ContentType{
		DocumentID: "ct-1",
		Slug:       "blog-posts",
		Kind:       entity.KindCollection,
		Fields: []entity.FieldDefinition{
			{Name: "title", Type: "text"},
			{Name: "slug", Type: "text"},
			{Name: "body", Type: "richtext"},
			{Name: "banner", Type: "component"},
		},
	}

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantFields []string
	}{
		{
			name:       "200 valid content fields",
			body:       `{"listFields":["title","slug"]}`,
			wantStatus: http.StatusOK,
			wantFields: []string{"title", "slug"},
		},
		{
			name:       "200 valid system fields",
			body:       `{"listFields":["title","createdAt","updatedByName"]}`,
			wantStatus: http.StatusOK,
			wantFields: []string{"title", "createdAt", "updatedByName"},
		},
		{
			name:       "200 empty array resets to defaults",
			body:       `{"listFields":[]}`,
			wantStatus: http.StatusOK,
			wantFields: []string{},
		},
		{
			name:       "400 invalid field name",
			body:       `{"listFields":["title","nonexistent"]}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 component field rejected",
			body:       `{"listFields":["title","banner"]}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockContentTypeUC{}
			uc.findBySlugFn = func(_ context.Context, _ string) (*entity.ContentType, error) {
				ctCopy := *blogCT
				return &ctCopy, nil
			}
			var savedListFields []string
			uc.updateFn = func(_ context.Context, ct *entity.ContentType) error {
				savedListFields = ct.ListFields
				return nil
			}
			handler := handler.NewContentTypeHandler(uc)

			recorder := httptest.NewRecorder()
			_, router := gin.CreateTestContext(recorder)
			router.PATCH("/api/content-types/:slug/list-fields", handler.UpdateListFields)

			req := httptest.NewRequest(http.MethodPatch, "/api/content-types/blog-posts/list-fields", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Errorf("UpdateListFields() status = %d, want %d, body = %s", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var out map[string]any
				if err := json.NewDecoder(recorder.Body).Decode(&out); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				rawFields, ok := out["listFields"].([]any)
				if !ok {
					t.Fatalf("response missing listFields array")
				}
				if len(rawFields) != len(tt.wantFields) {
					t.Errorf("response listFields len = %d, want %d", len(rawFields), len(tt.wantFields))
				}
				if len(savedListFields) != len(tt.wantFields) {
					t.Errorf("saved ListFields len = %d, want %d", len(savedListFields), len(tt.wantFields))
				}
			}
		})
	}
}

func TestContentTypeHandler_UpdateListFields_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockContentTypeUC{}
	uc.findBySlugFn = func(_ context.Context, _ string) (*entity.ContentType, error) {
		return nil, pkgerrors.ErrNotFound
	}
	handler := handler.NewContentTypeHandler(uc)

	recorder := httptest.NewRecorder()
	_, router := gin.CreateTestContext(recorder)
	router.PATCH("/api/content-types/:slug/list-fields", handler.UpdateListFields)

	req := httptest.NewRequest(http.MethodPatch, "/api/content-types/missing/list-fields", bytes.NewBufferString(`{"listFields":["title"]}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("UpdateListFields() status = %d, want 404", recorder.Code)
	}
}
