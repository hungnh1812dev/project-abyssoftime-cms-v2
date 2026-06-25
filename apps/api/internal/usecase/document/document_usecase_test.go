package document_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	repomock "project-abyssoftime-cms-v2/api/internal/domain/repository/mock"
	docuc "project-abyssoftime-cms-v2/api/internal/usecase/document"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var ctx = context.Background()

var supportedLocales = []string{"en", "vi"}

const testSlug = "test-slug"

// ---- Save --------------------------------------------------------------

func TestSave_NewEntry_GeneratesDocumentIDAndSetsAudit(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}
	var upserted *entity.Document
	repo.UpsertDraftFn = func(_ context.Context, _ string, doc *entity.Document) error {
		upserted = doc
		return nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	doc := &entity.Document{Fields: map[string]any{"title": "Hello"}}
	saved, err := uc.Save(ctx, testSlug, doc, nil, "user-1")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if saved.DocumentID == "" {
		t.Error("Save() did not generate a DocumentID")
	}
	if saved.CreatedBy != "user-1" || saved.UpdatedBy != "user-1" {
		t.Errorf("Save() CreatedBy/UpdatedBy = %q/%q, want user-1/user-1", saved.CreatedBy, saved.UpdatedBy)
	}
	if saved.Locale != "en" {
		t.Errorf("Save() Locale = %q, want en", saved.Locale)
	}
	if upserted != saved {
		t.Error("Save() did not call UpsertDraft with the same document it returned")
	}
}

func TestSave_ExistingEntry_PreservesCreatedAtAndCreatedBy(t *testing.T) {
	createdAt := time.Now().UTC().Add(-time.Hour)
	existing := &entity.Document{
		DocumentID: "entry-1",
		CreatedAt:  createdAt, CreatedBy: "original-author",
	}

	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, documentID, _ string) (*entity.Document, error) {
		if documentID == "entry-1" {
			return existing, nil
		}
		return nil, pkgerrors.ErrNotFound
	}
	var upserted *entity.Document
	repo.UpsertDraftFn = func(_ context.Context, _ string, doc *entity.Document) error {
		upserted = doc
		return nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	doc := &entity.Document{DocumentID: "entry-1", Fields: map[string]any{"title": "Updated"}}
	saved, err := uc.Save(ctx, testSlug, doc, nil, "editor-2")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if !saved.CreatedAt.Equal(createdAt) {
		t.Errorf("Save() CreatedAt = %v, want %v (preserved)", saved.CreatedAt, createdAt)
	}
	if saved.CreatedBy != "original-author" {
		t.Errorf("Save() CreatedBy = %q, want original-author (preserved)", saved.CreatedBy)
	}
	if saved.UpdatedBy != "editor-2" {
		t.Errorf("Save() UpdatedBy = %q, want editor-2", saved.UpdatedBy)
	}
	_ = upserted
}

func TestSave_RejectsUnsupportedLocale(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.UpsertDraftFn = func(_ context.Context, _ string, _ *entity.Document) error {
		t.Fatal("Save() must not call UpsertDraft for an unsupported locale")
		return nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	doc := &entity.Document{Locale: "fr", Fields: map[string]any{}}
	_, err := uc.Save(ctx, testSlug, doc, nil, "user-1")
	if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("Save() error = %v, want ErrValidation", err)
	}
}

// ---- Status (pure computation) ------------------------------------------

func TestStatus(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name      string
		draft     *entity.Document
		published *entity.Document
		want      string
	}{
		{
			name:  "no published record",
			draft: &entity.Document{UpdatedAt: now},
			want:  "draft",
		},
		{
			name:      "draft newer than published",
			draft:     &entity.Document{UpdatedAt: now},
			published: &entity.Document{UpdatedAt: now.Add(-time.Minute)},
			want:      "modified",
		},
		{
			name:      "timestamps match",
			draft:     &entity.Document{UpdatedAt: now},
			published: &entity.Document{UpdatedAt: now},
			want:      "published",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := docuc.Status(tt.draft, tt.published)
			if got != tt.want {
				t.Errorf("Status() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---- GetForEdit ----------------------------------------------------------

func TestGetForEdit(t *testing.T) {
	now := time.Now().UTC()

	t.Run("draft only", func(t *testing.T) {
		repo := &repomock.DocumentRepository{}
		repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
			return &entity.Document{DocumentID: "e1", UpdatedAt: now}, nil
		}
		repo.FindPublishedByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
			return nil, pkgerrors.ErrNotFound
		}
		uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

		draft, status, err := uc.GetForEdit(ctx, testSlug, "e1", "en", nil)
		if err != nil {
			t.Fatalf("GetForEdit() error = %v", err)
		}
		if draft.DocumentID != "e1" {
			t.Errorf("GetForEdit() draft.DocumentID = %q, want e1", draft.DocumentID)
		}
		if status != "draft" {
			t.Errorf("GetForEdit() status = %q, want draft", status)
		}
	})

	t.Run("draft not found", func(t *testing.T) {
		repo := &repomock.DocumentRepository{}
		repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
			return nil, pkgerrors.ErrNotFound
		}
		uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

		_, _, err := uc.GetForEdit(ctx, testSlug, "missing", "en", nil)
		if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			t.Errorf("GetForEdit() error = %v, want ErrNotFound", err)
		}
	})

	t.Run("defaults to first supported locale when omitted", func(t *testing.T) {
		var gotLocale string
		repo := &repomock.DocumentRepository{}
		repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, locale string) (*entity.Document, error) {
			gotLocale = locale
			return &entity.Document{DocumentID: "e1", UpdatedAt: now}, nil
		}
		repo.FindPublishedByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
			return nil, pkgerrors.ErrNotFound
		}
		uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

		if _, _, err := uc.GetForEdit(ctx, testSlug, "e1", "", nil); err != nil {
			t.Fatalf("GetForEdit() error = %v", err)
		}
		if gotLocale != "en" {
			t.Errorf("GetForEdit() resolved locale = %q, want en (first supported)", gotLocale)
		}
	})

	t.Run("rejects unsupported locale", func(t *testing.T) {
		repo := &repomock.DocumentRepository{}
		uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

		_, _, err := uc.GetForEdit(ctx, testSlug, "e1", "fr", nil)
		if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
			t.Errorf("GetForEdit() error = %v, want ErrValidation", err)
		}
	})
}

