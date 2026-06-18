package mock

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.ContentTypeRepository = (*ContentTypeRepository)(nil)

// ContentTypeRepository is a test double for repository.ContentTypeRepository.
// Set each Fn field to a stub before calling the method under test.
type ContentTypeRepository struct {
	CreateFn      func(ctx context.Context, ct *entity.ContentType) error
	FindByIDFn    func(ctx context.Context, id string) (*entity.ContentType, error)
	FindBySlugFn  func(ctx context.Context, slug string) (*entity.ContentType, error)
	FindAllFn     func(ctx context.Context) ([]*entity.ContentType, error)
	UpdateFn      func(ctx context.Context, ct *entity.ContentType) error
	DeleteFn      func(ctx context.Context, id string) error
}

func (m *ContentTypeRepository) Create(ctx context.Context, ct *entity.ContentType) error {
	return m.CreateFn(ctx, ct)
}

func (m *ContentTypeRepository) FindByID(ctx context.Context, id string) (*entity.ContentType, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *ContentTypeRepository) FindBySlug(ctx context.Context, slug string) (*entity.ContentType, error) {
	return m.FindBySlugFn(ctx, slug)
}

func (m *ContentTypeRepository) FindAll(ctx context.Context) ([]*entity.ContentType, error) {
	return m.FindAllFn(ctx)
}

func (m *ContentTypeRepository) Update(ctx context.Context, ct *entity.ContentType) error {
	return m.UpdateFn(ctx, ct)
}

func (m *ContentTypeRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}
