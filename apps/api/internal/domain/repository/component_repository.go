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
	EnsureCollection(ctx context.Context, contentTypeSlug, componentName string, fields []entity.FieldDefinition) error
	DropCollection(ctx context.Context, contentTypeSlug, componentName string) error
}
