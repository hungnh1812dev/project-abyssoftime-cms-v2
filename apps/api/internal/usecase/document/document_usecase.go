package document

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

type UseCase struct {
	repo             repository.DocumentRepository
	mediaRepo        repository.MediaAssetRepository
	supportedLocales []string
}

func New(repo repository.DocumentRepository, mediaRepo repository.MediaAssetRepository, supportedLocales []string) *UseCase {
	return &UseCase{repo: repo, mediaRepo: mediaRepo, supportedLocales: supportedLocales}
}

// resolveLocale defaults an empty locale to the first supported one and
// rejects any locale outside uc.supportedLocales.
func (uc *UseCase) resolveLocale(locale string) (string, error) {
	if locale == "" {
		locale = uc.supportedLocales[0]
	}
	for _, l := range uc.supportedLocales {
		if l == locale {
			return locale, nil
		}
	}
	return "", fmt.Errorf("%w: unsupported locale %q", pkgerrors.ErrValidation, locale)
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

// Save creates or updates the draft record for (doc.EntryID, doc.Locale)
// (generating a new EntryID if empty). It never touches the published
// record, so the public/content read API keeps serving the previous
// published data.
func (uc *UseCase) Save(ctx context.Context, doc *entity.Document, userID string) (*entity.Document, error) {
	locale, err := uc.resolveLocale(doc.Locale)
	if err != nil {
		return nil, err
	}
	doc.Locale = locale

	if doc.EntryID == "" {
		doc.EntryID = primitive.NewObjectID().Hex()
	}
	existing, err := uc.repo.FindDraftByEntryID(ctx, doc.EntryID, doc.Locale)
	if err != nil && !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		return nil, err
	}

	now := time.Now().UTC()
	doc.Version = entity.VersionDraft
	doc.UpdatedAt = now
	doc.UpdatedBy = userID
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
// editing screens — editors always see their latest unpublished work for
// the given locale.
func (uc *UseCase) GetForEdit(ctx context.Context, entryID, locale string) (*entity.Document, string, error) {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return nil, "", err
	}
	draft, err := uc.repo.FindDraftByEntryID(ctx, entryID, locale)
	if err != nil {
		return nil, "", err
	}
	published, err := uc.repo.FindPublishedByEntryID(ctx, entryID, locale)
	if err != nil {
		if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return nil, "", err
		}
		published = nil
	}
	return draft, Status(draft, published), nil
}

// GetPublished returns only the published record for (entryID, locale) —
// the public/content read path. Returns ErrNotFound if that locale variant
// has never been published, even if a draft exists.
func (uc *UseCase) GetPublished(ctx context.Context, entryID, locale string) (*entity.Document, error) {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return nil, err
	}
	return uc.repo.FindPublishedByEntryID(ctx, entryID, locale)
}

// GetAll returns the draft record of every entry for a content type — one
// row per logical entry, regardless of locale. Used for list views and by
// content_type.Sync's cascade-delete when a content type's definition file
// is removed.
func (uc *UseCase) GetAll(ctx context.Context, contentTypeID string) ([]*entity.Document, error) {
	return uc.repo.FindEntryDraftsByContentType(ctx, contentTypeID)
}

// Publish copies the draft into the published record for (entryID, locale),
// syncing UpdatedAt so the public read catches up to the latest saved
// draft. Publishing one locale never touches another locale's record.
func (uc *UseCase) Publish(ctx context.Context, entryID, locale, userID string) error {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return err
	}
	draft, err := uc.repo.FindDraftByEntryID(ctx, entryID, locale)
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

// Unpublish removes the published record for (entryID, locale), reverting
// that locale variant's computed status to draft. This is a CMS convenience
// beyond SPEC.md's draft/publish model (see tasks/plan.md).
func (uc *UseCase) Unpublish(ctx context.Context, entryID, locale string) error {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return err
	}
	return uc.repo.DeletePublishedByEntryID(ctx, entryID, locale)
}

// Delete removes both the draft and published record for entryID across
// every supported locale, cascading to any media assets that reference it
// first.
func (uc *UseCase) Delete(ctx context.Context, entryID string) error {
	if err := uc.mediaRepo.DeleteByDocumentRef(ctx, entryID); err != nil {
		return err
	}
	for _, locale := range uc.supportedLocales {
		if err := uc.repo.DeleteByEntryID(ctx, entryID, locale); err != nil {
			return err
		}
	}
	return nil
}
