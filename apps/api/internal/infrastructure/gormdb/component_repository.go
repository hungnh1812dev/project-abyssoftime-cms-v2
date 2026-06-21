package gormdb

import (
	"context"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.ComponentRepository = (*componentRepository)(nil)

type componentRepository struct {
	db *gorm.DB
}

func NewComponentRepository(db *gorm.DB) repository.ComponentRepository {
	return &componentRepository{db: db}
}

func (r *componentRepository) table(slug, component string) *gorm.DB {
	return r.db.Table(componentTableName(slug, component))
}

func (r *componentRepository) FindByDocumentID(ctx context.Context, contentTypeSlug, componentName, documentID, locale string, version entity.DocumentVersion) ([]*entity.Component, error) {
	var components []*entity.Component
	err := r.table(contentTypeSlug, componentName).WithContext(ctx).
		Where("document_id = ? AND version = ? AND locale = ?", documentID, version, locale).
		Order(`"order" ASC`).
		Find(&components).Error
	return components, err
}

func (r *componentRepository) UpsertAll(ctx context.Context, contentTypeSlug, componentName, documentID, locale string, version entity.DocumentVersion, components []*entity.Component) error {
	tbl := r.table(contentTypeSlug, componentName).WithContext(ctx)

	if err := tbl.
		Where("document_id = ? AND version = ? AND locale = ?", documentID, version, locale).
		Delete(&entity.Component{}).Error; err != nil {
		return err
	}

	if len(components) == 0 {
		return nil
	}

	for i, c := range components {
		c.DocumentID = documentID
		c.Version = version
		c.Locale = locale
		c.Order = i
	}

	return r.table(contentTypeSlug, componentName).WithContext(ctx).Create(&components).Error
}

func (r *componentRepository) DeleteByDocumentID(ctx context.Context, contentTypeSlug, componentName, documentID, locale string) error {
	return r.table(contentTypeSlug, componentName).WithContext(ctx).
		Where("document_id = ? AND locale = ?", documentID, locale).
		Delete(&entity.Component{}).Error
}

func (r *componentRepository) DeleteAllByContentType(ctx context.Context, contentTypeSlug, componentName string) error {
	return r.table(contentTypeSlug, componentName).WithContext(ctx).
		Session(&gorm.Session{AllowGlobalUpdate: true}).
		Delete(&entity.Component{}).Error
}

func (r *componentRepository) EnsureCollection(ctx context.Context, contentTypeSlug, componentName string) error {
	table := componentTableName(contentTypeSlug, componentName)
	if r.db.Migrator().HasTable(table) {
		return nil
	}
	return r.db.WithContext(ctx).Table(table).AutoMigrate(&entity.Component{})
}

func (r *componentRepository) DropCollection(ctx context.Context, contentTypeSlug, componentName string) error {
	table := componentTableName(contentTypeSlug, componentName)
	return r.db.WithContext(ctx).Migrator().DropTable(table)
}
