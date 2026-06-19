package dynamic

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	contenttype "project-abyssoftime-cms-v2/api/internal/usecase/content_type"
)

type mockDocUC struct {
	getForEditFn        func(ctx context.Context, slug, docID, locale string) (*entity.Document, string, error)
	getAllPaginatedFn    func(ctx context.Context, slug string, start, size int, locale string) ([]*entity.Document, []string, int64, error)
	getSingleTypeFn     func(ctx context.Context, slug, locale string) (*entity.Document, string, error)
	saveSingleTypeFn    func(ctx context.Context, slug string, data map[string]any, locale, userID string) (*entity.Document, error)
	publishSingleTypeFn func(ctx context.Context, slug, locale, userID string) error
	unpublishSingleTypeFn func(ctx context.Context, slug, locale string) error
	saveFn              func(ctx context.Context, slug string, doc *entity.Document, userID string) (*entity.Document, error)
	getPublishedFn      func(ctx context.Context, slug, docID, locale string) (*entity.Document, error)
	publishFn           func(ctx context.Context, slug, docID, locale, userID string) error
	unpublishFn         func(ctx context.Context, slug, docID, locale string) error
	deleteFn            func(ctx context.Context, slug, docID string) error
}

func (m *mockDocUC) Save(ctx context.Context, s string, d *entity.Document, u string) (*entity.Document, error) {
	return m.saveFn(ctx, s, d, u)
}
func (m *mockDocUC) GetForEdit(ctx context.Context, s, d, l string) (*entity.Document, string, error) {
	return m.getForEditFn(ctx, s, d, l)
}
func (m *mockDocUC) GetPublished(ctx context.Context, s, d, l string) (*entity.Document, error) {
	return m.getPublishedFn(ctx, s, d, l)
}
func (m *mockDocUC) Publish(ctx context.Context, s, d, l, u string) error {
	return m.publishFn(ctx, s, d, l, u)
}
func (m *mockDocUC) Unpublish(ctx context.Context, s, d, l string) error {
	return m.unpublishFn(ctx, s, d, l)
}
func (m *mockDocUC) Delete(ctx context.Context, s, d string) error {
	return m.deleteFn(ctx, s, d)
}
func (m *mockDocUC) GetSingleType(ctx context.Context, s, l string) (*entity.Document, string, error) {
	return m.getSingleTypeFn(ctx, s, l)
}
func (m *mockDocUC) SaveSingleType(ctx context.Context, s string, data map[string]any, l, u string) (*entity.Document, error) {
	return m.saveSingleTypeFn(ctx, s, data, l, u)
}
func (m *mockDocUC) PublishSingleType(ctx context.Context, s, l, u string) error {
	return m.publishSingleTypeFn(ctx, s, l, u)
}
func (m *mockDocUC) UnpublishSingleType(ctx context.Context, s, l string) error {
	return m.unpublishSingleTypeFn(ctx, s, l)
}
func (m *mockDocUC) GetAllPaginated(ctx context.Context, s string, start, size int, l string) ([]*entity.Document, []string, int64, error) {
	return m.getAllPaginatedFn(ctx, s, start, size, l)
}

type mockCTUC struct {
	findAllFn func(ctx context.Context) ([]*entity.ContentType, error)
}

func (m *mockCTUC) FindAll(ctx context.Context) ([]*entity.ContentType, error) {
	return m.findAllFn(ctx)
}

func gqlQuery(t *testing.T, h http.Handler, query string) map[string]any {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"query": query})
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GraphQL status = %d, body: %s", w.Code, w.Body.String())
	}
	var result map[string]any
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return result
}

func TestResolverFactory_ContentTypesQuery(t *testing.T) {
	ctUC := &mockCTUC{
		findAllFn: func(_ context.Context) ([]*entity.ContentType, error) {
			return []*entity.ContentType{
				{ID: "1", Name: "Blog", Slug: "blog", Kind: entity.KindCollection},
			}, nil
		},
	}

	factory := NewResolverFactory(&mockDocUC{}, ctUC)
	h, err := factory.BuildHandler(nil)
	if err != nil {
		t.Fatalf("BuildHandler: %v", err)
	}

	result := gqlQuery(t, h, `{ contentTypes { id name slug kind } }`)
	data, ok := result["data"].(map[string]any)
	if !ok {
		t.Fatalf("no data in response: %v", result)
	}
	cts, ok := data["contentTypes"].([]any)
	if !ok || len(cts) != 1 {
		t.Fatalf("contentTypes = %v", data["contentTypes"])
	}
	ct := cts[0].(map[string]any)
	if ct["slug"] != "blog" {
		t.Errorf("slug = %v, want blog", ct["slug"])
	}
}

func TestResolverFactory_SingleTypeQuery(t *testing.T) {
	docUC := &mockDocUC{
		getSingleTypeFn: func(_ context.Context, slug, locale string) (*entity.Document, string, error) {
			return &entity.Document{
				DocumentID: "d1",
				Locale:     "en",
				Data:       map[string]any{"headline": "Hello"},
			}, "draft", nil
		},
	}

	defs := []contenttype.ContentTypeDefinition{
		{Slug: "about-page", Kind: "single", Fields: []entity.FieldDefinition{{Name: "headline", Type: "text"}}},
	}

	factory := NewResolverFactory(docUC, &mockCTUC{findAllFn: func(_ context.Context) ([]*entity.ContentType, error) { return nil, nil }})
	h, err := factory.BuildHandler(defs)
	if err != nil {
		t.Fatalf("BuildHandler: %v", err)
	}

	result := gqlQuery(t, h, `{ aboutPage(locale: "en") { documentId headline status } }`)
	data := result["data"].(map[string]any)
	ap := data["aboutPage"].(map[string]any)
	if ap["documentId"] != "d1" {
		t.Errorf("documentId = %v, want d1", ap["documentId"])
	}
	if ap["headline"] != "Hello" {
		t.Errorf("headline = %v, want Hello", ap["headline"])
	}
	if ap["status"] != "draft" {
		t.Errorf("status = %v, want draft", ap["status"])
	}
}

func TestResolverFactory_CollectionListQuery(t *testing.T) {
	docUC := &mockDocUC{
		getAllPaginatedFn: func(_ context.Context, _ string, start, size int, _ string) ([]*entity.Document, []string, int64, error) {
			return []*entity.Document{
				{DocumentID: "d1", Locale: "en", Data: map[string]any{"title": "Post 1"}},
			}, []string{"draft"}, 5, nil
		},
	}

	defs := []contenttype.ContentTypeDefinition{
		{Slug: "blog-posts", Kind: "collection", Fields: []entity.FieldDefinition{{Name: "title", Type: "text"}}},
	}

	factory := NewResolverFactory(docUC, &mockCTUC{findAllFn: func(_ context.Context) ([]*entity.ContentType, error) { return nil, nil }})
	h, err := factory.BuildHandler(defs)
	if err != nil {
		t.Fatalf("BuildHandler: %v", err)
	}

	result := gqlQuery(t, h, `{ blogPostsList(start: 0, size: 10, locale: "en") { items { documentId title status } total } }`)
	data := result["data"].(map[string]any)
	list := data["blogPostsList"].(map[string]any)
	if list["total"].(float64) != 5 {
		t.Errorf("total = %v, want 5", list["total"])
	}
	items := list["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("items count = %d, want 1", len(items))
	}
	item := items[0].(map[string]any)
	if item["title"] != "Post 1" {
		t.Errorf("title = %v, want Post 1", item["title"])
	}
}
