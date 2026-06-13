package document_test

import (
	"context"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	repomock "project-abyssoftime-cms-v2/api/internal/domain/repository/mock"
	docuc "project-abyssoftime-cms-v2/api/internal/usecase/document"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var ctx = context.Background()

// ---- Create ----------------------------------------------------------------

func TestCreate(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.CreateFn = func(_ context.Context, doc *entity.Document) error {
		doc.ID = "new-id"
		doc.Status = entity.StatusDraft
		return nil
	}
	uc := docuc.New(repo, &repomock.MediaAssetRepository{})

	doc := &entity.Document{ContentTypeID: "ct-1", Data: map[string]any{"title": "Hello"}}
	if err := uc.Create(ctx, doc); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if doc.ID == "" {
		t.Error("Create() did not set ID")
	}
}

// ---- GetOne ----------------------------------------------------------------

func TestGetOne(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		repoErr error
		wantErr error
	}{
		{"found", "abc", nil, nil},
		{"not found", "x", pkgerrors.ErrNotFound, pkgerrors.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repomock.DocumentRepository{}
			repo.FindByIDFn = func(_ context.Context, id string) (*entity.Document, error) {
				if tt.repoErr != nil {
					return nil, tt.repoErr
				}
				return &entity.Document{ID: id}, nil
			}
			uc := docuc.New(repo, &repomock.MediaAssetRepository{})
			doc, err := uc.GetOne(ctx, tt.id)
			if !pkgerrors.Is(err, tt.wantErr) {
				t.Errorf("GetOne() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && doc.ID != tt.id {
				t.Errorf("GetOne() ID = %v, want %v", doc.ID, tt.id)
			}
		})
	}
}

// ---- GetAll ----------------------------------------------------------------

func TestGetAll(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.FindByContentTypeFn = func(_ context.Context, contentTypeID string) ([]*entity.Document, error) {
		return []*entity.Document{
			{ID: "1", ContentTypeID: contentTypeID},
			{ID: "2", ContentTypeID: contentTypeID},
		}, nil
	}
	uc := docuc.New(repo, &repomock.MediaAssetRepository{})
	docs, err := uc.GetAll(ctx, "ct-1")
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(docs) != 2 {
		t.Errorf("GetAll() count = %d, want 2", len(docs))
	}
}

// ---- Update ----------------------------------------------------------------

func TestUpdate(t *testing.T) {
	repo := &repomock.DocumentRepository{}
	repo.UpdateFn = func(_ context.Context, _ *entity.Document) error { return nil }
	uc := docuc.New(repo, &repomock.MediaAssetRepository{})

	doc := &entity.Document{ID: "abc", Data: map[string]any{"title": "Updated"}}
	if err := uc.Update(ctx, doc); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
}

// ---- Publish / Unpublish ---------------------------------------------------

func TestPublish(t *testing.T) {
	var capturedStatus entity.DocumentStatus
	repo := &repomock.DocumentRepository{}
	repo.UpdateStatusFn = func(_ context.Context, _ string, status entity.DocumentStatus) error {
		capturedStatus = status
		return nil
	}
	uc := docuc.New(repo, &repomock.MediaAssetRepository{})

	if err := uc.Publish(ctx, "abc"); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if capturedStatus != entity.StatusPublished {
		t.Errorf("Publish() set status = %q, want %q", capturedStatus, entity.StatusPublished)
	}
}

func TestUnpublish(t *testing.T) {
	var capturedStatus entity.DocumentStatus
	repo := &repomock.DocumentRepository{}
	repo.UpdateStatusFn = func(_ context.Context, _ string, status entity.DocumentStatus) error {
		capturedStatus = status
		return nil
	}
	uc := docuc.New(repo, &repomock.MediaAssetRepository{})

	if err := uc.Unpublish(ctx, "abc"); err != nil {
		t.Fatalf("Unpublish() error = %v", err)
	}
	if capturedStatus != entity.StatusDraft {
		t.Errorf("Unpublish() set status = %q, want %q", capturedStatus, entity.StatusDraft)
	}
}

// ---- Delete (cascade) ------------------------------------------------------

func TestDelete_CascadeOrder(t *testing.T) {
	var callOrder []string

	mediaRepo := &repomock.MediaAssetRepository{}
	mediaRepo.DeleteByDocumentRefFn = func(_ context.Context, docRef string) error {
		callOrder = append(callOrder, "media:"+docRef)
		return nil
	}

	docRepo := &repomock.DocumentRepository{}
	docRepo.DeleteFn = func(_ context.Context, id string) error {
		callOrder = append(callOrder, "doc:"+id)
		return nil
	}

	uc := docuc.New(docRepo, mediaRepo)
	if err := uc.Delete(ctx, "abc"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if len(callOrder) != 2 {
		t.Fatalf("Delete() called %d operations, want 2", len(callOrder))
	}
	if callOrder[0] != "media:abc" {
		t.Errorf("Delete() first call = %q, want media:abc", callOrder[0])
	}
	if callOrder[1] != "doc:abc" {
		t.Errorf("Delete() second call = %q, want doc:abc", callOrder[1])
	}
}

func TestDelete_MediaError_Aborts(t *testing.T) {
	mediaRepo := &repomock.MediaAssetRepository{}
	mediaRepo.DeleteByDocumentRefFn = func(_ context.Context, _ string) error {
		return pkgerrors.ErrNotFound
	}
	docDeleteCalled := false
	docRepo := &repomock.DocumentRepository{}
	docRepo.DeleteFn = func(_ context.Context, _ string) error {
		docDeleteCalled = true
		return nil
	}

	uc := docuc.New(docRepo, mediaRepo)
	err := uc.Delete(ctx, "abc")
	if err == nil {
		t.Error("Delete() should have returned error when media delete fails")
	}
	if docDeleteCalled {
		t.Error("Delete() must not delete document when media cascade fails")
	}
}