// ---- GetPublished ----------------------------------------------------------

func TestGetPublished(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		wantErr error
	}{
		{"found", nil, nil},
		{"not found", pkgerrors.ErrNotFound, pkgerrors.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repomock.DocumentRepository{}
			repo.FindPublishedByDocumentIDFn = func(_ context.Context, _, documentID, _ string) (*entity.Document, error) {
				if tt.repoErr != nil {
					return nil, tt.repoErr
				}
				return &entity.Document{DocumentID: documentID}, nil
			}
			uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

			_, err := uc.GetPublished(ctx, testSlug, "e1", "en", nil)
			if !pkgerrors.Is(err, tt.wantErr) {
				t.Errorf("GetPublished() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetPublished_RejectsUnsupportedLocale(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	_, err := uc.GetPublished(ctx, testSlug, "e1", "fr", nil)
	if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("GetPublished() error = %v, want ErrValidation", err)
	}
}

// ---- GetAll ----------------------------------------------------------------

func TestGetAll_ReturnsEntryDrafts(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftsByContentTypeFn = func(_ context.Context, slug string) ([]*entity.Document, error) {
		return []*entity.Document{
			{DocumentID: "1", },
			{DocumentID: "2", },
		}, nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	docs, err := uc.GetAll(ctx, testSlug)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(docs) != 2 {
		t.Errorf("GetAll() count = %d, want 2", len(docs))
	}
}

// ---- Publish ----------------------------------------------------------------

func TestPublish_CopiesDraftAndSyncsTimestamps(t *testing.T) {
	draftUpdatedAt := time.Now().UTC()
	draft := &entity.Document{
		DocumentID: "e1", Fields: map[string]any{"title": "v2"},
		Locale: "en", UpdatedAt: draftUpdatedAt, UpdatedBy: "editor-1",
	}

	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return draft, nil
	}
	var published *entity.Document
	repo.UpsertPublishedFn = func(_ context.Context, _ string, doc *entity.Document) error {
		published = doc
		return nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	if err := uc.Publish(ctx, testSlug, "e1", "en", nil, "publisher-1"); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if published == nil {
		t.Fatal("Publish() did not call UpsertPublished")
	}
	if !published.UpdatedAt.Equal(draftUpdatedAt) {
		t.Errorf("Publish() published.UpdatedAt = %v, want %v (synced from draft)", published.UpdatedAt, draftUpdatedAt)
	}
	if published.PublishedBy != "publisher-1" {
		t.Errorf("Publish() published.PublishedBy = %q, want publisher-1", published.PublishedBy)
	}
	if published.PublishedAt.IsZero() {
		t.Error("Publish() did not set PublishedAt")
	}
	if published.Fields["title"] != "v2" {
		t.Errorf("Publish() did not copy draft data, got %v", published.Fields)
	}
}

func TestPublish_DraftNotFound(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	if err := uc.Publish(ctx, testSlug, "missing", "en", nil, "publisher-1"); !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("Publish() error = %v, want ErrNotFound", err)
	}
}

func TestPublish_RejectsUnsupportedLocale(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		t.Fatal("Publish() must not query the repository for an unsupported locale")
		return nil, nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	if err := uc.Publish(ctx, testSlug, "e1", "fr", nil, "publisher-1"); !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("Publish() error = %v, want ErrValidation", err)
	}
}

func TestPublish_LocaleIsolation_OnlyTouchesRequestedLocale(t *testing.T) {
	enDraft := &entity.Document{DocumentID: "e1", Locale: "en", Fields: map[string]any{"title": "en-title"}}

	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, locale string) (*entity.Document, error) {
		if locale != "en" {
			t.Fatalf("Publish(%q) queried draft for locale %q, want only en", "en", locale)
		}
		return enDraft, nil
	}
	var publishedLocale string
	repo.UpsertPublishedFn = func(_ context.Context, _ string, doc *entity.Document) error {
		publishedLocale = doc.Locale
		return nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	if err := uc.Publish(ctx, testSlug, "e1", "en", nil, "publisher-1"); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if publishedLocale != "en" {
		t.Errorf("Publish() published locale = %q, want en", publishedLocale)
	}
}

// ---- Unpublish ---------------------------------------------------------------

func TestUnpublish_DeletesPublishedRecord(t *testing.T) {
	var deletedDocID, deletedLocale string
	repo := &repomock.DocumentRepository{}
	repo.DeletePublishedByDocumentIDFn = func(_ context.Context, _, documentID, locale string) error {
		deletedDocID = documentID
		deletedLocale = locale
		return nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	if err := uc.Unpublish(ctx, testSlug, "e1", "en", nil); err != nil {
		t.Fatalf("Unpublish() error = %v", err)
	}
	if deletedDocID != "e1" {
		t.Errorf("Unpublish() deleted document = %q, want e1", deletedDocID)
	}
	if deletedLocale != "en" {
		t.Errorf("Unpublish() deleted locale = %q, want en", deletedLocale)
	}
}

func TestUnpublish_RejectsUnsupportedLocale(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.DeletePublishedByDocumentIDFn = func(_ context.Context, _, _, _ string) error {
		t.Fatal("Unpublish() must not call the repository for an unsupported locale")
		return nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	if err := uc.Unpublish(ctx, testSlug, "e1", "fr", nil); !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("Unpublish() error = %v, want ErrValidation", err)
	}
}

// ---- GetSingleType -----------------------------------------------------------

func TestGetSingleType_ReturnsDocumentAndStatus(t *testing.T) {
	now := time.Now().UTC()
	draft := &entity.Document{DocumentID: "e1", UpdatedAt: now, Locale: "en"}

	repo := &repomock.DocumentRepository{}
	repo.FindDraftsByContentTypePaginatedFn = func(_ context.Context, _ string, _, _ int, _, _ string, _ int) ([]*entity.Document, int64, error) {
		return []*entity.Document{draft}, 1, nil
	}
	repo.FindPublishedByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	doc, status, err := uc.GetSingleType(ctx, testSlug, "en", nil)
	if err != nil {
		t.Fatalf("GetSingleType() error = %v", err)
	}
	if doc.DocumentID != "e1" {
		t.Errorf("GetSingleType() doc.DocumentID = %q, want e1", doc.DocumentID)
	}
	if status != "draft" {
		t.Errorf("GetSingleType() status = %q, want draft", status)
	}
}

func TestGetSingleType_NoDocument_ReturnsNotFound(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftsByContentTypePaginatedFn = func(_ context.Context, _ string, _, _ int, _, _ string, _ int) ([]*entity.Document, int64, error) {
		return nil, 0, nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	_, _, err := uc.GetSingleType(ctx, testSlug, "en", nil)
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("GetSingleType() error = %v, want ErrNotFound", err)
	}
}

func TestGetSingleType_InvalidLocale(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	_, _, err := uc.GetSingleType(ctx, testSlug, "fr", nil)
	if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("GetSingleType() error = %v, want ErrValidation", err)
	}
}

// ---- SaveSingleType ----------------------------------------------------------

func TestSaveSingleType_FirstSave_CreatesNewDocument(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftsByContentTypePaginatedFn = func(_ context.Context, _ string, _, _ int, _, _ string, _ int) ([]*entity.Document, int64, error) {
		return nil, 0, nil
	}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}
	repo.UpsertDraftFn = func(_ context.Context, _ string, _ *entity.Document) error { return nil }
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	doc, err := uc.SaveSingleType(ctx, testSlug, map[string]any{"title": "Hello"}, "en", nil, "user-1")
	if err != nil {
		t.Fatalf("SaveSingleType() error = %v", err)
	}
	if doc.DocumentID == "" {
		t.Error("SaveSingleType() did not generate a DocumentID")
	}
}

