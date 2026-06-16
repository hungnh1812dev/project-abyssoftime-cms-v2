package document

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

type UseCase struct {
	repo      repository.DocumentRepository
	mediaRepo repository.MediaAssetRepository
}

func New(repo repository.DocumentRepository, mediaRepo repository.MediaAssetRepository) *UseCase {
	return &UseCase{repo: repo, mediaRepo: mediaRepo}
}

// Status computes an entry's lifecycle status from its draft and (possibly
// nil, if never published) published record.
func Status(draft, published *entity.Document) string {
	if published == nil {
		return "draft"
	}
	if draft.UpdatedAt.After(published.UpdatedAt) {
		return "modified"
	}
	return "published"
}

// Save creates or updates the draft record for doc.EntryID (generating a
// new EntryID if empty). It never touches the published record, so the
// public/content read API keeps serving the previous published data.
func (uc *UseCase) Save(ctx context.Context, doc *entity.Document, userID string) (*entity.Document, error) {
	if doc.EntryID == "" {
		doc.EntryID = primitive.NewObjectID().Hex()
	}
	existing, err := uc.repo.FindDraftByEntryID(ctx, doc.EntryID)
	if err != nil && !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		return nil, err
	}

	now := time.Now().UTC()
	doc.Version = entity.VersionDraft
	doc.UpdatedAt = now
	doc.UpdatedBy = userID
	if doc.Locale == "" {
		doc.Locale = "en"
	}
	if existing != nil {
		doc.ID = existing.ID
		doc.CreatedAt = existing.CreatedAt
		doc.CreatedBy = existing.CreatedBy
		if doc.ContentTypeID == "" {
			doc.ContentTypeID = existing.ContentTypeID
		}
	} else {
		doc.CreatedAt = now
		doc.CreatedBy = userID
	}

	if err := uc.repo.UpsertDraft(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// GetForEdit returns the draft record plus its computed status, for admin
// editing screens — editors always see their latest unpublished work.
func (uc *UseCase) GetForEdit(ctx context.Context, entryID string) (*entity.Document, string, error) {
	draft, err := uc.repo.FindDraftByEntryID(ctx, entryID)
	if err != nil {
		return nil, "", err
	}
	published, err := uc.repo.FindPublishedByEntryID(ctx, entryID)
	if err != nil {
		if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return nil, "", err
		}
		published = nil
	}
	return draft, Status(draft, published), nil
}

// GetPublished returns only the published record — the public/content read
// path. Returns ErrNotFound if the entry has never been published, even if
// a draft exists.
func (uc *UseCase) GetPublished(ctx context.Context, entryID string) (*entity.Document, error) {
	return uc.repo.FindPublishedByEntryID(ctx, entryID)
}

// GetAll returns the draft record of every entry for a content type — one
// row per logical entry. Used for list views and by content_type.Sync's
// cascade-delete when a content type's definition file is removed.
func (uc *UseCase) GetAll(ctx context.Context, contentTypeID string) ([]*entity.Document, error) {
	return uc.repo.FindEntryDraftsByContentType(ctx, contentTypeID)
}

// Publish copies the draft into the published record, syncing UpdatedAt so
// the public read catches up to the latest saved draft.
func (uc *UseCase) Publish(ctx context.Context, entryID, userID string) error {
	draft, err := uc.repo.FindDraftByEntryID(ctx, entryID)
	if err != nil {
		return err
	}
	published := &entity.Document{
		EntryID:       draft.EntryID,
		ContentTypeID: draft.ContentTypeID,
		Data:          draft.Data,
		Locale:        draft.Locale,
		CreatedAt:     draft.CreatedAt,
		CreatedBy:     draft.CreatedBy,
		UpdatedAt:     draft.UpdatedAt,
		UpdatedBy:     draft.UpdatedBy,
		PublishedAt:   time.Now().UTC(),
		PublishedBy:   userID,
	}
	return uc.repo.UpsertPublished(ctx, published)
}

// Unpublish removes the published record, reverting the entry's computed
// status to draft. This is a CMS convenience beyond SPEC.md's draft/publish
// model (see tasks/plan.md).
func (uc *UseCase) Unpublish(ctx context.Context, entryID string) error {
	return uc.repo.DeletePublishedByEntryID(ctx, entryID)
}

// Delete removes both the draft and published record for entryID, cascading
// to any media assets that reference it first.
func (uc *UseCase) Delete(ctx context.Context, entryID string) error {
	if err := uc.mediaRepo.DeleteByDocumentRef(ctx, entryID); err != nil {
		return err
	}
	return uc.repo.DeleteByEntryID(ctx, entryID)
}
