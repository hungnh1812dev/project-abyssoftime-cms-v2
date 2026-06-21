package mock

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.ComponentRepository = (*ComponentRepository)(nil)

type ComponentRepository struct {
	FindByDocumentIDFn     func(ctx context.Context, contentTypeSlug, componentName, documentID, locale string, version entity.DocumentVersion) ([]*entity.Component, error)
	UpsertAllFn            func(ctx context.Context, contentTypeSlug, componentName, documentID, locale string, version entity.DocumentVersion, components []*entity.Component) error
	DeleteByDocumentIDFn   func(ctx context.Context, contentTypeSlug, componentName, documentID, locale string) error
	DeleteAllByContentTypeFn func(ctx context.Context, contentTypeSlug, componentName string) error
	EnsureCollectionFn     func(ctx context.Context, contentTypeSlug, componentName string, fields []entity.FieldDefinition) error
	DropCollectionFn       func(ctx context.Context, contentTypeSlug, componentName string) error
}

func (m *ComponentRepository) FindByDocumentID(ctx context.Context, contentTypeSlug, componentName, documentID, locale string, version entity.DocumentVersion) ([]*entity.Component, error) {
	if m.FindByDocumentIDFn != nil {
		return m.FindByDocumentIDFn(ctx, contentTypeSlug, componentName, documentID, locale, version)
	}
	return nil, nil
}

func (m *ComponentRepository) UpsertAll(ctx context.Context, contentTypeSlug, componentName, documentID, locale string, version entity.DocumentVersion, components []*entity.Component) error {
	if m.UpsertAllFn != nil {
		return m.UpsertAllFn(ctx, contentTypeSlug, componentName, documentID, locale, version, components)
	}
	return nil
}

func (m *ComponentRepository) DeleteByDocumentID(ctx context.Context, contentTypeSlug, componentName, documentID, locale string) error {
	if m.DeleteByDocumentIDFn != nil {
		return m.DeleteByDocumentIDFn(ctx, contentTypeSlug, componentName, documentID, locale)
	}
	return nil
}

func (m *ComponentRepository) DeleteAllByContentType(ctx context.Context, contentTypeSlug, componentName string) error {
	if m.DeleteAllByContentTypeFn != nil {
		return m.DeleteAllByContentTypeFn(ctx, contentTypeSlug, componentName)
	}
	return nil
}

func (m *ComponentRepository) EnsureCollection(ctx context.Context, contentTypeSlug, componentName string, fields []entity.FieldDefinition) error {
	if m.EnsureCollectionFn != nil {
		return m.EnsureCollectionFn(ctx, contentTypeSlug, componentName, fields)
	}
	return nil
}

func (m *ComponentRepository) DropCollection(ctx context.Context, contentTypeSlug, componentName string) error {
	if m.DropCollectionFn != nil {
		return m.DropCollectionFn(ctx, contentTypeSlug, componentName)
	}
	return nil
}
