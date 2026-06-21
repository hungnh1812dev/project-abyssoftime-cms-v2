package document

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

type UseCase struct {
	repo             repository.DocumentRepository
	compRepo         repository.ComponentRepository
	mediaRepo        repository.MediaAssetRepository
	supportedLocales []string
}

func New(repo repository.DocumentRepository, compRepo repository.ComponentRepository, mediaRepo repository.MediaAssetRepository, supportedLocales []string) *UseCase {
	return &UseCase{repo: repo, compRepo: compRepo, mediaRepo: mediaRepo, supportedLocales: supportedLocales}
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

func componentFieldNames(fields []entity.FieldDefinition) []string {
	var names []string
	for _, f := range fields {
		if f.Type == "component" {
			names = append(names, f.Name)
		}
	}
	return names
}

func (uc *UseCase) extractAndSaveComponents(ctx context.Context, slug string, doc *entity.Document, fields []entity.FieldDefinition) error {
	if uc.compRepo == nil {
		return nil
	}
	now := time.Now().UTC()
	for _, name := range componentFieldNames(fields) {
		raw, ok := doc.Fields[name]
		if !ok {
			continue
		}
		delete(doc.Fields, name)

		var components []*entity.Component
		switch v := raw.(type) {
		case []any:
			for _, item := range v {
				if m, ok := item.(map[string]any); ok {
					components = append(components, &entity.Component{
						ComponentID: uuid.New().String(),
						Fields:      m,
						CreatedAt:   now,
						UpdatedAt:   now,
					})
				}
			}
		case map[string]any:
			components = append(components, &entity.Component{
				ComponentID: uuid.New().String(),
				Fields:      v,
				CreatedAt:   now,
				UpdatedAt:   now,
			})
		}

		if err := uc.compRepo.UpsertAll(ctx, slug, name, doc.DocumentID, doc.Locale, doc.Version, components); err != nil {
			return err
		}
	}
	return nil
}

func (uc *UseCase) mergeComponents(ctx context.Context, slug string, doc *entity.Document, fields []entity.FieldDefinition) error {
	if uc.compRepo == nil {
		return nil
	}
	for _, name := range componentFieldNames(fields) {
		components, err := uc.compRepo.FindByDocumentID(ctx, slug, name, doc.DocumentID, doc.Locale, doc.Version)
		if err != nil {
			return err
		}
		if len(components) == 1 {
			doc.Fields[name] = components[0].Fields
		} else if len(components) > 1 {
			arr := make([]map[string]any, len(components))
			for i, c := range components {
				arr[i] = c.Fields
			}
			doc.Fields[name] = arr
		}
	}
	return nil
}

func (uc *UseCase) Save(ctx context.Context, contentTypeSlug string, doc *entity.Document, fields []entity.FieldDefinition, userID string) (*entity.Document, error) {
	locale, err := uc.resolveLocale(doc.Locale)
	if err != nil {
		return nil, err
	}
	doc.Locale = locale

	if doc.DocumentID == "" {
		doc.DocumentID = uuid.New().String()
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
	} else {
		doc.CreatedAt = now
		doc.CreatedBy = userID
	}

	if err := uc.extractAndSaveComponents(ctx, contentTypeSlug, doc, fields); err != nil {
		return nil, err
	}

	if err := uc.repo.UpsertDraft(ctx, contentTypeSlug, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (uc *UseCase) GetForEdit(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition) (*entity.Document, string, error) {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return nil, "", err
	}
	draft, err := uc.repo.FindDraftByDocumentID(ctx, contentTypeSlug, documentID, locale)
	if err != nil {
		return nil, "", err
	}
	if err := uc.mergeComponents(ctx, contentTypeSlug, draft, fields); err != nil {
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

func (uc *UseCase) GetPublished(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition) (*entity.Document, error) {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return nil, err
	}
	doc, err := uc.repo.FindPublishedByDocumentID(ctx, contentTypeSlug, documentID, locale)
	if err != nil {
		return nil, err
	}
	if err := uc.mergeComponents(ctx, contentTypeSlug, doc, fields); err != nil {
		return nil, err
	}
	return doc, nil
}

func (uc *UseCase) GetPublishedPaginated(ctx context.Context, contentTypeSlug string, start, size int, locale string, fields []entity.FieldDefinition) ([]*entity.Document, int64, error) {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return nil, 0, err
	}
	docs, total, err := uc.repo.FindPublishedByContentTypePaginated(ctx, contentTypeSlug, start, size, locale, "createdAt", -1)
	if err != nil {
		return nil, 0, err
	}
	for _, doc := range docs {
		if err := uc.mergeComponents(ctx, contentTypeSlug, doc, fields); err != nil {
			return nil, 0, err
		}
	}
	return docs, total, nil
}

func (uc *UseCase) GetPublishedSingleType(ctx context.Context, contentTypeSlug, locale string, fields []entity.FieldDefinition) (*entity.Document, error) {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return nil, err
	}
	docs, total, err := uc.repo.FindPublishedByContentTypePaginated(ctx, contentTypeSlug, 0, 1, locale, "createdAt", -1)
	if err != nil {
		return nil, err
	}
	if total == 0 || len(docs) == 0 {
		return nil, pkgerrors.ErrNotFound
	}
	doc := docs[0]
	if err := uc.mergeComponents(ctx, contentTypeSlug, doc, fields); err != nil {
		return nil, err
	}
	return doc, nil
}

func (uc *UseCase) GetAll(ctx context.Context, contentTypeSlug string) ([]*entity.Document, error) {
	return uc.repo.FindDraftsByContentType(ctx, contentTypeSlug)
}

func (uc *UseCase) Publish(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition, userID string) error {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return err
	}
	draft, err := uc.repo.FindDraftByDocumentID(ctx, contentTypeSlug, documentID, locale)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	published := &entity.Document{
		DocumentID:  draft.DocumentID,
		Fields:      draft.Fields,
		Locale:      draft.Locale,
		CreatedAt:   draft.CreatedAt,
		CreatedBy:   draft.CreatedBy,
		UpdatedAt:   draft.UpdatedAt,
		UpdatedBy:   draft.UpdatedBy,
		PublishedAt: &now,
		PublishedBy: userID,
	}
	if err := uc.repo.UpsertPublished(ctx, contentTypeSlug, published); err != nil {
		return err
	}

	if uc.compRepo != nil {
		for _, name := range componentFieldNames(fields) {
			draftComps, err := uc.compRepo.FindByDocumentID(ctx, contentTypeSlug, name, documentID, locale, entity.VersionDraft)
			if err != nil {
				return err
			}
			if err := uc.compRepo.UpsertAll(ctx, contentTypeSlug, name, documentID, locale, entity.VersionPublished, draftComps); err != nil {
				return err
			}
		}
	}
	return nil
}

