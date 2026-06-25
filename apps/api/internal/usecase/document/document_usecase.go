package document

import (
	"context"
	"encoding/json"
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

func componentFields(fields []entity.FieldDefinition) []entity.FieldDefinition {
	var result []entity.FieldDefinition
	for _, f := range fields {
		if f.Type == "component" {
			result = append(result, f)
		}
	}
	return result
}

func sanitizeFields(data map[string]any, fields []entity.FieldDefinition) {
	for _, field := range fields {
		val, exists := data[field.Name]
		if !exists || val == nil {
			continue
		}
		if field.Type == "component" {
			if field.Repeatable {
				arr, ok := val.([]any)
				if !ok {
					continue
				}
				for _, item := range arr {
					compMap, ok := item.(map[string]any)
					if ok {
						sanitizeFields(compMap, field.Fields)
					}
				}
			} else {
				compMap, ok := val.(map[string]any)
				if ok {
					sanitizeFields(compMap, field.Fields)
				}
			}
			continue
		}
		str, ok := val.(string)
		if ok && str == "" && field.Type != "text" && field.Type != "richtext" && field.Type != "json" {
			data[field.Name] = nil
		}
	}
}

// --- Save flow (top-down) ---

func (uc *UseCase) extractAndSaveComponents(ctx context.Context, slug string, doc *entity.Document, fields []entity.FieldDefinition) error {
	if uc.compRepo == nil {
		return nil
	}
	if err := uc.cleanupNestedComponents(ctx, slug, doc.DocumentID, doc.Locale, doc.Version, fields); err != nil {
		return err
	}
	return uc.saveTopLevelComponents(ctx, slug, doc.DocumentID, doc.Locale, doc.Version, doc.Fields, fields)
}

func (uc *UseCase) cleanupNestedComponents(ctx context.Context, slug, documentID, locale string, version entity.DocumentVersion, fields []entity.FieldDefinition) error {
	for _, field := range componentFields(fields) {
		oldParents, err := uc.compRepo.FindByDocumentID(ctx, slug, field.Name, documentID, locale, version)
		if err != nil {
			return err
		}
		for _, parent := range oldParents {
			if err := uc.cleanupChildren(ctx, slug, field.Name, parent.ComponentID, locale, version, field.Fields); err != nil {
				return err
			}
		}
	}
	return nil
}

func (uc *UseCase) cleanupChildren(ctx context.Context, slug, parentPath, parentComponentID, locale string, version entity.DocumentVersion, childFields []entity.FieldDefinition) error {
	for _, field := range componentFields(childFields) {
		path := parentPath + "_" + field.Name
		oldChildren, err := uc.compRepo.FindByParentComponentID(ctx, slug, path, parentComponentID, locale, version)
		if err != nil {
			return err
		}
		for _, child := range oldChildren {
			if err := uc.cleanupChildren(ctx, slug, path, child.ComponentID, locale, version, field.Fields); err != nil {
				return err
			}
		}
		if err := uc.compRepo.DeleteByParentComponentID(ctx, slug, path, parentComponentID, locale); err != nil {
			return err
		}
	}
	return nil
}

func (uc *UseCase) saveTopLevelComponents(ctx context.Context, slug, documentID, locale string, version entity.DocumentVersion, data map[string]any, fields []entity.FieldDefinition) error {
	now := time.Now().UTC()
	for _, field := range componentFields(fields) {
		raw, ok := data[field.Name]
		if !ok || raw == nil {
			continue
		}
		delete(data, field.Name)

		var components []*entity.Component
		if field.Repeatable {
			arr, ok := raw.([]any)
			if !ok {
				return fmt.Errorf("%w: field %q expects an array of components", pkgerrors.ErrValidation, field.Name)
			}
			for idx, item := range arr {
				compMap, ok := item.(map[string]any)
				if !ok {
					continue
				}
				comp := &entity.Component{
					ComponentID: uuid.New().String(),
					SortOrder:   idx,
					Fields:      compMap,
					CreatedAt:   now,
					UpdatedAt:   now,
				}
				components = append(components, comp)
			}
		} else {
			compMap, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("%w: field %q expects a single component, not an array", pkgerrors.ErrValidation, field.Name)
			}
			components = append(components, &entity.Component{
				ComponentID: uuid.New().String(),
				SortOrder:   0,
				Fields:      compMap,
				CreatedAt:   now,
				UpdatedAt:   now,
			})
		}

		for _, comp := range components {
			if err := uc.saveNestedComponents(ctx, slug, field.Name, comp.ComponentID, locale, version, comp.Fields, field.Fields); err != nil {
				return err
			}
		}

		if err := uc.compRepo.UpsertAll(ctx, slug, field.Name, documentID, locale, version, components); err != nil {
			return err
		}
	}
	return nil
}

