package resolver_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"project-abyssoftime-cms-v2/api/graphql/resolver"
	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"
)

// ---- mock usecases ---------------------------------------------------------

type mockDocumentUC struct {
	saveFn         func(ctx context.Context, contentTypeSlug string, doc *entity.Document, userID string) (*entity.Document, error)
	getPublishedFn func(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error)
	getForEditFn   func(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, string, error)
	publishFn      func(ctx context.Context, contentTypeSlug, documentID, locale, userID string) error
	unpublishFn    func(ctx context.Context, contentTypeSlug, documentID, locale string) error
	deleteFn       func(ctx context.Context, contentTypeSlug, documentID string) error
}

func (m *mockDocumentUC) Save(ctx context.Context, contentTypeSlug string, doc *entity.Document, userID string) (*entity.Document, error) {
	return m.saveFn(ctx, contentTypeSlug, doc, userID)
}
func (m *mockDocumentUC) GetPublished(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error) {
	return m.getPublishedFn(ctx, contentTypeSlug, documentID, locale)
}
func (m *mockDocumentUC) GetForEdit(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, string, error) {
	return m.getForEditFn(ctx, contentTypeSlug, documentID, locale)
}
func (m *mockDocumentUC) Publish(ctx context.Context, contentTypeSlug, documentID, locale, userID string) error {
	return m.publishFn(ctx, contentTypeSlug, documentID, locale, userID)
}
func (m *mockDocumentUC) Unpublish(ctx context.Context, contentTypeSlug, documentID, locale string) error {
	return m.unpublishFn(ctx, contentTypeSlug, documentID, locale)
}
func (m *mockDocumentUC) Delete(ctx context.Context, contentTypeSlug, documentID string) error {
	return m.deleteFn(ctx, contentTypeSlug, documentID)
}

type mockContentTypeUC struct {
	findAllFn func(ctx context.Context) ([]*entity.ContentType, error)
}

func (m *mockContentTypeUC) FindAll(ctx context.Context) ([]*entity.ContentType, error) {
	return m.findAllFn(ctx)
}

// ---- helpers ---------------------------------------------------------------

func ctxWithRequest(r *http.Request) context.Context {
	return resolver.WithRequest(context.Background(), r)
}

func authedCtx() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.ContextKeyUserID, "user-123")
	ctx = context.WithValue(ctx, middleware.ContextKeyRole, "admin")
	return ctx
}

func sampleDoc(documentID string) *entity.Document {
	now := time.Now().UTC()
	return &entity.Document{
		DocumentID: documentID, Version: entity.VersionDraft,
		ContentTypeID: "ct-1", Data: map[string]any{"title": "Hello"},
		Locale: "en", CreatedAt: now, UpdatedAt: now,
		CreatedBy: "user-1", UpdatedBy: "user-1",
	}
}

// ---- AuthDirective tests ---------------------------------------------------

func TestAuthDirective_NoHeader_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	ctx := ctxWithRequest(req)

	_, err := resolver.AuthDirective(ctx, nil, func(ctx context.Context) (any, error) {
		return "ok", nil
	})
	if err == nil {
		t.Error("expected error for missing Authorization header, got nil")
	}
}

func TestAuthDirective_InvalidToken_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	ctx := ctxWithRequest(req)

	_, err := resolver.AuthDirective(ctx, nil, func(ctx context.Context) (any, error) {
		return "ok", nil
	})
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}
}