func (uc *UseCase) Unpublish(ctx context.Context, contentTypeSlug, documentID, locale string) error {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return err
	}
	return uc.repo.DeletePublishedByDocumentID(ctx, contentTypeSlug, documentID, locale)
}

func (uc *UseCase) GetSingleType(ctx context.Context, contentTypeSlug, locale string, fields []entity.FieldDefinition) (*entity.Document, string, error) {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return nil, "", err
	}
	drafts, total, err := uc.repo.FindDraftsByContentTypePaginated(ctx, contentTypeSlug, 0, 1, locale, "createdAt", -1)
	if err != nil {
		return nil, "", err
	}
	if total == 0 || len(drafts) == 0 {
		return nil, "", pkgerrors.ErrNotFound
	}
	draft := drafts[0]
	if err := uc.mergeComponents(ctx, contentTypeSlug, draft, fields); err != nil {
		return nil, "", err
	}
	published, err := uc.repo.FindPublishedByDocumentID(ctx, contentTypeSlug, draft.DocumentID, locale)
	if err != nil && !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
		return nil, "", err
	}
	return draft, Status(draft, published), nil
}

func (uc *UseCase) SaveSingleType(ctx context.Context, contentTypeSlug string, data map[string]any, locale string, fields []entity.FieldDefinition, userID string) (*entity.Document, error) {
	resolvedLocale, err := uc.resolveLocale(locale)
	if err != nil {
		return nil, err
	}
	drafts, _, err := uc.repo.FindDraftsByContentTypePaginated(ctx, contentTypeSlug, 0, 1, resolvedLocale, "createdAt", -1)
	if err != nil {
		return nil, err
	}
	doc := &entity.Document{Fields: data, Locale: resolvedLocale}
	if len(drafts) > 0 {
		doc.DocumentID = drafts[0].DocumentID
	}
	return uc.Save(ctx, contentTypeSlug, doc, fields, userID)
}