func TestSaveSingleType_SubsequentSave_ReusesDocumentID(t *testing.T) {
	existing := &entity.Document{DocumentID: "existing-id", Locale: "en"}

	repo := &repomock.DocumentRepository{}
	repo.FindDraftsByContentTypePaginatedFn = func(_ context.Context, _ string, _, _ int, _, _ string, _ int) ([]*entity.Document, int64, error) {
		return []*entity.Document{existing}, 1, nil
	}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, docID, _ string) (*entity.Document, error) {
		if docID == "existing-id" {
			return existing, nil
		}
		return nil, pkgerrors.ErrNotFound
	}
	repo.UpsertDraftFn = func(_ context.Context, _ string, _ *entity.Document) error { return nil }
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	doc, err := uc.SaveSingleType(ctx, testSlug, map[string]any{"title": "Updated"}, "en", nil, "user-1")
	if err != nil {
		t.Fatalf("SaveSingleType() error = %v", err)
	}
	if doc.DocumentID != "existing-id" {
		t.Errorf("SaveSingleType() DocumentID = %q, want existing-id", doc.DocumentID)
	}
}

// ---- PublishSingleType -------------------------------------------------------

func TestPublishSingleType_Delegates(t *testing.T) {
	draft := &entity.Document{DocumentID: "e1", Locale: "en", Fields: map[string]any{"title": "v1"}, UpdatedAt: time.Now().UTC()}

	repo := &repomock.DocumentRepository{}
	repo.FindDraftsByContentTypePaginatedFn = func(_ context.Context, _ string, _, _ int, _, _ string, _ int) ([]*entity.Document, int64, error) {
		return []*entity.Document{draft}, 1, nil
	}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return draft, nil
	}
	var publishedDocID string
	repo.UpsertPublishedFn = func(_ context.Context, _ string, doc *entity.Document) error {
		publishedDocID = doc.DocumentID
		return nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	if err := uc.PublishSingleType(ctx, testSlug, "en", nil, "pub-1"); err != nil {
		t.Fatalf("PublishSingleType() error = %v", err)
	}
	if publishedDocID != "e1" {
		t.Errorf("PublishSingleType() published documentID = %q, want e1", publishedDocID)
	}
}