func (uc *UseCase) saveNestedComponents(ctx context.Context, slug, parentPath, parentComponentID, locale string, version entity.DocumentVersion, data map[string]any, childFields []entity.FieldDefinition) error {
	now := time.Now().UTC()
	for _, field := range componentFields(childFields) {
		raw, ok := data[field.Name]
		if !ok || raw == nil {
			continue
		}
		delete(data, field.Name)

		path := parentPath + "_" + field.Name

		var components []*entity.Component
		if field.Repeatable {
			arr, ok := raw.([]any)
			if !ok {
				return fmt.Errorf("%w: field %q expects an array of components", pkgerrors.ErrValidation, field.Name)
			}
			for idx, item := range arr {
				compMap, ok := item.(map[string]any)
				if !ok {
					continue
				}
				comp := &entity.Component{
					ComponentID: uuid.New().String(),
					SortOrder:   idx,
					Fields:      compMap,
					CreatedAt:   now,
					UpdatedAt:   now,
				}
				components = append(components, comp)
			}
		} else {
			compMap, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("%w: field %q expects a single component, not an array", pkgerrors.ErrValidation, field.Name)
			}
			components = append(components, &entity.Component{
				ComponentID: uuid.New().String(),
				SortOrder:   0,
				Fields:      compMap,
				CreatedAt:   now,
				UpdatedAt:   now,
			})
		}

		for _, comp := range components {
			if err := uc.saveNestedComponents(ctx, slug, path, comp.ComponentID, locale, version, comp.Fields, field.Fields); err != nil {
				return err
			}
		}

		if err := uc.compRepo.UpsertAllByParent(ctx, slug, path, parentComponentID, locale, version, components); err != nil {
			return err
		}
	}
	return nil
}

// --- Read/Merge flow (chain traversal) ---

func (uc *UseCase) mergeComponents(ctx context.Context, slug string, doc *entity.Document, fields []entity.FieldDefinition) error {
	if uc.compRepo == nil {
		return nil
	}
	return uc.mergeTopLevel(ctx, slug, doc.DocumentID, doc.Locale, doc.Version, doc.Fields, fields)
}

func (uc *UseCase) mergeTopLevel(ctx context.Context, slug, documentID, locale string, version entity.DocumentVersion, data map[string]any, fields []entity.FieldDefinition) error {
	for _, field := range componentFields(fields) {
		components, err := uc.compRepo.FindByDocumentID(ctx, slug, field.Name, documentID, locale, version)
		if err != nil {
			return err
		}
		if field.Repeatable {
			arr := make([]map[string]any, len(components))
			for idx, comp := range components {
				if err := uc.mergeNested(ctx, slug, field.Name, comp.ComponentID, locale, version, comp.Fields, field.Fields); err != nil {
					return err
				}
				arr[idx] = comp.Fields
			}
			data[field.Name] = arr
		} else {
			if len(components) >= 1 {
				if err := uc.mergeNested(ctx, slug, field.Name, components[0].ComponentID, locale, version, components[0].Fields, field.Fields); err != nil {
					return err
				}
				data[field.Name] = components[0].Fields
			}
		}
	}
	return nil
}

func (uc *UseCase) mergeNested(ctx context.Context, slug, parentPath, parentComponentID, locale string, version entity.DocumentVersion, parentData map[string]any, childFields []entity.FieldDefinition) error {
	for _, field := range componentFields(childFields) {
		path := parentPath + "_" + field.Name
		children, err := uc.compRepo.FindByParentComponentID(ctx, slug, path, parentComponentID, locale, version)
		if err != nil {
			return err
		}
		if field.Repeatable {
			arr := make([]map[string]any, len(children))
			for idx, child := range children {
				if err := uc.mergeNested(ctx, slug, path, child.ComponentID, locale, version, child.Fields, field.Fields); err != nil {
					return err
				}
				arr[idx] = child.Fields
			}
			parentData[field.Name] = arr
		} else {
			if len(children) >= 1 {
				if err := uc.mergeNested(ctx, slug, path, children[0].ComponentID, locale, version, children[0].Fields, field.Fields); err != nil {
					return err
				}
				parentData[field.Name] = children[0].Fields
			}
		}
	}
	return nil
}

// --- Publish flow (chain traversal) ---

func (uc *UseCase) publishComponents(ctx context.Context, slug, documentID, locale string, fields []entity.FieldDefinition) error {
	for _, field := range componentFields(fields) {
		draftParents, err := uc.compRepo.FindByDocumentID(ctx, slug, field.Name, documentID, locale, entity.VersionDraft)
		if err != nil {
			return err
		}
		if err := uc.compRepo.UpsertAll(ctx, slug, field.Name, documentID, locale, entity.VersionPublished, draftParents); err != nil {
			return err
		}
		for _, parent := range draftParents {
			if err := uc.publishNested(ctx, slug, field.Name, parent.ComponentID, locale, field.Fields); err != nil {
				return err
			}
		}
	}
	return nil
}

