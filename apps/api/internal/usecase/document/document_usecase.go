package document

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

type UseCase struct {
	repo             repository.DocumentRepository
	mediaRepo        repository.MediaAssetRepository
	supportedLocales []string
}

func New(repo repository.DocumentRepository, mediaRepo repository.MediaAssetRepository, supportedLocales []string) *UseCase {
	return &UseCase{repo: repo, mediaRepo: mediaRepo, supportedLocales: supportedLocales}
}

func (uc *UseCase) resolveLocale(locale string) (string, error) {
	if locale == "" {
		locale = uc.supportedLocales[0]
	}
	for _, l := range uc.supportedLocales {
		if l == locale {
			return locale, nil
		}
	}
	return "", fmt.Errorf("%w: unsupported locale %q", pkgerrors.ErrValidation, locale)
}

func Status(draft, published *entity.Document) string {
	if published == nil {
		return "draft"
	}
	if draft.UpdatedAt.After(published.UpdatedAt) {
		return "modified"
	}
	return "published"
}

func (uc *UseCase) Save(ctx context.Context, contentTypeSlug string, doc *entity.Document, userID string) (*entity.Document, error) {
	locale, err := uc.resolveLocale(doc.Locale)
	if err != nil {
		return nil, err
	}
	doc.Locale = locale

	if doc.DocumentID == "" {
		doc.DocumentID = primitive.NewObjectID().Hex()
	}
	existing, err := uc.repo.FindDraftByDocumentID(ctx, contentTypeSlug, doc.DocumentID, doc.Locale)
	if err != nil && !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		return nil, err
	}

	now := time.Now().UTC()
	doc.Version = entity.VersionDraft
	doc.UpdatedAt = now
	doc.UpdatedBy = userID
	if existing != nil {
		doc.CreatedAt = existing.CreatedAt
		doc.CreatedBy = existing.CreatedBy
		if doc.ContentTypeID == "" {
			doc.ContentTypeID = existing.ContentTypeID
		}
	} else {
		doc.CreatedAt = now
		doc.CreatedBy = userID
	}

	if err := uc.repo.UpsertDraft(ctx, contentTypeSlug, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (uc *UseCase) GetForEdit(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, string, error) {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return nil, "", err
	}
	draft, err := uc.repo.FindDraftByDocumentID(ctx, contentTypeSlug, documentID, locale)
	if err != nil {
		return nil, "", err
	}
	published, err := uc.repo.FindPublishedByDocumentID(ctx, contentTypeSlug, documentID, locale)
	if err != nil {
		if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return nil, "", err
		}
		published = nil
	}
	return draft, Status(draft, published), nil
}

func (uc *UseCase) GetPublished(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error) {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return nil, err
	}
	return uc.repo.FindPublishedByDocumentID(ctx, contentTypeSlug, documentID, locale)
}

func (uc *UseCase) GetAll(ctx context.Context, contentTypeSlug string) ([]*entity.Document, error) {
	return uc.repo.FindDraftsByContentType(ctx, contentTypeSlug)
}

func (uc *UseCase) Publish(ctx context.Context, contentTypeSlug, documentID, locale, userID string) error {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return err
	}
	draft, err := uc.repo.FindDraftByDocumentID(ctx, contentTypeSlug, documentID, locale)
	if err != nil {
		return err
	}
	published := &entity.Document{
		DocumentID:    draft.DocumentID,
		ContentTypeID: draft.ContentTypeID,
		Data:          draft.Data,
		Locale:        draft.Locale,
		CreatedAt:     draft.CreatedAt,
		CreatedBy:     draft.CreatedBy,
		UpdatedAt:     draft.UpdatedAt,
		UpdatedBy:     draft.UpdatedBy,
		PublishedAt:   time.Now().UTC(),
		PublishedBy:   userID,
	}
	return uc.repo.UpsertPublished(ctx, contentTypeSlug, published)
}

func (uc *UseCase) Unpublish(ctx context.Context, contentTypeSlug, documentID, locale string) error {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return err
	}
	return uc.repo.DeletePublishedByDocumentID(ctx, contentTypeSlug, documentID, locale)
}

func (uc *UseCase) Delete(ctx context.Context, contentTypeSlug, documentID string) error {
	if err := uc.mediaRepo.DeleteByDocumentRef(ctx, documentID); err != nil {
		return err
	}
	for _, locale := range uc.supportedLocales {
		if err := uc.repo.DeleteByDocumentID(ctx, contentTypeSlug, documentID, locale); err != nil {
			return err
		}
	}
	return nil
}