func TestAuthDirective_ValidToken_CallsNextWithUserInContext(t *testing.T) {
	token, err := pkgjwt.GenerateAccessToken("user-123", "admin")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	ctx := ctxWithRequest(req)

	var capturedCtx context.Context
	_, err = resolver.AuthDirective(ctx, nil, func(ctx context.Context) (any, error) {
		capturedCtx = ctx
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uid := middleware.UserID(capturedCtx); uid != "user-123" {
		t.Errorf("userID = %q, want user-123", uid)
	}
	if role := middleware.Role(capturedCtx); role != "admin" {
		t.Errorf("role = %q, want admin", role)
	}
}

func TestAuthDirective_NoRequestInCtx_Unauthorized(t *testing.T) {
	_, err := resolver.AuthDirective(context.Background(), nil, func(ctx context.Context) (any, error) {
		return "ok", nil
	})
	if err == nil {
		t.Error("expected error when no request in context, got nil")
	}
}

// ---- PublishedDocument query tests -----------------------------------------

func TestPublishedDocument_Found(t *testing.T) {
	docUC := &mockDocumentUC{
		getPublishedFn: func(_ context.Context, _, documentID, locale string) (*entity.Document, error) {
			return sampleDoc(documentID), nil
		},
	}
	r := &resolver.Resolver{DocumentUC: docUC, ContentTypeUC: &mockContentTypeUC{}}
	locale := "en"
	doc, err := r.Query().PublishedDocument(context.Background(), "blog", "doc-1", &locale)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc == nil || doc.DocumentID != "doc-1" {
		t.Errorf("got %+v, want DocumentID doc-1", doc)
	}
}

func TestPublishedDocument_NotFound_ReturnsNil(t *testing.T) {
	docUC := &mockDocumentUC{
		getPublishedFn: func(_ context.Context, _, _, _ string) (*entity.Document, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}
	r := &resolver.Resolver{DocumentUC: docUC, ContentTypeUC: &mockContentTypeUC{}}
	doc, err := r.Query().PublishedDocument(context.Background(), "blog", "doc-1", nil)
	if err != nil {
		t.Fatalf("expected nil error for not-found, got: %v", err)
	}
	if doc != nil {
		t.Errorf("expected nil document, got %+v", doc)
	}
}

// ---- ContentTypes query tests ----------------------------------------------

func TestContentTypes(t *testing.T) {
	ctUC := &mockContentTypeUC{
		findAllFn: func(_ context.Context) ([]*entity.ContentType, error) {
			return []*entity.ContentType{
				{ID: "ct-1", Name: "Blog", Slug: "blog", Kind: entity.KindCollection},
			}, nil
		},
	}
	r := &resolver.Resolver{DocumentUC: &mockDocumentUC{}, ContentTypeUC: ctUC}
	cts, err := r.Query().ContentTypes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cts) != 1 || cts[0].Slug != "blog" {
		t.Errorf("got %v, want 1 ContentType with slug blog", cts)
	}
}

// ---- Mutation resolver tests -----------------------------------------------

func TestSaveDocument(t *testing.T) {
	now := time.Now().UTC()
	docUC := &mockDocumentUC{
		saveFn: func(_ context.Context, _ string, doc *entity.Document, _ string) (*entity.Document, error) {
			doc.DocumentID = "doc-1"
			doc.Version = entity.VersionDraft
			doc.CreatedAt = now
			doc.UpdatedAt = now
			doc.CreatedBy = "user-123"
			doc.UpdatedBy = "user-123"
			return doc, nil
		},
	}
	r := &resolver.Resolver{DocumentUC: docUC, ContentTypeUC: &mockContentTypeUC{}}
	locale := "en"
	doc, err := r.Mutation().SaveDocument(authedCtx(), "blog", "doc-1", &locale, map[string]any{"title": "Test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc == nil || doc.DocumentID != "doc-1" {
		t.Errorf("got %+v, want DocumentID doc-1", doc)
	}
}

func TestPublishDocument(t *testing.T) {
	now := time.Now().UTC()
	published := &entity.Document{
		DocumentID: "doc-1", Version: entity.VersionPublished,
		ContentTypeID: "ct-1", Data: map[string]any{"k": "v"},
		Locale: "en", CreatedAt: now, UpdatedAt: now, PublishedAt: now,
		CreatedBy: "u", UpdatedBy: "u", PublishedBy: "user-123",
	}
	docUC := &mockDocumentUC{
		publishFn:      func(_ context.Context, _, _, _, _ string) error { return nil },
		getPublishedFn: func(_ context.Context, _, _, _ string) (*entity.Document, error) { return published, nil },
	}
	r := &resolver.Resolver{DocumentUC: docUC, ContentTypeUC: &mockContentTypeUC{}}
	locale := "en"
	doc, err := r.Mutation().PublishDocument(authedCtx(), "blog", "doc-1", &locale)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc == nil || doc.DocumentID != "doc-1" {
		t.Errorf("got %+v, want DocumentID doc-1", doc)
	}
}

func TestUnpublishDocument(t *testing.T) {
	draft := sampleDoc("doc-1")
	docUC := &mockDocumentUC{
		unpublishFn: func(_ context.Context, _, _, _ string) error { return nil },
		getForEditFn: func(_ context.Context, _, documentID, _ string) (*entity.Document, string, error) {
			return draft, "draft", nil
		},
	}
	r := &resolver.Resolver{DocumentUC: docUC, ContentTypeUC: &mockContentTypeUC{}}
	locale := "en"
	doc, err := r.Mutation().UnpublishDocument(authedCtx(), "blog", "doc-1", &locale)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc == nil || doc.Version != "draft" {
		t.Errorf("got %+v, want version=draft", doc)
	}
}

func TestDeleteDocument(t *testing.T) {
	docUC := &mockDocumentUC{
		deleteFn: func(_ context.Context, _, _ string) error { return nil },
	}
	r := &resolver.Resolver{DocumentUC: docUC, ContentTypeUC: &mockContentTypeUC{}}
	ok, err := r.Mutation().DeleteDocument(authedCtx(), "blog", "doc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected DeleteDocument to return true")
	}
}
