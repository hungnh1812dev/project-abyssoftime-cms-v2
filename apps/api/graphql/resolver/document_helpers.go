package resolver

import (
	"context"
	"encoding/json"
	"reflect"

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

func (resolver *Resolver) getDocumentList(ctx context.Context, slug string, filters interface{}, orderBy interface{}, start, size *int, locale, status *string, fields []entity.FieldDefinition) ([]map[string]interface{}, error) {
	startVal := 0
	if start != nil {
		startVal = *start
	}
	sizeVal := 20
	if size != nil {
		sizeVal = *size
	}
	localeStr := derefString(locale)

	orderField, sortDir := extractOrderBy(orderBy)

	var parsedFilters []entity.FilterNode
	if filters != nil {
		parsedFilters = convertFiltersByReflection(filters)
	}

	if derefString(status) == "draft" && middleware.UserID(ctx) != "" {
		docs, _, _, err := resolver.docUC.GetAllPaginated(ctx, slug, startVal, sizeVal, localeStr, fields, orderField, sortDir, parsedFilters)
		if err != nil {
			return nil, err
		}
		return resolver.docsToMaps(ctx, docs, fields), nil
	}

	docs, _, err := resolver.docUC.GetPublishedPaginated(ctx, slug, startVal, sizeVal, localeStr, fields, parsedFilters)
	if err != nil {
		return nil, err
	}
	return resolver.docsToMaps(ctx, docs, fields), nil
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
