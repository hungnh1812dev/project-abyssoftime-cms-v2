package gormdb

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var _ repository.DocumentRepository = (*documentRepository)(nil)

type documentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) repository.DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) findOne(ctx context.Context, slug, documentID, locale string, version entity.DocumentVersion) (*entity.Document, error) {
	var doc entity.Document
	err := r.db.WithContext(ctx).
		Where("slug = ? AND document_id = ? AND version = ? AND locale = ?", slug, documentID, version, locale).
		First(&doc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}
	return &doc, nil
}

func (r *documentRepository) FindDraftByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error) {
	return r.findOne(ctx, contentTypeSlug, documentID, locale, entity.VersionDraft)
}

func (r *documentRepository) FindPublishedByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error) {
	return r.findOne(ctx, contentTypeSlug, documentID, locale, entity.VersionPublished)
}

func (r *documentRepository) upsert(ctx context.Context, contentTypeSlug string, doc *entity.Document) error {
	doc.Slug = contentTypeSlug
	var existing entity.Document
	err := r.db.WithContext(ctx).
		Where("slug = ? AND document_id = ? AND version = ? AND locale = ?",
			contentTypeSlug, doc.DocumentID, doc.Version, doc.Locale).
		First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.db.WithContext(ctx).Create(doc).Error
		}
		return err
	}
	doc.GormID = existing.GormID
	return r.db.WithContext(ctx).Save(doc).Error
}

func (r *documentRepository) UpsertDraft(ctx context.Context, contentTypeSlug string, doc *entity.Document) error {
	doc.Version = entity.VersionDraft
	return r.upsert(ctx, contentTypeSlug, doc)
}

func (r *documentRepository) UpsertPublished(ctx context.Context, contentTypeSlug string, doc *entity.Document) error {
	doc.Version = entity.VersionPublished
	return r.upsert(ctx, contentTypeSlug, doc)
}

func (r *documentRepository) FindDraftsByContentType(ctx context.Context, contentTypeSlug string) ([]*entity.Document, error) {
	var docs []*entity.Document
	err := r.db.WithContext(ctx).
		Where("slug = ? AND version = ?", contentTypeSlug, entity.VersionDraft).
		Order("created_at DESC").
		Find(&docs).Error
	return docs, err
}

func (r *documentRepository) FindDraftsByContentTypePaginated(ctx context.Context, contentTypeSlug string, start, size int, locale string) ([]*entity.Document, int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&entity.Document{}).
		Where("slug = ? AND version = ? AND locale = ?", contentTypeSlug, entity.VersionDraft, locale)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var docs []*entity.Document
	if err := q.Order("created_at DESC").Offset(start).Limit(size).Find(&docs).Error; err != nil {
		return nil, 0, err
	}
	return docs, total, nil
}

func (r *documentRepository) FindPublishedByDocumentIDs(ctx context.Context, contentTypeSlug string, documentIDs []string, locale string) ([]*entity.Document, error) {
	var docs []*entity.Document
	err := r.db.WithContext(ctx).
		Where("slug = ? AND version = ? AND locale = ? AND document_id IN ?",
			contentTypeSlug, entity.VersionPublished, locale, documentIDs).
		Find(&docs).Error
	return docs, err
}

func (r *documentRepository) DeleteByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) error {
	return r.db.WithContext(ctx).
		Where("slug = ? AND document_id = ? AND locale = ?", contentTypeSlug, documentID, locale).
		Delete(&entity.Document{}).Error
}

func (r *documentRepository) DeletePublishedByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) error {
	return r.db.WithContext(ctx).
		Where("slug = ? AND document_id = ? AND version = ? AND locale = ?",
			contentTypeSlug, documentID, entity.VersionPublished, locale).
		Delete(&entity.Document{}).Error
}

func (r *documentRepository) DeleteAllByContentType(ctx context.Context, contentTypeSlug string) error {
	return r.db.WithContext(ctx).
		Where("slug = ?", contentTypeSlug).
		Delete(&entity.Document{}).Error
}

func (r *documentRepository) EnsureCollection(_ context.Context, _ string) error {
	return nil
}

func (r *documentRepository) DropCollection(_ context.Context, _ string) error {
	return nil
}
