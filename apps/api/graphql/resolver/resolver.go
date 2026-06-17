package resolver

import (
	"context"
	"net/http"
	"time"

	"project-abyssoftime-cms-v2/api/graphql/model"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type documentUseCase interface {
	Save(ctx context.Context, doc *entity.Document, userID string) (*entity.Document, error)
	GetPublished(ctx context.Context, entryID, locale string) (*entity.Document, error)
	GetForEdit(ctx context.Context, entryID, locale string) (*entity.Document, string, error)
	Publish(ctx context.Context, entryID, locale, userID string) error
	Unpublish(ctx context.Context, entryID, locale string) error
	Delete(ctx context.Context, entryID string) error
}

type contentTypeUseCase interface {
	FindAll(ctx context.Context) ([]*entity.ContentType, error)
}

type Resolver struct {
	DocumentUC    documentUseCase
	ContentTypeUC contentTypeUseCase
}

type ctxKey string

// RequestCtxKey is the context key for the incoming *http.Request, injected
// by the GraphQL handler middleware so the auth directive can read headers.
const RequestCtxKey ctxKey = "gql-request"

// WithRequest injects r into ctx for auth directive access.
func WithRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, RequestCtxKey, r)
}

func toModelDocument(d *entity.Document) *model.Document {
	m := &model.Document{
		ID:            d.ID,
		EntryID:       d.EntryID,
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
