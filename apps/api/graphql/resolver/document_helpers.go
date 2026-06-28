package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"project-abyssoftime-cms-v2/api/graphql/model"
	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

// ── Collection-type query helpers ──

func (resolver *Resolver) getDocument(ctx context.Context, slug, docID string, locale, status *string, fields []entity.FieldDefinition) (map[string]interface{}, error) {
	localeStr := derefString(locale)
	if derefString(status) == "draft" && middleware.UserID(ctx) != "" {
		doc, _, err := resolver.docUC.GetForEdit(ctx, slug, docID, localeStr, fields)
		if err != nil {
			return nil, err
		}
		return resolver.docToMap(ctx, doc, fields), nil
	}
	doc, err := resolver.docUC.GetPublished(ctx, slug, docID, localeStr, fields)
	if err != nil {
		return nil, err
	}
	return resolver.docToMap(ctx, doc, fields), nil
}

func (resolver *Resolver) getDocumentList(ctx context.Context, slug string, pagination *model.PaginationInput, filters interface{}, orderBy interface{}, locale, status *string, fields []entity.FieldDefinition) (map[string]interface{}, error) {
	startVal, limitVal, err := parsePagination(pagination)
	if err != nil {
		return nil, err
	}
	localeStr := derefString(locale)

	orderField, sortDir := extractOrderBy(orderBy)

	var parsedFilters []entity.FilterNode
	if filters != nil {
		parsedFilters = convertFiltersByReflection(filters)
	}

	if derefString(status) == "draft" && middleware.UserID(ctx) != "" {
		docs, _, total, err := resolver.docUC.GetAllPaginated(ctx, slug, startVal, limitVal, localeStr, fields, orderField, sortDir, parsedFilters)
		if err != nil {
			return nil, err
		}
		return resolver.buildListResponse(ctx, docs, fields, total, startVal, limitVal), nil
	}

	docs, total, err := resolver.docUC.GetPublishedPaginated(ctx, slug, startVal, limitVal, localeStr, fields, orderField, sortDir, parsedFilters)
	if err != nil {
		return nil, err
	}
	return resolver.buildListResponse(ctx, docs, fields, total, startVal, limitVal), nil
}

// ── Collection-type mutation helpers ──

func (resolver *Resolver) createDocument(ctx context.Context, slug string, data interface{}, fields []entity.FieldDefinition) (map[string]interface{}, error) {
	dataMap := structToMap(data)
	doc := &entity.Document{Fields: dataMap}
	saved, err := resolver.docUC.Save(ctx, slug, doc, fields, middleware.UserID(ctx))
	if err != nil {
		return nil, err
	}
	return resolver.docToMap(ctx, saved, fields), nil
}

func (resolver *Resolver) updateDocument(ctx context.Context, slug, docID string, data interface{}, fields []entity.FieldDefinition) (map[string]interface{}, error) {
	dataMap := structToMap(data)
	doc := &entity.Document{DocumentID: docID, Fields: dataMap}
	saved, err := resolver.docUC.Save(ctx, slug, doc, fields, middleware.UserID(ctx))
	if err != nil {
		return nil, err
	}
	return resolver.docToMap(ctx, saved, fields), nil
}

func (resolver *Resolver) deleteDocument(ctx context.Context, slug, docID string, fields []entity.FieldDefinition) (bool, error) {
	return true, resolver.docUC.Delete(ctx, slug, docID, fields)
}

func (resolver *Resolver) publishDocument(ctx context.Context, slug, docID string, locale *string, fields []entity.FieldDefinition) (map[string]interface{}, error) {
	localeStr := derefString(locale)
	if err := resolver.docUC.Publish(ctx, slug, docID, localeStr, fields, middleware.UserID(ctx)); err != nil {
		return nil, err
	}
	doc, err := resolver.docUC.GetPublished(ctx, slug, docID, localeStr, fields)
	if err != nil {
		return nil, err
	}
	return resolver.docToMap(ctx, doc, fields), nil
}

func (resolver *Resolver) unpublishDocument(ctx context.Context, slug, docID string, locale *string, fields []entity.FieldDefinition) (map[string]interface{}, error) {
	localeStr := derefString(locale)
	if err := resolver.docUC.Unpublish(ctx, slug, docID, localeStr, fields); err != nil {
		return nil, err
	}
	doc, _, err := resolver.docUC.GetForEdit(ctx, slug, docID, localeStr, fields)
	if err != nil {
		return nil, err
	}
	return resolver.docToMap(ctx, doc, fields), nil
}

// ── Single-type helpers ──

func (resolver *Resolver) getSingleType(ctx context.Context, slug string, locale, status *string, fields []entity.FieldDefinition) (map[string]interface{}, error) {
	localeStr := derefString(locale)
	if derefString(status) == "draft" && middleware.UserID(ctx) != "" {
		doc, _, err := resolver.docUC.GetSingleType(ctx, slug, localeStr, fields)
		if err != nil {
			return nil, nil
		}
		return resolver.docToMap(ctx, doc, fields), nil
	}
	doc, err := resolver.docUC.GetPublishedSingleType(ctx, slug, localeStr, fields)
	if err != nil {
		return nil, nil
	}
	return resolver.docToMap(ctx, doc, fields), nil
}

