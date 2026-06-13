package document

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

type UseCase struct {
	repo      repository.DocumentRepository
	mediaRepo repository.MediaAssetRepository
}

func New(repo repository.DocumentRepository, mediaRepo repository.MediaAssetRepository) *UseCase {
	return &UseCase{repo: repo, mediaRepo: mediaRepo}
}

func (uc *UseCase) Create(ctx context.Context, doc *entity.Document) error {
	if doc.ID == "" {
		doc.ID = primitive.NewObjectID().Hex()
	}
	if doc.Status == "" {
		doc.Status = entity.StatusDraft
	}
	return uc.repo.Create(ctx, doc)
}

func (uc *UseCase) GetOne(ctx context.Context, id string) (*entity.Document, error) {
	return uc.repo.FindByID(ctx, id)
}

func (uc *UseCase) GetAll(ctx context.Context, contentTypeID string) ([]*entity.Document, error) {
	return uc.repo.FindByContentType(ctx, contentTypeID)
}

func (uc *UseCase) Update(ctx context.Context, doc *entity.Document) error {
	return uc.repo.Update(ctx, doc)
}

func (uc *UseCase) Publish(ctx context.Context, id string) error {
	return uc.repo.UpdateStatus(ctx, id, entity.StatusPublished)
}

func (uc *UseCase) Unpublish(ctx context.Context, id string) error {
	return uc.repo.UpdateStatus(ctx, id, entity.StatusDraft)
}

func (uc *UseCase) Delete(ctx context.Context, id string) error {
	if err := uc.mediaRepo.DeleteByDocumentRef(ctx, id); err != nil {
		return err
	}
	return uc.repo.Delete(ctx, id)
}
