package resolver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	graphqlpkg "project-abyssoftime-cms-v2/api/graphql"
	"project-abyssoftime-cms-v2/api/graphql/resolver"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type mockDocUC struct {
	getForEditFn             func(ctx context.Context, slug, docID, locale string) (*entity.Document, string, error)
	getAllPaginatedFn         func(ctx context.Context, slug string, start, size int, locale string, orderBy string, sortDir int, filters []entity.FilterNode) ([]*entity.Document, []string, int64, error)
	getPublishedPaginatedFn  func(ctx context.Context, slug string, start, size int, locale string, filters []entity.FilterNode) ([]*entity.Document, int64, error)
	getPublishedSingleTypeFn func(ctx context.Context, slug, locale string) (*entity.Document, error)
	getSingleTypeFn          func(ctx context.Context, slug, locale string) (*entity.Document, string, error)
	saveSingleTypeFn         func(ctx context.Context, slug string, data map[string]any, locale, userID string) (*entity.Document, error)
	publishSingleTypeFn      func(ctx context.Context, slug, locale, userID string) error
	unpublishSingleTypeFn    func(ctx context.Context, slug, locale string) error
	saveFn                   func(ctx context.Context, slug string, doc *entity.Document, userID string) (*entity.Document, error)
	getPublishedFn           func(ctx context.Context, slug, docID, locale string) (*entity.Document, error)
	publishFn                func(ctx context.Context, slug, docID, locale, userID string) error
	unpublishFn              func(ctx context.Context, slug, docID, locale string) error
	deleteFn                 func(ctx context.Context, slug, docID string) error
}

func (mock *mockDocUC) Save(ctx context.Context, slug string, doc *entity.Document, _ []entity.FieldDefinition, userID string) (*entity.Document, error) {
	return mock.saveFn(ctx, slug, doc, userID)
}
func (mock *mockDocUC) GetForEdit(ctx context.Context, slug, docID, locale string, _ []entity.FieldDefinition) (*entity.Document, string, error) {
	return mock.getForEditFn(ctx, slug, docID, locale)
}
func (mock *mockDocUC) GetPublished(ctx context.Context, slug, docID, locale string, _ []entity.FieldDefinition) (*entity.Document, error) {
	return mock.getPublishedFn(ctx, slug, docID, locale)
}
func (mock *mockDocUC) Publish(ctx context.Context, slug, docID, locale string, _ []entity.FieldDefinition, userID string) error {
	return mock.publishFn(ctx, slug, docID, locale, userID)
}
func (mock *mockDocUC) Unpublish(ctx context.Context, slug, docID, locale string, _ []entity.FieldDefinition) error {
	return mock.unpublishFn(ctx, slug, docID, locale)
}
func (mock *mockDocUC) Delete(ctx context.Context, slug, docID string, _ []entity.FieldDefinition) error {
	return mock.deleteFn(ctx, slug, docID)
}
func (mock *mockDocUC) GetSingleType(ctx context.Context, slug, locale string, _ []entity.FieldDefinition) (*entity.Document, string, error) {
	return mock.getSingleTypeFn(ctx, slug, locale)
}
func (mock *mockDocUC) SaveSingleType(ctx context.Context, slug string, data map[string]any, locale string, _ []entity.FieldDefinition, userID string) (*entity.Document, error) {
	return mock.saveSingleTypeFn(ctx, slug, data, locale, userID)
}
func (mock *mockDocUC) PublishSingleType(ctx context.Context, slug, locale string, _ []entity.FieldDefinition, userID string) error {
	return mock.publishSingleTypeFn(ctx, slug, locale, userID)
}
func (mock *mockDocUC) UnpublishSingleType(ctx context.Context, slug, locale string, _ []entity.FieldDefinition) error {
	return mock.unpublishSingleTypeFn(ctx, slug, locale)
}
func (mock *mockDocUC) GetAllPaginated(ctx context.Context, slug string, start, size int, locale string, _ []entity.FieldDefinition, orderBy string, sortDir int, filters []entity.FilterNode) ([]*entity.Document, []string, int64, error) {
	return mock.getAllPaginatedFn(ctx, slug, start, size, locale, orderBy, sortDir, filters)
}
func (mock *mockDocUC) GetPublishedPaginated(ctx context.Context, slug string, start, size int, locale string, _ []entity.FieldDefinition, filters []entity.FilterNode) ([]*entity.Document, int64, error) {
	if mock.getPublishedPaginatedFn != nil {
		return mock.getPublishedPaginatedFn(ctx, slug, start, size, locale, filters)
	}
	return nil, 0, nil
}
func (mock *mockDocUC) GetPublishedSingleType(ctx context.Context, slug, locale string, _ []entity.FieldDefinition) (*entity.Document, error) {
	if mock.getPublishedSingleTypeFn != nil {
		return mock.getPublishedSingleTypeFn(ctx, slug, locale)
	}
	return nil, nil
}