func (uc *UseCase) publishNested(ctx context.Context, slug, parentPath, parentComponentID, locale string, childFields []entity.FieldDefinition) error {
	for _, field := range componentFields(childFields) {
		path := parentPath + "_" + field.Name
		draftChildren, err := uc.compRepo.FindByParentComponentID(ctx, slug, path, parentComponentID, locale, entity.VersionDraft)
		if err != nil {
			return err
		}
		if err := uc.compRepo.UpsertAllByParent(ctx, slug, path, parentComponentID, locale, entity.VersionPublished, draftChildren); err != nil {
			return err
		}
		for _, child := range draftChildren {
			if err := uc.publishNested(ctx, slug, path, child.ComponentID, locale, field.Fields); err != nil {
				return err
			}
		}
	}
	return nil
}

// --- Unpublish flow (bottom-up chain traversal on published version) ---

func (uc *UseCase) unpublishComponents(ctx context.Context, slug, documentID, locale string, fields []entity.FieldDefinition) error {
	for _, field := range componentFields(fields) {
		parents, err := uc.compRepo.FindByDocumentID(ctx, slug, field.Name, documentID, locale, entity.VersionPublished)
		if err != nil {
			return err
		}
		for _, parent := range parents {
			if err := uc.unpublishChildren(ctx, slug, field.Name, parent.ComponentID, locale, field.Fields); err != nil {
				return err
			}
		}
		if err := uc.compRepo.UpsertAll(ctx, slug, field.Name, documentID, locale, entity.VersionPublished, nil); err != nil {
			return err
		}
	}
	return nil
}

func (uc *UseCase) unpublishChildren(ctx context.Context, slug, parentPath, parentComponentID, locale string, childFields []entity.FieldDefinition) error {
	for _, field := range componentFields(childFields) {
		path := parentPath + "_" + field.Name
		children, err := uc.compRepo.FindByParentComponentID(ctx, slug, path, parentComponentID, locale, entity.VersionPublished)
		if err != nil {
			return err
		}
		for _, child := range children {
			if err := uc.unpublishChildren(ctx, slug, path, child.ComponentID, locale, field.Fields); err != nil {
				return err
			}
		}
		if err := uc.compRepo.UpsertAllByParent(ctx, slug, path, parentComponentID, locale, entity.VersionPublished, nil); err != nil {
			return err
		}
	}
	return nil
}

// --- Delete flow (bottom-up chain traversal, both versions) ---

func (uc *UseCase) deleteComponents(ctx context.Context, slug, documentID string, fields []entity.FieldDefinition) error {
	for _, locale := range uc.supportedLocales {
		for _, field := range componentFields(fields) {
			uniqueIDs, err := uc.collectParentIDs(ctx, slug, field.Name, documentID, locale)
			if err != nil {
				return err
			}
			for _, compID := range uniqueIDs {
				if err := uc.deleteChildren(ctx, slug, field.Name, compID, locale, field.Fields); err != nil {
					return err
				}
			}
			if err := uc.compRepo.DeleteByDocumentID(ctx, slug, field.Name, documentID, locale); err != nil {
				return err
			}
		}
	}
	return nil
}

func (uc *UseCase) collectParentIDs(ctx context.Context, slug, componentPath, documentID, locale string) ([]string, error) {
	draftParents, err := uc.compRepo.FindByDocumentID(ctx, slug, componentPath, documentID, locale, entity.VersionDraft)
	if err != nil {
		return nil, err
	}
	pubParents, err := uc.compRepo.FindByDocumentID(ctx, slug, componentPath, documentID, locale, entity.VersionPublished)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var ids []string
	for _, comp := range draftParents {
		if !seen[comp.ComponentID] {
			seen[comp.ComponentID] = true
			ids = append(ids, comp.ComponentID)
		}
	}
	for _, comp := range pubParents {
		if !seen[comp.ComponentID] {
			seen[comp.ComponentID] = true
			ids = append(ids, comp.ComponentID)
		}
	}
	return ids, nil
}

func (uc *UseCase) collectChildIDs(ctx context.Context, slug, componentPath, parentComponentID, locale string) ([]string, error) {
	draftChildren, err := uc.compRepo.FindByParentComponentID(ctx, slug, componentPath, parentComponentID, locale, entity.VersionDraft)
	if err != nil {
		return nil, err
	}
	pubChildren, err := uc.compRepo.FindByParentComponentID(ctx, slug, componentPath, parentComponentID, locale, entity.VersionPublished)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var ids []string
	for _, comp := range draftChildren {
		if !seen[comp.ComponentID] {
			seen[comp.ComponentID] = true
			ids = append(ids, comp.ComponentID)
		}
	}
	for _, comp := range pubChildren {
		if !seen[comp.ComponentID] {
			seen[comp.ComponentID] = true
			ids = append(ids, comp.ComponentID)
		}
	}
	return ids, nil
}

