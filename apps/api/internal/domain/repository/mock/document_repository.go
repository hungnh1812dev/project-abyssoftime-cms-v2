package mock

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.DocumentRepository = (*DocumentRepository)(nil)

// DocumentRepository is a test double for repository.DocumentRepository.
// Set each Fn field to a stub before calling the method under test.
type DocumentRepository struct {
	CreateFn            func(ctx context.Context, doc *entity.Document) error
	FindByIDFn          func(ctx context.Context, id string) (*entity.Document, error)
	FindByContentTypeFn func(ctx context.Context, contentTypeID string) ([]*entity.Document, error)
	UpdateFn            func(ctx context.Context, doc *entity.Document) error
	UpdateStatusFn      func(ctx context.Context, id string, status entity.DocumentStatus) error
	DeleteFn            func(ctx context.Context, id string) error
}

func (m *DocumentRepository) Create(ctx context.Context, doc *entity.Document) error {
	return m.CreateFn(ctx, doc)
}

func (m *DocumentRepository) FindByID(ctx context.Context, id string) (*entity.Document, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *DocumentRepository) FindByContentType(ctx context.Context, contentTypeID string) ([]*entity.Document, error) {
	return m.FindByContentTypeFn(ctx, contentTypeID)
}

func (m *DocumentRepository) Update(ctx context.Context, doc *entity.Document) error {
	return m.UpdateFn(ctx, doc)
}

func (m *DocumentRepository) UpdateStatus(ctx context.Context, id string, status entity.DocumentStatus) error {
	return m.UpdateStatusFn(ctx, id, status)
}

func (m *DocumentRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}