func (resolver *Resolver) saveSingleType(ctx context.Context, slug string, data interface{}, locale *string, fields []entity.FieldDefinition) (map[string]interface{}, error) {
	dataMap := structToMap(data)
	localeStr := derefString(locale)
	saved, err := resolver.docUC.SaveSingleType(ctx, slug, dataMap, localeStr, fields, middleware.UserID(ctx))
	if err != nil {
		return nil, err
	}
	doc, _, err := resolver.docUC.GetSingleType(ctx, slug, saved.Locale, fields)
	if err != nil {
		return nil, err
	}
	return resolver.docToMap(ctx, doc, fields), nil
}

func (resolver *Resolver) publishSingleType(ctx context.Context, slug string, locale *string, fields []entity.FieldDefinition) (map[string]interface{}, error) {
	localeStr := derefString(locale)
	if err := resolver.docUC.PublishSingleType(ctx, slug, localeStr, fields, middleware.UserID(ctx)); err != nil {
		return nil, err
	}
	doc, _, err := resolver.docUC.GetSingleType(ctx, slug, localeStr, fields)
	if err != nil {
		return nil, err
	}
	return resolver.docToMap(ctx, doc, fields), nil
}

func (resolver *Resolver) unpublishSingleType(ctx context.Context, slug string, locale *string, fields []entity.FieldDefinition) (map[string]interface{}, error) {
	localeStr := derefString(locale)
	if err := resolver.docUC.UnpublishSingleType(ctx, slug, localeStr, fields); err != nil {
		return nil, err
	}
	doc, _, err := resolver.docUC.GetSingleType(ctx, slug, localeStr, fields)
	if err != nil {
		return nil, err
	}
	return resolver.docToMap(ctx, doc, fields), nil
}

// ── Shared helpers ──

func (resolver *Resolver) buildListResponse(ctx context.Context, docs []*entity.Document, fields []entity.FieldDefinition, total int64, start, limit int) map[string]interface{} {
	page := 1
	pageSize := limit
	if limit == -1 {
		pageSize = int(total)
	} else if limit > 0 {
		page = start/limit + 1
	}

	return map[string]interface{}{
		"items": resolver.docsToMaps(ctx, docs, fields),
		"meta": map[string]interface{}{
			"pagination": map[string]interface{}{
				"page":     page,
				"pageSize": pageSize,
				"total":    total,
			},
		},
	}
}

func parsePagination(input *model.PaginationInput) (start, limit int, err error) {
	if input == nil {
		return 0, 10, nil
	}

	hasOffset := input.Start != nil || input.Limit != nil
	hasPage := input.Page != nil || input.PageSize != nil

	if hasOffset && hasPage {
		return 0, 0, fmt.Errorf("cannot mix offset (start/limit) and page (page/pageSize) modes")
	}

	if hasPage {
		if input.Page == nil || input.PageSize == nil {
			return 0, 0, fmt.Errorf("page and pageSize must both be provided")
		}
		if *input.Page < 1 {
			return 0, 0, fmt.Errorf("page must be >= 1")
		}
		if *input.PageSize == 0 {
			return 0, 0, fmt.Errorf("pageSize must not be 0")
		}
		pageSize := *input.PageSize
		if pageSize > 100 {
			pageSize = 100
		}
		return (*input.Page - 1) * pageSize, pageSize, nil
	}

	start = 0
	if input.Start != nil {
		start = *input.Start
		if start < 0 {
			start = 0
		}
	}
	limit = 10
	if input.Limit != nil {
		if *input.Limit == 0 {
			return 0, 0, fmt.Errorf("limit must not be 0")
		}
		if *input.Limit == -1 {
			limit = -1
		} else {
			limit = *input.Limit
			if limit > 100 {
				limit = 100
			}
		}
	}
	return start, limit, nil
}

func (resolver *Resolver) docsToMaps(ctx context.Context, docs []*entity.Document, fields []entity.FieldDefinition) []map[string]interface{} {
	items := make([]map[string]interface{}, len(docs))
	for idx, doc := range docs {
		items[idx] = resolver.docToMap(ctx, doc, fields)
	}
	return items
}

func structToMap(val interface{}) map[string]interface{} {
	switch typed := val.(type) {
	case map[string]interface{}:
		return typed
	default:
		encoded, err := json.Marshal(val)
		if err != nil {
			return nil
		}
		var result map[string]interface{}
		if err := json.Unmarshal(encoded, &result); err != nil {
			return nil
		}
		return result
	}
}

func derefString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func convertFiltersByReflection(filters interface{}) []entity.FilterNode {
	val := reflect.ValueOf(filters)
	if val.Kind() != reflect.Slice || val.Len() == 0 {
		return nil
	}

	var nodes []entity.FilterNode
	for idx := 0; idx < val.Len(); idx++ {
		elem := val.Index(idx)
		if elem.IsNil() {
			continue
		}
		nodes = append(nodes, convertSingleFilter(elem.Elem())...)
	}
	if len(nodes) == 0 {
		return nil
	}
	return nodes
}