func TestPublishSingleType_NoDocument(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftsByContentTypePaginatedFn = func(_ context.Context, _ string, _, _ int, _, _ string, _ int) ([]*entity.Document, int64, error) {
		return nil, 0, nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	if err := uc.PublishSingleType(ctx, testSlug, "en", nil, "pub-1"); !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("PublishSingleType() error = %v, want ErrNotFound", err)
	}
}

// ---- UnpublishSingleType -----------------------------------------------------

func TestUnpublishSingleType_Delegates(t *testing.T) {
	draft := &entity.Document{DocumentID: "e1", Locale: "en"}

	repo := &repomock.DocumentRepository{}
	repo.FindDraftsByContentTypePaginatedFn = func(_ context.Context, _ string, _, _ int, _, _ string, _ int) ([]*entity.Document, int64, error) {
		return []*entity.Document{draft}, 1, nil
	}
	var deletedDocID string
	repo.DeletePublishedByDocumentIDFn = func(_ context.Context, _, documentID, _ string) error {
		deletedDocID = documentID
		return nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	if err := uc.UnpublishSingleType(ctx, testSlug, "en", nil); err != nil {
		t.Fatalf("UnpublishSingleType() error = %v", err)
	}
	if deletedDocID != "e1" {
		t.Errorf("UnpublishSingleType() deleted documentID = %q, want e1", deletedDocID)
	}
}

func TestUnpublishSingleType_NoDocument(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftsByContentTypePaginatedFn = func(_ context.Context, _ string, _, _ int, _, _ string, _ int) ([]*entity.Document, int64, error) {
		return nil, 0, nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	if err := uc.UnpublishSingleType(ctx, testSlug, "en", nil); !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("UnpublishSingleType() error = %v, want ErrNotFound", err)
	}
}

// ---- GetAllPaginated ---------------------------------------------------------

func TestGetAllPaginated_ReturnsPageWithStatuses(t *testing.T) {
	now := time.Now().UTC()
	drafts := []*entity.Document{
		{DocumentID: "d1", UpdatedAt: now, Locale: "en"},
		{DocumentID: "d2", UpdatedAt: now, Locale: "en"},
	}
	pub := &entity.Document{DocumentID: "d1", UpdatedAt: now, Locale: "en"}

	repo := &repomock.DocumentRepository{}
	repo.FindDraftsByContentTypePaginatedFn = func(_ context.Context, _ string, start, size int, _, _ string, _ int) ([]*entity.Document, int64, error) {
		if start != 0 || size != 2 {
			t.Errorf("unexpected start=%d size=%d", start, size)
		}
		return drafts, 5, nil
	}
	repo.FindPublishedByDocumentIDsFn = func(_ context.Context, _ string, ids []string, _ string) ([]*entity.Document, error) {
		if len(ids) != 2 {
			t.Errorf("FindPublishedByDocumentIDs() ids = %v, want 2 items", ids)
		}
		return []*entity.Document{pub}, nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	docs, statuses, total, err := uc.GetAllPaginated(ctx, testSlug, 0, 2, "en", nil, "id", -1)
	if err != nil {
		t.Fatalf("GetAllPaginated() error = %v", err)
	}
	if len(docs) != 2 {
		t.Fatalf("GetAllPaginated() docs count = %d, want 2", len(docs))
	}
	if total != 5 {
		t.Errorf("GetAllPaginated() total = %d, want 5", total)
	}
	if statuses[0] != "published" {
		t.Errorf("GetAllPaginated() statuses[0] = %q, want published", statuses[0])
	}
	if statuses[1] != "draft" {
		t.Errorf("GetAllPaginated() statuses[1] = %q, want draft", statuses[1])
	}
}

func TestGetAllPaginated_EmptyResult(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftsByContentTypePaginatedFn = func(_ context.Context, _ string, _, _ int, _, _ string, _ int) ([]*entity.Document, int64, error) {
		return nil, 0, nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	docs, statuses, total, err := uc.GetAllPaginated(ctx, testSlug, 0, 20, "en", nil, "id", -1)
	if err != nil {
		t.Fatalf("GetAllPaginated() error = %v", err)
	}
	if total != 0 || len(docs) != 0 || statuses != nil {
		t.Errorf("GetAllPaginated() empty result unexpected: docs=%v, statuses=%v, total=%d", docs, statuses, total)
	}
}

func TestGetAllPaginated_InvalidLocale(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	_, _, _, err := uc.GetAllPaginated(ctx, testSlug, 0, 20, "fr", nil, "id", -1)
	if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("GetAllPaginated() error = %v, want ErrValidation", err)
	}
}

// ---- Duplicate -------------------------------------------------------

func TestDuplicate_Success(t *testing.T) {
	source := &entity.Document{
		DocumentID: "src-1",
		Fields:     map[string]any{"title": "Original", "count": float64(42)},
		Locale:     "en",
	}

	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, documentID, _ string) (*entity.Document, error) {
		if documentID == "src-1" {
			return source, nil
		}
		return nil, pkgerrors.ErrNotFound
	}
	var upserted *entity.Document
	repo.UpsertDraftFn = func(_ context.Context, _ string, doc *entity.Document) error {
		upserted = doc
		return nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	saved, err := uc.Duplicate(ctx, testSlug, "src-1", "en", nil, "user-2")
	if err != nil {
		t.Fatalf("Duplicate() error = %v", err)
	}
	if saved.DocumentID == "" || saved.DocumentID == "src-1" {
		t.Errorf("Duplicate() DocumentID = %q, want new UUID", saved.DocumentID)
	}
	if saved.Fields["title"] != "Original" {
		t.Errorf("Duplicate() title = %v, want Original", saved.Fields["title"])
	}
	if saved.Fields["count"] != float64(42) {
		t.Errorf("Duplicate() count = %v, want 42", saved.Fields["count"])
	}
	if saved.Version != entity.VersionDraft {
		t.Errorf("Duplicate() Version = %q, want draft", saved.Version)
	}
	if saved.CreatedBy != "user-2" || saved.UpdatedBy != "user-2" {
		t.Errorf("Duplicate() CreatedBy/UpdatedBy = %q/%q, want user-2", saved.CreatedBy, saved.UpdatedBy)
	}
	if upserted == nil {
		t.Fatal("Duplicate() did not call UpsertDraft")
	}
}

func TestDuplicate_CopiesComponents(t *testing.T) {
	source := &entity.Document{
		DocumentID: "src-1",
		Fields:     map[string]any{"title": "With components"},
		Locale:     "en",
	}
	compFields := []entity.FieldDefinition{
		{Name: "title", Type: "text"},
		{Name: "banner", Type: "component"},
	}
	bannerData := []*entity.Component{
		{ComponentID: "comp-1", Fields: map[string]any{"heading": "Hello"}},
	}

	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, documentID, _ string) (*entity.Document, error) {
		if documentID == "src-1" {
			return source, nil
		}
		return nil, pkgerrors.ErrNotFound
	}
	repo.UpsertDraftFn = func(_ context.Context, _ string, _ *entity.Document) error { return nil }

	compRepo := &repomock.ComponentRepository{}
	compRepo.FindByDocumentIDFn = func(_ context.Context, _, compName, docID, _ string, _ entity.DocumentVersion) ([]*entity.Component, error) {
		if docID == "src-1" && compName == "banner" {
			return bannerData, nil
		}
		return nil, nil
	}
	var savedComponents []*entity.Component
	compRepo.UpsertAllFn = func(_ context.Context, _, _, _, _ string, _ entity.DocumentVersion, comps []*entity.Component) error {
		savedComponents = comps
		return nil
	}

	uc := docuc.New(repo, compRepo, &repomock.MediaAssetRepository{}, supportedLocales)

	saved, err := uc.Duplicate(ctx, testSlug, "src-1", "en", compFields, "user-2")
	if err != nil {
		t.Fatalf("Duplicate() error = %v", err)
	}
	if saved == nil {
		t.Fatal("Duplicate() returned nil")
	}
	if len(savedComponents) == 0 {
		t.Fatal("Duplicate() did not save components")
	}
	if savedComponents[0].Fields["heading"] != "Hello" {
		t.Errorf("Duplicate() component heading = %v, want Hello", savedComponents[0].Fields["heading"])
	}
}

func TestDuplicate_SourceNotFound(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	_, err := uc.Duplicate(ctx, testSlug, "missing", "en", nil, "user-1")
	if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		t.Errorf("Duplicate() error = %v, want ErrNotFound", err)
	}
}

