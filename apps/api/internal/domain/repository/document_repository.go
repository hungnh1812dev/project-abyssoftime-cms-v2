package repository

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

// DocumentRepository addresses a logical entry by documentID across its
// draft/published record pair. Each content type's documents live in a
// standalone collection (documents_<slug>), so every method requires the
// content-type slug to route to the correct collection.
type DocumentRepository interface {
	FindDraftByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error)
	FindPublishedByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error)
	UpsertDraft(ctx context.Context, contentTypeSlug string, doc *entity.Document) error
	UpsertPublished(ctx context.Context, contentTypeSlug string, doc *entity.Document) error
	FindDraftsByContentType(ctx context.Context, contentTypeSlug string) ([]*entity.Document, error)
	FindDraftsByContentTypePaginated(ctx context.Context, contentTypeSlug string, start, size int, locale, orderBy string, sortDir int) ([]*entity.Document, int64, error)
	FindPublishedByContentTypePaginated(ctx context.Context, contentTypeSlug string, start, size int, locale, orderBy string, sortDir int) ([]*entity.Document, int64, error)
	FindPublishedByDocumentIDs(ctx context.Context, contentTypeSlug string, documentIDs []string, locale string) ([]*entity.Document, error)
	DeleteByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) error
	DeletePublishedByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) error
	DeleteAllByContentType(ctx context.Context, contentTypeSlug string) error
	EnsureCollection(ctx context.Context, contentTypeSlug string, fields []entity.FieldDefinition) error
	DropCollection(ctx context.Context, contentTypeSlug string) error
}