type mockCTUC struct {
	findAllFn func(ctx context.Context) ([]*entity.ContentType, error)
}

func (mock *mockCTUC) FindAll(ctx context.Context) ([]*entity.ContentType, error) {
	return mock.findAllFn(ctx)
}

type mockTokenValidator struct{}

func (mock *mockTokenValidator) Validate(_ context.Context, _ string) (*entity.AccessToken, error) {
	return &entity.AccessToken{}, nil
}

func gqlQuery(test *testing.T, handler http.Handler, query string) map[string]any {
	test.Helper()
	body, _ := json.Marshal(map[string]string{"query": query})
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		test.Fatalf("GraphQL status = %d, body: %s", recorder.Code, recorder.Body.String())
	}
	var result map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&result); err != nil {
		test.Fatalf("decode response: %v", err)
	}
	return result
}

func buildHandler(test *testing.T, docUC resolver.DocumentUseCase, ctUC resolver.ContentTypeUseCase) http.Handler {
	test.Helper()
	gqlResolver := resolver.NewResolver(docUC, ctUC, nil)
	return graphqlpkg.NewHandler(gqlResolver, &mockTokenValidator{})
}

func TestResolver_ContentTypesQuery(test *testing.T) {
	ctUC := &mockCTUC{
		findAllFn: func(_ context.Context) ([]*entity.ContentType, error) {
			return []*entity.ContentType{
				{ID: 1, Name: "Blog", Slug: "blog", Kind: entity.KindCollection},
			}, nil
		},
	}

	handler := buildHandler(test, &mockDocUC{}, ctUC)
	result := gqlQuery(test, handler, `{ contentTypes { id name slug kind } }`)

	data, hasData := result["data"].(map[string]any)
	if !hasData {
		test.Fatalf("no data in response: %v", result)
	}
	contentTypes, hasCTs := data["contentTypes"].([]any)
	if !hasCTs || len(contentTypes) != 1 {
		test.Fatalf("contentTypes = %v", data["contentTypes"])
	}
	contentType := contentTypes[0].(map[string]any)
	if contentType["slug"] != "blog" {
		test.Errorf("slug = %v, want blog", contentType["slug"])
	}
}

func TestResolver_SingleTypeQuery(test *testing.T) {
	docUC := &mockDocUC{
		getPublishedSingleTypeFn: func(_ context.Context, slug, locale string) (*entity.Document, error) {
			return &entity.Document{
				DocumentID: "d1",
				Locale:     "en",
				Fields:     map[string]any{"headline": "Hello"},
			}, nil
		},
	}

	emptyCTUC := &mockCTUC{findAllFn: func(_ context.Context) ([]*entity.ContentType, error) { return nil, nil }}
	handler := buildHandler(test, docUC, emptyCTUC)

	result := gqlQuery(test, handler, `{ commonText(locale: "en") { documentId } }`)
	data := result["data"].(map[string]any)
	commonText := data["commonText"].(map[string]any)
	if commonText["documentId"] != "d1" {
		test.Errorf("documentId = %v, want d1", commonText["documentId"])
	}
}

func TestResolver_CollectionListQuery(test *testing.T) {
	docUC := &mockDocUC{
		getPublishedPaginatedFn: func(_ context.Context, _ string, start, size int, _ string, _ []entity.FilterNode) ([]*entity.Document, int64, error) {
			return []*entity.Document{
				{DocumentID: "d1", Locale: "en", Fields: map[string]any{"position": "Engineer"}},
			}, 5, nil
		},
	}

	emptyCTUC := &mockCTUC{findAllFn: func(_ context.Context) ([]*entity.ContentType, error) { return nil, nil }}
	handler := buildHandler(test, docUC, emptyCTUC)

	result := gqlQuery(test, handler, `{ cvPageList(start: 0, size: 10, locale: "en") { documentId position } }`)
	data := result["data"].(map[string]any)
	items := data["cvPageList"].([]any)
	if len(items) != 1 {
		test.Fatalf("cvPageList count = %d, want 1", len(items))
	}
	item := items[0].(map[string]any)
	if item["position"] != "Engineer" {
		test.Errorf("position = %v, want Engineer", item["position"])
	}
}

func TestResolver_UnauthorizedRejected(test *testing.T) {
	handler := buildHandler(test, &mockDocUC{}, &mockCTUC{findAllFn: func(_ context.Context) ([]*entity.ContentType, error) { return nil, nil }})

	body, _ := json.Marshal(map[string]string{"query": `{ contentTypes { id } }`})
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		test.Errorf("expected 401, got %d", recorder.Code)
	}
}