func TestDuplicate_SharesMediaRefs(t *testing.T) {
	source := &entity.Document{
		DocumentID: "src-1",
		Fields:     map[string]any{"title": "Post", "cover": "media-doc-id-123"},
		Locale:     "en",
	}

	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, documentID, _ string) (*entity.Document, error) {
		if documentID == "src-1" {
			return source, nil
		}
		return nil, pkgerrors.ErrNotFound
	}
	var upserted *entity.Document
	repo.UpsertDraftFn = func(_ context.Context, _ string, doc *entity.Document) error {
		upserted = doc
		return nil
	}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	_, err := uc.Duplicate(ctx, testSlug, "src-1", "en", nil, "user-2")
	if err != nil {
		t.Fatalf("Duplicate() error = %v", err)
	}
	if upserted.Fields["cover"] != "media-doc-id-123" {
		t.Errorf("Duplicate() cover = %v, want media-doc-id-123 (shared ref)", upserted.Fields["cover"])
	}
}

func TestDuplicate_InvalidLocale(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)

	_, err := uc.Duplicate(ctx, testSlug, "src-1", "fr", nil, "user-1")
	if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("Duplicate() error = %v, want ErrValidation", err)
	}
}

// ---- Delete ---------------------------------------------------------

// ---- Repeatable Component Validation ----------------------------------------

