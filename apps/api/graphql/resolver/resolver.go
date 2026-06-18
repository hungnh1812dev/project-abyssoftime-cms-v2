package resolver

import (
	"context"
	"net/http"
	"time"

	"project-abyssoftime-cms-v2/api/graphql/model"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type documentUseCase interface {
	Save(ctx context.Context, contentTypeSlug string, doc *entity.Document, userID string) (*entity.Document, error)
	GetPublished(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error)
	GetForEdit(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, string, error)
	Publish(ctx context.Context, contentTypeSlug, documentID, locale, userID string) error
	Unpublish(ctx context.Context, contentTypeSlug, documentID, locale string) error
	Delete(ctx context.Context, contentTypeSlug, documentID string) error
}

type contentTypeUseCase interface {
	FindAll(ctx context.Context) ([]*entity.ContentType, error)
}

type Resolver struct {
	DocumentUC    documentUseCase
	ContentTypeUC contentTypeUseCase
}

type ctxKey string

const RequestCtxKey ctxKey = "gql-request"

func WithRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, RequestCtxKey, r)
}

func toModelDocument(d *entity.Document) *model.Document {
	m := &model.Document{
		DocumentID:    d.DocumentID,
		Version:       string(d.Version),
		ContentTypeID: d.ContentTypeID,
		Data:          d.Data,
		Locale:        d.Locale,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
		CreatedBy:     d.CreatedBy,
		UpdatedBy:     d.UpdatedBy,
	}
	if !d.PublishedAt.Equal(time.Time{}) {
		m.PublishedAt = &d.PublishedAt
	}
	if d.PublishedBy != "" {
		m.PublishedBy = &d.PublishedBy
	}
	return m
}