func (uc *UseCase) PublishSingleType(ctx context.Context, contentTypeSlug, locale string, fields []entity.FieldDefinition, userID string) error {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return err
	}
	drafts, total, err := uc.repo.FindDraftsByContentTypePaginated(ctx, contentTypeSlug, 0, 1, locale, "createdAt", -1)
	if err != nil {
		return err
	}
	if total == 0 || len(drafts) == 0 {
		return pkgerrors.ErrNotFound
	}
	return uc.Publish(ctx, contentTypeSlug, drafts[0].DocumentID, locale, fields, userID)
}

func (uc *UseCase) UnpublishSingleType(ctx context.Context, contentTypeSlug, locale string) error {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return err
	}
	drafts, total, err := uc.repo.FindDraftsByContentTypePaginated(ctx, contentTypeSlug, 0, 1, locale, "createdAt", -1)
	if err != nil {
		return err
	}
	if total == 0 || len(drafts) == 0 {
		return pkgerrors.ErrNotFound
	}
	return uc.Unpublish(ctx, contentTypeSlug, drafts[0].DocumentID, locale)
}

func (uc *UseCase) GetAllPaginated(ctx context.Context, contentTypeSlug string, start, size int, locale string, fields []entity.FieldDefinition, orderBy string, sortDir int) ([]*entity.Document, []string, int64, error) {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return nil, nil, 0, err
	}
	drafts, total, err := uc.repo.FindDraftsByContentTypePaginated(ctx, contentTypeSlug, start, size, locale, orderBy, sortDir)
	if err != nil {
		return nil, nil, 0, err
	}
	if len(drafts) == 0 {
		return drafts, nil, total, nil
	}

	for _, draft := range drafts {
		if err := uc.mergeComponents(ctx, contentTypeSlug, draft, fields); err != nil {
			return nil, nil, 0, err
		}
	}

	ids := make([]string, len(drafts))
	for i, d := range drafts {
		ids[i] = d.DocumentID
	}
	published, err := uc.repo.FindPublishedByDocumentIDs(ctx, contentTypeSlug, ids, locale)
	if err != nil {
		return nil, nil, 0, err
	}
	pubMap := make(map[string]*entity.Document, len(published))
	for _, p := range published {
		pubMap[p.DocumentID] = p
	}
	statuses := make([]string, len(drafts))
	for i, d := range drafts {
		statuses[i] = Status(d, pubMap[d.DocumentID])
	}
	return drafts, statuses, total, nil
}

func (uc *UseCase) Delete(ctx context.Context, contentTypeSlug, documentID string, fields []entity.FieldDefinition) error {
	if uc.compRepo != nil {
		for _, name := range componentFieldNames(fields) {
			for _, locale := range uc.supportedLocales {
				if err := uc.compRepo.DeleteByDocumentID(ctx, contentTypeSlug, name, documentID, locale); err != nil {
					return err
				}
			}
		}
	}
	for _, locale := range uc.supportedLocales {
		if err := uc.repo.DeleteByDocumentID(ctx, contentTypeSlug, documentID, locale); err != nil {
			return err
		}
	}
	return nil
}
