package repository

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *entity.Document) error
	FindByID(ctx context.Context, id string) (*entity.Document, error)
	FindByContentType(ctx context.Context, contentTypeID string) ([]*entity.Document, error)
	Update(ctx context.Context, doc *entity.Document) error
	UpdateStatus(ctx context.Context, id string, status entity.DocumentStatus) error
	Delete(ctx context.Context, id string) error
}