func TestSave_RepeatableComponent_ArrayData_Success(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}
	repo.UpsertDraftFn = func(_ context.Context, _ string, _ *entity.Document) error { return nil }

	var savedComps []*entity.Component
	compRepo := &repomock.ComponentRepository{}
	compRepo.UpsertAllFn = func(_ context.Context, _, _, _, _ string, _ entity.DocumentVersion, comps []*entity.Component) error {
		savedComps = comps
		return nil
	}

	fields := []entity.FieldDefinition{
		{Name: "skills", Type: "component", Repeatable: true},
	}
	doc := &entity.Document{
		Fields: map[string]any{
			"skills": []any{
				map[string]any{"category": "Frontend"},
				map[string]any{"category": "Backend"},
			},
		},
	}

	uc := docuc.New(repo, compRepo, &repomock.MediaAssetRepository{}, supportedLocales)
	_, err := uc.Save(ctx, testSlug, doc, fields, "user-1")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if len(savedComps) != 2 {
		t.Fatalf("saved %d components, want 2", len(savedComps))
	}
	if savedComps[0].SortOrder != 0 {
		t.Errorf("savedComps[0].SortOrder = %d, want 0", savedComps[0].SortOrder)
	}
	if savedComps[1].SortOrder != 1 {
		t.Errorf("savedComps[1].SortOrder = %d, want 1", savedComps[1].SortOrder)
	}
}

func TestSave_RepeatableComponent_ObjectData_ReturnsError(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}
	repo.UpsertDraftFn = func(_ context.Context, _ string, _ *entity.Document) error { return nil }
	compRepo := &repomock.ComponentRepository{}

	fields := []entity.FieldDefinition{
		{Name: "skills", Type: "component", Repeatable: true},
	}
	doc := &entity.Document{
		Fields: map[string]any{
			"skills": map[string]any{"category": "Frontend"},
		},
	}

	uc := docuc.New(repo, compRepo, &repomock.MediaAssetRepository{}, supportedLocales)
	_, err := uc.Save(ctx, testSlug, doc, fields, "user-1")
	if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("Save() error = %v, want ErrValidation", err)
	}
}

func TestSave_NonRepeatableComponent_ObjectData_Success(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}
	repo.UpsertDraftFn = func(_ context.Context, _ string, _ *entity.Document) error { return nil }

	var savedComps []*entity.Component
	compRepo := &repomock.ComponentRepository{}
	compRepo.UpsertAllFn = func(_ context.Context, _, _, _, _ string, _ entity.DocumentVersion, comps []*entity.Component) error {
		savedComps = comps
		return nil
	}

	fields := []entity.FieldDefinition{
		{Name: "banner", Type: "component", Repeatable: false},
	}
	doc := &entity.Document{
		Fields: map[string]any{
			"banner": map[string]any{"title": "Hello"},
		},
	}

	uc := docuc.New(repo, compRepo, &repomock.MediaAssetRepository{}, supportedLocales)
	_, err := uc.Save(ctx, testSlug, doc, fields, "user-1")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if len(savedComps) != 1 {
		t.Fatalf("saved %d components, want 1", len(savedComps))
	}
	if savedComps[0].SortOrder != 0 {
		t.Errorf("savedComps[0].SortOrder = %d, want 0", savedComps[0].SortOrder)
	}
}

func TestSave_NonRepeatableComponent_ArrayData_ReturnsError(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}
	repo.UpsertDraftFn = func(_ context.Context, _ string, _ *entity.Document) error { return nil }
	compRepo := &repomock.ComponentRepository{}

	fields := []entity.FieldDefinition{
		{Name: "banner", Type: "component", Repeatable: false},
	}
	doc := &entity.Document{
		Fields: map[string]any{
			"banner": []any{map[string]any{"title": "Hello"}},
		},
	}

	uc := docuc.New(repo, compRepo, &repomock.MediaAssetRepository{}, supportedLocales)
	_, err := uc.Save(ctx, testSlug, doc, fields, "user-1")
	if !pkgerrors.Is(err, pkgerrors.ErrValidation) {
		t.Errorf("Save() error = %v, want ErrValidation", err)
	}
}

