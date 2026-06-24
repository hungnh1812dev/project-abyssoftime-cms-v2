package repository

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type ComponentRepository interface {
	FindByDocumentID(ctx context.Context, contentTypeSlug, componentName, documentID, locale string, version entity.DocumentVersion) ([]*entity.Component, error)
	UpsertAll(ctx context.Context, contentTypeSlug, componentName, documentID, locale string, version entity.DocumentVersion, components []*entity.Component) error
	DeleteByDocumentID(ctx context.Context, contentTypeSlug, componentName, documentID, locale string) error
	DeleteAllByContentType(ctx context.Context, contentTypeSlug, componentName string) error

	FindByParentComponentID(ctx context.Context, contentTypeSlug, componentPath, parentComponentID, locale string, version entity.DocumentVersion) ([]*entity.Component, error)
	UpsertAllByParent(ctx context.Context, contentTypeSlug, componentPath, parentComponentID, locale string, version entity.DocumentVersion, components []*entity.Component) error
	DeleteByParentComponentID(ctx context.Context, contentTypeSlug, componentPath, parentComponentID, locale string) error

	EnsureCollection(ctx context.Context, contentTypeSlug, componentName string, fields []entity.FieldDefinition, isNested bool) error
	DropCollection(ctx context.Context, contentTypeSlug, componentName string) error
}
