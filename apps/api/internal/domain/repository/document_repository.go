package repository

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type DocumentRepository interface {
	// Deprecated: the methods below operate on a single Mongo record by its
	// own _id. They are superseded by the entry-aware methods further down,
	// which address a logical entry by entryID across its draft/published
	// record pair. Removed once the document usecase migrates (see B2).
	Create(ctx context.Context, doc *entity.Document) error
	FindByID(ctx context.Context, id string) (*entity.Document, error)
	FindByContentType(ctx context.Context, contentTypeID string) ([]*entity.Document, error)
	Update(ctx context.Context, doc *entity.Document) error
	UpdateStatus(ctx context.Context, id string, status entity.DocumentStatus) error
	Delete(ctx context.Context, id string) error

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
	// DeleteByContentType removes every draft and published record
	// belonging to contentTypeID.
	DeleteByContentType(ctx context.Context, contentTypeID string) error
}