func TestSave_RepeatableComponent_EmptyArray_Success(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}
	repo.UpsertDraftFn = func(_ context.Context, _ string, _ *entity.Document) error { return nil }

	var savedComps []*entity.Component
	compRepo := &repomock.ComponentRepository{}
	compRepo.UpsertAllFn = func(_ context.Context, _, _, _, _ string, _ entity.DocumentVersion, comps []*entity.Component) error {
		savedComps = comps
		return nil
	}

	fields := []entity.FieldDefinition{
		{Name: "skills", Type: "component", Repeatable: true},
	}
	doc := &entity.Document{
		Fields: map[string]any{
			"skills": []any{},
		},
	}

	uc := docuc.New(repo, compRepo, &repomock.MediaAssetRepository{}, supportedLocales)
	_, err := uc.Save(ctx, testSlug, doc, fields, "user-1")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if len(savedComps) != 0 {
		t.Errorf("saved %d components, want 0", len(savedComps))
	}
}

// ---- Merge Component Shape --------------------------------------------------

func TestMergeComponents_Repeatable_SingleItem_ReturnsArray(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	compRepo := &repomock.ComponentRepository{}
	compRepo.FindByDocumentIDFn = func(_ context.Context, _, _, _, _ string, _ entity.DocumentVersion) ([]*entity.Component, error) {
		return []*entity.Component{
			{ComponentID: "c1", Fields: map[string]any{"category": "Frontend"}},
		}, nil
	}

	fields := []entity.FieldDefinition{
		{Name: "skills", Type: "component", Repeatable: true},
	}

	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return &entity.Document{DocumentID: "d1", Locale: "en", Version: entity.VersionDraft, Fields: map[string]any{}, UpdatedAt: time.Now()}, nil
	}
	repo.FindPublishedByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}

	uc2 := docuc.New(repo, compRepo, &repomock.MediaAssetRepository{}, supportedLocales)
	result, _, err := uc2.GetForEdit(ctx, testSlug, "d1", "en", fields)
	if err != nil {
		t.Fatalf("GetForEdit() error = %v", err)
	}

	raw := result.Fields["skills"]
	arr, ok := raw.([]map[string]any)
	if !ok {
		t.Fatalf("skills type = %T, want []map[string]any", raw)
	}
	if len(arr) != 1 {
		t.Fatalf("skills len = %d, want 1", len(arr))
	}
	if arr[0]["category"] != "Frontend" {
		t.Errorf("skills[0].category = %v, want Frontend", arr[0]["category"])
	}
}

func TestMergeComponents_Repeatable_ZeroItems_ReturnsEmptyArray(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return &entity.Document{DocumentID: "d1", Locale: "en", Version: entity.VersionDraft, Fields: map[string]any{}, UpdatedAt: time.Now()}, nil
	}
	repo.FindPublishedByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}

	compRepo := &repomock.ComponentRepository{}
	compRepo.FindByDocumentIDFn = func(_ context.Context, _, _, _, _ string, _ entity.DocumentVersion) ([]*entity.Component, error) {
		return nil, nil
	}

	fields := []entity.FieldDefinition{
		{Name: "skills", Type: "component", Repeatable: true},
	}

	uc := docuc.New(repo, compRepo, &repomock.MediaAssetRepository{}, supportedLocales)
	result, _, err := uc.GetForEdit(ctx, testSlug, "d1", "en", fields)
	if err != nil {
		t.Fatalf("GetForEdit() error = %v", err)
	}

	raw := result.Fields["skills"]
	arr, ok := raw.([]map[string]any)
	if !ok {
		t.Fatalf("skills type = %T, want []map[string]any", raw)
	}
	if len(arr) != 0 {
		t.Errorf("skills len = %d, want 0", len(arr))
	}
}

func TestMergeComponents_NonRepeatable_SingleItem_ReturnsObject(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return &entity.Document{DocumentID: "d1", Locale: "en", Version: entity.VersionDraft, Fields: map[string]any{}, UpdatedAt: time.Now()}, nil
	}
	repo.FindPublishedByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}

	compRepo := &repomock.ComponentRepository{}
	compRepo.FindByDocumentIDFn = func(_ context.Context, _, _, _, _ string, _ entity.DocumentVersion) ([]*entity.Component, error) {
		return []*entity.Component{
			{ComponentID: "c1", Fields: map[string]any{"title": "Banner Title"}},
		}, nil
	}

	fields := []entity.FieldDefinition{
		{Name: "banner", Type: "component", Repeatable: false},
	}

	uc := docuc.New(repo, compRepo, &repomock.MediaAssetRepository{}, supportedLocales)
	result, _, err := uc.GetForEdit(ctx, testSlug, "d1", "en", fields)
	if err != nil {
		t.Fatalf("GetForEdit() error = %v", err)
	}

	raw := result.Fields["banner"]
	m, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("banner type = %T, want map[string]any", raw)
	}
	if m["title"] != "Banner Title" {
		t.Errorf("banner.title = %v, want Banner Title", m["title"])
	}
}

// ---- Delete ---------------------------------------------------------

