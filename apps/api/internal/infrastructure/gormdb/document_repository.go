package gormdb

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var gormSortColumn = map[string]string{
	"id":        "gorm_id",
	"createdAt": "created_at",
	"updatedAt": "updated_at",
}

func resolveGormSortClause(orderBy string, sortDir int) string {
	col, ok := gormSortColumn[orderBy]
	if !ok {
		col = "created_at"
	}
	dir := "DESC"
	if sortDir == 1 {
		dir = "ASC"
	}
	return fmt.Sprintf("%s %s", col, dir)
}

var _ repository.DocumentRepository = (*documentRepository)(nil)

type documentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) repository.DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) table(slug string) *gorm.DB {
	return r.db.Table(documentTableName(slug))
}

func (r *documentRepository) findOne(ctx context.Context, slug, documentID, locale string, version entity.DocumentVersion) (*entity.Document, error) {
	var doc entity.Document
	err := r.table(slug).WithContext(ctx).
		Where("document_id = ? AND version = ? AND locale = ?", documentID, version, locale).
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
	var existing entity.Document
	err := r.table(contentTypeSlug).WithContext(ctx).
		Where("document_id = ? AND version = ? AND locale = ?",
			doc.DocumentID, doc.Version, doc.Locale).
		First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.table(contentTypeSlug).WithContext(ctx).Create(doc).Error
		}
		return err
	}
	doc.GormID = existing.GormID
	return r.table(contentTypeSlug).WithContext(ctx).Save(doc).Error
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
	err := r.table(contentTypeSlug).WithContext(ctx).
		Where("version = ?", entity.VersionDraft).
		Order("created_at DESC").
		Find(&docs).Error
	return docs, err
}

func (r *documentRepository) FindDraftsByContentTypePaginated(ctx context.Context, contentTypeSlug string, start, size int, locale, orderBy string, sortDir int) ([]*entity.Document, int64, error) {
	var total int64
	q := r.table(contentTypeSlug).WithContext(ctx).
		Where("version = ? AND locale = ?", entity.VersionDraft, locale)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var docs []*entity.Document
	if err := q.Order(resolveGormSortClause(orderBy, sortDir)).Offset(start).Limit(size).Find(&docs).Error; err != nil {
		return nil, 0, err
	}
	return docs, total, nil
}

func (r *documentRepository) FindPublishedByContentTypePaginated(ctx context.Context, contentTypeSlug string, start, size int, locale, orderBy string, sortDir int) ([]*entity.Document, int64, error) {
	var total int64
	q := r.table(contentTypeSlug).WithContext(ctx).
		Where("version = ? AND locale = ?", entity.VersionPublished, locale)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var docs []*entity.Document
	if err := q.Order(resolveGormSortClause(orderBy, sortDir)).Offset(start).Limit(size).Find(&docs).Error; err != nil {
		return nil, 0, err
	}
	return docs, total, nil
}

func (r *documentRepository) FindPublishedByDocumentIDs(ctx context.Context, contentTypeSlug string, documentIDs []string, locale string) ([]*entity.Document, error) {
	var docs []*entity.Document
	err := r.table(contentTypeSlug).WithContext(ctx).
		Where("version = ? AND locale = ? AND document_id IN ?",
			entity.VersionPublished, locale, documentIDs).
		Find(&docs).Error
	return docs, err
}

func (r *documentRepository) DeleteByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) error {
	return r.table(contentTypeSlug).WithContext(ctx).
		Where("document_id = ? AND locale = ?", documentID, locale).
		Delete(&entity.Document{}).Error
}

func (r *documentRepository) DeletePublishedByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) error {
	return r.table(contentTypeSlug).WithContext(ctx).
		Where("document_id = ? AND version = ? AND locale = ?",
			documentID, entity.VersionPublished, locale).
		Delete(&entity.Document{}).Error
}

func (r *documentRepository) DeleteAllByContentType(ctx context.Context, contentTypeSlug string) error {
	return r.table(contentTypeSlug).WithContext(ctx).
		Session(&gorm.Session{AllowGlobalUpdate: true}).
		Delete(&entity.Document{}).Error
}

func (r *documentRepository) EnsureCollection(ctx context.Context, contentTypeSlug string) error {
	table := documentTableName(contentTypeSlug)
	if r.db.Migrator().HasTable(table) {
		return nil
	}
	return r.db.WithContext(ctx).Table(table).AutoMigrate(&entity.Document{})
}

func (r *documentRepository) DropCollection(ctx context.Context, contentTypeSlug string) error {
	table := documentTableName(contentTypeSlug)
	return r.db.WithContext(ctx).Migrator().DropTable(table)
}
