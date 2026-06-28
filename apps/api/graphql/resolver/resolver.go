package resolver

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

type DocumentUseCase interface {
	Save(ctx context.Context, contentTypeSlug string, doc *entity.Document, fields []entity.FieldDefinition, userID string) (*entity.Document, error)
	GetForEdit(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition) (*entity.Document, string, error)
	GetPublished(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition) (*entity.Document, error)
	Publish(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition, userID string) error
	Unpublish(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition) error
	Delete(ctx context.Context, contentTypeSlug, documentID string, fields []entity.FieldDefinition) error
	GetSingleType(ctx context.Context, contentTypeSlug, locale string, fields []entity.FieldDefinition) (*entity.Document, string, error)
	SaveSingleType(ctx context.Context, contentTypeSlug string, data map[string]any, locale string, fields []entity.FieldDefinition, userID string) (*entity.Document, error)
	PublishSingleType(ctx context.Context, contentTypeSlug, locale string, fields []entity.FieldDefinition, userID string) error
	UnpublishSingleType(ctx context.Context, contentTypeSlug, locale string, fields []entity.FieldDefinition) error
	GetAllPaginated(ctx context.Context, contentTypeSlug string, start, size int, locale string, fields []entity.FieldDefinition, orderBy string, sortDir int, filters []entity.FilterNode) ([]*entity.Document, []string, int64, error)
	GetPublishedPaginated(ctx context.Context, contentTypeSlug string, start, size int, locale string, fields []entity.FieldDefinition, orderBy string, sortDir int, filters []entity.FilterNode) ([]*entity.Document, int64, error)
	GetPublishedSingleType(ctx context.Context, contentTypeSlug, locale string, fields []entity.FieldDefinition) (*entity.Document, error)
}

type ContentTypeUseCase interface {
	FindAll(ctx context.Context) ([]*entity.ContentType, error)
}

type Resolver struct {
	docUC     DocumentUseCase
	ctUC      ContentTypeUseCase
	mediaRepo repository.MediaAssetRepository
}

func NewResolver(docUC DocumentUseCase, ctUC ContentTypeUseCase, mediaRepo repository.MediaAssetRepository) *Resolver {
	return &Resolver{docUC: docUC, ctUC: ctUC, mediaRepo: mediaRepo}
}
