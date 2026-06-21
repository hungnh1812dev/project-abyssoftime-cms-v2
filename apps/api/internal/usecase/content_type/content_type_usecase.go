package content_type

import (
	"context"

	"github.com/google/uuid"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

type UseCase struct {
	repo repository.ContentTypeRepository
}

func New(repo repository.ContentTypeRepository) *UseCase {
	return &UseCase{repo: repo}
}

func validateKind(kind entity.ContentKind) error {
	if kind != entity.KindSingle && kind != entity.KindCollection {
		return pkgerrors.ErrBadRequest
	}
	return nil
}

func (uc *UseCase) Create(ctx context.Context, ct *entity.ContentType) error {
	if err := validateKind(ct.Kind); err != nil {
		return err
	}
	existing, err := uc.repo.FindBySlug(ctx, ct.Slug)
	if err != nil && !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		return err
	}
	if existing != nil {
		return pkgerrors.ErrConflict
	}
	if ct.DocumentID == "" {
		ct.DocumentID = uuid.New().String()
	}
	return uc.repo.Create(ctx, ct)
}

func (uc *UseCase) FindByID(ctx context.Context, id string) (*entity.ContentType, error) {
	return uc.repo.FindByID(ctx, id)
}

func (uc *UseCase) FindBySlug(ctx context.Context, slug string) (*entity.ContentType, error) {
	return uc.repo.FindBySlug(ctx, slug)
}

func (uc *UseCase) FindAll(ctx context.Context) ([]*entity.ContentType, error) {
	return uc.repo.FindAll(ctx)
}

func (uc *UseCase) Update(ctx context.Context, ct *entity.ContentType) error {
	if err := validateKind(ct.Kind); err != nil {
		return err
	}
	return uc.repo.Update(ctx, ct)
}

func (uc *UseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}