func TestDelete_DeletesAllLocales(t *testing.T) {
	var deletedLocales []string

	docRepo := &repomock.DocumentRepository{}
	docRepo.DeleteByDocumentIDFn = func(_ context.Context, _, documentID, locale string) error {
		deletedLocales = append(deletedLocales, fmt.Sprintf("%s:%s", documentID, locale))
		return nil
	}

	uc := docuc.New(docRepo, nil, nil, supportedLocales)
	if err := uc.Delete(ctx, testSlug, "e1", nil); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	want := []string{"e1:en", "e1:vi"}
	if len(deletedLocales) != len(want) {
		t.Fatalf("Delete() called %v, want %v", deletedLocales, want)
	}
	for i, w := range want {
		if deletedLocales[i] != w {
			t.Errorf("Delete() call[%d] = %q, want %q", i, deletedLocales[i], w)
		}
	}
}

func TestSave_SanitizesEmptyStringOnNumberField(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}
	var upserted *entity.Document
	repo.UpsertDraftFn = func(_ context.Context, _ string, doc *entity.Document) error {
		upserted = doc
		return nil
	}

	fields := []entity.FieldDefinition{
		{Name: "title", Type: "text"},
		{Name: "count", Type: "number"},
		{Name: "active", Type: "boolean"},
		{Name: "avatar", Type: "media"},
		{Name: "metadata", Type: "json"},
	}
	doc := &entity.Document{
		Fields: map[string]any{
			"title":    "",
			"count":    "",
			"active":   "",
			"avatar":   "",
			"metadata": "",
		},
	}

	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)
	_, err := uc.Save(ctx, testSlug, doc, fields, "user-1")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if upserted.Fields["title"] != "" {
		t.Errorf("title = %v, want empty string (text fields preserve empty strings)", upserted.Fields["title"])
	}
	if upserted.Fields["count"] != nil {
		t.Errorf("count = %v, want nil (number fields coerce empty string to nil)", upserted.Fields["count"])
	}
	if upserted.Fields["active"] != nil {
		t.Errorf("active = %v, want nil (boolean fields coerce empty string to nil)", upserted.Fields["active"])
	}
	if upserted.Fields["avatar"] != nil {
		t.Errorf("avatar = %v, want nil (media fields coerce empty string to nil)", upserted.Fields["avatar"])
	}
	if upserted.Fields["metadata"] != nil {
		t.Errorf("metadata = %v, want nil (json fields coerce empty string to nil)", upserted.Fields["metadata"])
	}
}

func TestSave_SanitizesEmptyStringInsideComponent(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}
	repo.UpsertDraftFn = func(_ context.Context, _ string, _ *entity.Document) error { return nil }

	var savedComps []*entity.Component
	compRepo := &repomock.ComponentRepository{}
	compRepo.UpsertAllFn = func(_ context.Context, _, _, _, _ string, _ entity.DocumentVersion, comps []*entity.Component) error {
		savedComps = comps
		return nil
	}

	fields := []entity.FieldDefinition{
		{
			Name: "banner",
			Type: "component",
			Fields: []entity.FieldDefinition{
				{Name: "subtitle", Type: "text"},
				{Name: "rating", Type: "number"},
			},
		},
	}
	doc := &entity.Document{
		Fields: map[string]any{
			"banner": map[string]any{
				"subtitle": "",
				"rating":   "",
			},
		},
	}

	uc := docuc.New(repo, compRepo, &repomock.MediaAssetRepository{}, supportedLocales)
	_, err := uc.Save(ctx, testSlug, doc, fields, "user-1")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if len(savedComps) != 1 {
		t.Fatalf("saved %d components, want 1", len(savedComps))
	}
	if savedComps[0].Fields["subtitle"] != "" {
		t.Errorf("subtitle = %v, want empty string", savedComps[0].Fields["subtitle"])
	}
	if savedComps[0].Fields["rating"] != nil {
		t.Errorf("rating = %v, want nil", savedComps[0].Fields["rating"])
	}
}

func TestSave_SanitizesMediaObjectToDocumentID(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindDraftByDocumentIDFn = func(_ context.Context, _, _, _ string) (*entity.Document, error) {
		return nil, pkgerrors.ErrNotFound
	}
	var upserted *entity.Document
	repo.UpsertDraftFn = func(_ context.Context, _ string, doc *entity.Document) error {
		upserted = doc
		return nil
	}

	fields := []entity.FieldDefinition{
		{Name: "avatar", Type: "media"},
	}
	doc := &entity.Document{
		Fields: map[string]any{
			"avatar": map[string]any{
				"documentId":   "abc-123",
				"url":          "https://example.com/photo.jpg",
				"thumbnailUrl": "https://example.com/photo_thumb.jpg",
				"fileName":     "photo.jpg",
			},
		},
	}

	uc := docuc.New(repo, nil, &repomock.MediaAssetRepository{}, supportedLocales)
	_, err := uc.Save(ctx, testSlug, doc, fields, "user-1")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if upserted.Fields["avatar"] != "abc-123" {
		t.Errorf("avatar = %v, want %q (media object should be reduced to documentId)", upserted.Fields["avatar"], "abc-123")
	}
}