func (uc *UseCase) deleteChildren(ctx context.Context, slug, parentPath, parentComponentID, locale string, childFields []entity.FieldDefinition) error {
	for _, field := range componentFields(childFields) {
		path := parentPath + "_" + field.Name
		childIDs, err := uc.collectChildIDs(ctx, slug, path, parentComponentID, locale)
		if err != nil {
			return err
		}
		for _, childID := range childIDs {
			if err := uc.deleteChildren(ctx, slug, path, childID, locale, field.Fields); err != nil {
				return err
			}
		}
		if err := uc.compRepo.DeleteByParentComponentID(ctx, slug, path, parentComponentID, locale); err != nil {
			return err
		}
	}
	return nil
}

func (uc *UseCase) resolveMediaFields(ctx context.Context, data map[string]any, fields []entity.FieldDefinition) {
	for _, f := range fields {
		raw, ok := data[f.Name]
		if !ok || raw == nil {
			continue
		}

		if f.Type == "media" {
			if docID, ok := raw.(string); ok && docID != "" {
				asset, err := uc.mediaRepo.FindByDocumentID(ctx, docID)
				if err == nil && asset != nil {
					data[f.Name] = asset
				}
			}
		} else if f.Type == "component" {
			if arr, ok := raw.([]map[string]any); ok {
				for _, compData := range arr {
					uc.resolveMediaFields(ctx, compData, f.Fields)
				}
			} else if arr, ok := raw.([]any); ok {
				for _, item := range arr {
					if compData, ok := item.(map[string]any); ok {
						uc.resolveMediaFields(ctx, compData, f.Fields)
					}
				}
			} else if compData, ok := raw.(map[string]any); ok {
				uc.resolveMediaFields(ctx, compData, f.Fields)
			}
		}
	}
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

	sanitizeFields(doc.Fields, fields)

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
	uc.resolveMediaFields(ctx, draft.Fields, fields)
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
	uc.resolveMediaFields(ctx, doc.Fields, fields)
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
		uc.resolveMediaFields(ctx, doc.Fields, fields)
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
	uc.resolveMediaFields(ctx, doc.Fields, fields)
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
		if err := uc.publishComponents(ctx, contentTypeSlug, documentID, locale, fields); err != nil {
			return err
		}
	}
	return nil
}

func (uc *UseCase) Unpublish(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition) error {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return err
	}
	if uc.compRepo != nil {
		if err := uc.unpublishComponents(ctx, contentTypeSlug, documentID, locale, fields); err != nil {
			return err
		}
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
	uc.resolveMediaFields(ctx, draft.Fields, fields)
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

func (uc *UseCase) UnpublishSingleType(ctx context.Context, contentTypeSlug, locale string, fields []entity.FieldDefinition) error {
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
	return uc.Unpublish(ctx, contentTypeSlug, drafts[0].DocumentID, locale, fields)
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
		uc.resolveMediaFields(ctx, draft.Fields, fields)
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

func (uc *UseCase) Duplicate(ctx context.Context, contentTypeSlug, sourceDocumentID, locale string, fields []entity.FieldDefinition, userID string) (*entity.Document, error) {
	locale, err := uc.resolveLocale(locale)
	if err != nil {
		return nil, err
	}

	source, err := uc.repo.FindDraftByDocumentID(ctx, contentTypeSlug, sourceDocumentID, locale)
	if err != nil {
		return nil, err
	}

	if err := uc.mergeComponents(ctx, contentTypeSlug, source, fields); err != nil {
		return nil, err
	}

	raw, err := json.Marshal(source.Fields)
	if err != nil {
		return nil, fmt.Errorf("duplicate: marshal fields: %w", err)
	}
	var copiedFields map[string]any
	if err := json.Unmarshal(raw, &copiedFields); err != nil {
		return nil, fmt.Errorf("duplicate: unmarshal fields: %w", err)
	}

	newDoc := &entity.Document{
		Fields: copiedFields,
		Locale: source.Locale,
	}
	return uc.Save(ctx, contentTypeSlug, newDoc, fields, userID)
}

func (uc *UseCase) Delete(ctx context.Context, contentTypeSlug, documentID string, fields []entity.FieldDefinition) error {
	if uc.compRepo != nil {
		if err := uc.deleteComponents(ctx, contentTypeSlug, documentID, fields); err != nil {
			return err
		}
	}
	for _, locale := range uc.supportedLocales {
		if err := uc.repo.DeleteByDocumentID(ctx, contentTypeSlug, documentID, locale); err != nil {
			return err
		}
	}
	return nil
}
