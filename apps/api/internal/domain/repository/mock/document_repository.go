package mock

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.DocumentRepository = (*DocumentRepository)(nil)

type DocumentRepository struct {
	FindDraftByDocumentIDFn     func(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error)
	FindPublishedByDocumentIDFn func(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error)
	UpsertDraftFn               func(ctx context.Context, contentTypeSlug string, doc *entity.Document) error
	UpsertPublishedFn           func(ctx context.Context, contentTypeSlug string, doc *entity.Document) error
	FindDraftsByContentTypeFn            func(ctx context.Context, contentTypeSlug string) ([]*entity.Document, error)
	FindDraftsByContentTypePaginatedFn      func(ctx context.Context, contentTypeSlug string, start, size int, locale, orderBy string, sortDir int, filters []entity.FilterNode) ([]*entity.Document, int64, error)
	FindPublishedByContentTypePaginatedFn  func(ctx context.Context, contentTypeSlug string, start, size int, locale, orderBy string, sortDir int, filters []entity.FilterNode) ([]*entity.Document, int64, error)
	FindPublishedByDocumentIDsFn           func(ctx context.Context, contentTypeSlug string, documentIDs []string, locale string) ([]*entity.Document, error)
	DeleteByDocumentIDFn                 func(ctx context.Context, contentTypeSlug, documentID, locale string) error
	DeletePublishedByDocumentIDFn func(ctx context.Context, contentTypeSlug, documentID, locale string) error
	DeleteAllByContentTypeFn    func(ctx context.Context, contentTypeSlug string) error
	EnsureCollectionFn          func(ctx context.Context, contentTypeSlug string, fields []entity.FieldDefinition) error
	DropCollectionFn            func(ctx context.Context, contentTypeSlug string) error
	TableInfoFn                 func(ctx context.Context, contentTypeSlug string) (bool, int64, error)
	CountByLocaleFn             func(ctx context.Context, contentTypeSlug, locale string) (int64, error)
}

func (m *DocumentRepository) FindDraftByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error) {
	return m.FindDraftByDocumentIDFn(ctx, contentTypeSlug, documentID, locale)
}

func (m *DocumentRepository) FindPublishedByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error) {
	return m.FindPublishedByDocumentIDFn(ctx, contentTypeSlug, documentID, locale)
}

func (m *DocumentRepository) UpsertDraft(ctx context.Context, contentTypeSlug string, doc *entity.Document) error {
	return m.UpsertDraftFn(ctx, contentTypeSlug, doc)
}

func (m *DocumentRepository) UpsertPublished(ctx context.Context, contentTypeSlug string, doc *entity.Document) error {
	return m.UpsertPublishedFn(ctx, contentTypeSlug, doc)
}

func (m *DocumentRepository) FindDraftsByContentType(ctx context.Context, contentTypeSlug string) ([]*entity.Document, error) {
	return m.FindDraftsByContentTypeFn(ctx, contentTypeSlug)
}

func (m *DocumentRepository) FindDraftsByContentTypePaginated(ctx context.Context, contentTypeSlug string, start, size int, locale, orderBy string, sortDir int, filters []entity.FilterNode) ([]*entity.Document, int64, error) {
	return m.FindDraftsByContentTypePaginatedFn(ctx, contentTypeSlug, start, size, locale, orderBy, sortDir, filters)
}

func (m *DocumentRepository) FindPublishedByContentTypePaginated(ctx context.Context, contentTypeSlug string, start, size int, locale, orderBy string, sortDir int, filters []entity.FilterNode) ([]*entity.Document, int64, error) {
	return m.FindPublishedByContentTypePaginatedFn(ctx, contentTypeSlug, start, size, locale, orderBy, sortDir, filters)
}

func (m *DocumentRepository) FindPublishedByDocumentIDs(ctx context.Context, contentTypeSlug string, documentIDs []string, locale string) ([]*entity.Document, error) {
	return m.FindPublishedByDocumentIDsFn(ctx, contentTypeSlug, documentIDs, locale)
}

func (m *DocumentRepository) DeleteByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) error {
	return m.DeleteByDocumentIDFn(ctx, contentTypeSlug, documentID, locale)
}

func (m *DocumentRepository) DeletePublishedByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) error {
	return m.DeletePublishedByDocumentIDFn(ctx, contentTypeSlug, documentID, locale)
}

func (m *DocumentRepository) DeleteAllByContentType(ctx context.Context, contentTypeSlug string) error {
	return m.DeleteAllByContentTypeFn(ctx, contentTypeSlug)
}

func (m *DocumentRepository) EnsureCollection(ctx context.Context, contentTypeSlug string, fields []entity.FieldDefinition) error {
	if m.EnsureCollectionFn != nil {
		return m.EnsureCollectionFn(ctx, contentTypeSlug, fields)
	}
	return nil
}

func (m *DocumentRepository) DropCollection(ctx context.Context, contentTypeSlug string) error {
	if m.DropCollectionFn != nil {
		return m.DropCollectionFn(ctx, contentTypeSlug)
	}
	return nil
}

func (m *DocumentRepository) TableInfo(ctx context.Context, contentTypeSlug string) (bool, int64, error) {
	if m.TableInfoFn != nil {
		return m.TableInfoFn(ctx, contentTypeSlug)
	}
	return false, 0, nil
}

func (m *DocumentRepository) CountByLocale(ctx context.Context, contentTypeSlug, locale string) (int64, error) {
	if m.CountByLocaleFn != nil {
		return m.CountByLocaleFn(ctx, contentTypeSlug, locale)
	}
	return 0, nil
}
