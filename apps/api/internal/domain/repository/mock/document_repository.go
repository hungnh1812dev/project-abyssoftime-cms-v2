package mock

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.DocumentRepository = (*DocumentRepository)(nil)

// DocumentRepository is a test double for repository.DocumentRepository.
// Set each Fn field to a stub before calling the method under test.
type DocumentRepository struct {
	FindDraftByEntryIDFn           func(ctx context.Context, entryID string) (*entity.Document, error)
	FindPublishedByEntryIDFn       func(ctx context.Context, entryID string) (*entity.Document, error)
	UpsertDraftFn                  func(ctx context.Context, doc *entity.Document) error
	UpsertPublishedFn              func(ctx context.Context, doc *entity.Document) error
	FindEntryDraftsByContentTypeFn func(ctx context.Context, contentTypeID string) ([]*entity.Document, error)
	DeleteByEntryIDFn              func(ctx context.Context, entryID string) error
	DeletePublishedByEntryIDFn     func(ctx context.Context, entryID string) error
	DeleteByContentTypeFn          func(ctx context.Context, contentTypeID string) error
}

func (m *DocumentRepository) FindDraftByEntryID(ctx context.Context, entryID string) (*entity.Document, error) {
	return m.FindDraftByEntryIDFn(ctx, entryID)
}

func (m *DocumentRepository) FindPublishedByEntryID(ctx context.Context, entryID string) (*entity.Document, error) {
	return m.FindPublishedByEntryIDFn(ctx, entryID)
}

func (m *DocumentRepository) UpsertDraft(ctx context.Context, doc *entity.Document) error {
	return m.UpsertDraftFn(ctx, doc)
}

func (m *DocumentRepository) UpsertPublished(ctx context.Context, doc *entity.Document) error {
	return m.UpsertPublishedFn(ctx, doc)
}

func (m *DocumentRepository) FindEntryDraftsByContentType(ctx context.Context, contentTypeID string) ([]*entity.Document, error) {
	return m.FindEntryDraftsByContentTypeFn(ctx, contentTypeID)
}

func (m *DocumentRepository) DeleteByEntryID(ctx context.Context, entryID string) error {
	return m.DeleteByEntryIDFn(ctx, entryID)
}

func (m *DocumentRepository) DeletePublishedByEntryID(ctx context.Context, entryID string) error {
	return m.DeletePublishedByEntryIDFn(ctx, entryID)
}

func (m *DocumentRepository) DeleteByContentType(ctx context.Context, contentTypeID string) error {
	return m.DeleteByContentTypeFn(ctx, contentTypeID)
}
