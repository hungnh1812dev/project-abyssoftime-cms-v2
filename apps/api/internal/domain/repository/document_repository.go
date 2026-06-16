package repository

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

// DocumentRepository addresses a logical entry by entryID across its
// draft/published record pair (two physical Mongo documents in the same
// collection — see Domain Rules: Draft & Publish in SPEC.md).
type DocumentRepository interface {
	// FindDraftByEntryID returns the draft record for entryID, or
	// pkgerrors.ErrNotFound if it doesn't exist.
	FindDraftByEntryID(ctx context.Context, entryID string) (*entity.Document, error)
	// FindPublishedByEntryID returns the published record for entryID, or
	// pkgerrors.ErrNotFound if the entry has never been published.
	FindPublishedByEntryID(ctx context.Context, entryID string) (*entity.Document, error)
	// UpsertDraft creates or replaces the draft record for doc.EntryID.
	UpsertDraft(ctx context.Context, doc *entity.Document) error
	// UpsertPublished creates or replaces the published record for doc.EntryID.
	UpsertPublished(ctx context.Context, doc *entity.Document) error
	// FindEntryDraftsByContentType returns the draft record of every entry
	// belonging to contentTypeID — one row per logical entry.
	FindEntryDraftsByContentType(ctx context.Context, contentTypeID string) ([]*entity.Document, error)
	// DeleteByEntryID removes both the draft and published record (if any)
	// for entryID.
	DeleteByEntryID(ctx context.Context, entryID string) error
	// DeletePublishedByEntryID removes only the published record for
	// entryID, leaving the draft untouched. Backs the Unpublish convenience
	// (see tasks/plan.md — kept beyond what SPEC.md defines).
	DeletePublishedByEntryID(ctx context.Context, entryID string) error
	// DeleteByContentType removes every draft and published record
	// belonging to contentTypeID.
	DeleteByContentType(ctx context.Context, contentTypeID string) error
}
