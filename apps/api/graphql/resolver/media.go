package resolver

import (
	"context"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

func (resolver *Resolver) docToMap(ctx context.Context, doc *entity.Document, fields []entity.FieldDefinition) map[string]interface{} {
	result := map[string]interface{}{
		"documentId":  doc.DocumentID,
		"locale":      doc.Locale,
		"createdAt":   doc.CreatedAt,
		"updatedAt":   doc.UpdatedAt,
		"publishedAt": nil,
	}
	if doc.PublishedAt != nil {
		result["publishedAt"] = *doc.PublishedAt
	}
	for _, field := range fields {
		if field.Type == "media" {
			result[field.Name] = resolver.resolveMediaField(ctx, doc.Fields[field.Name])
		} else if field.Type == "component" {
			result[field.Name] = resolver.resolveComponentMedia(ctx, doc.Fields[field.Name], field.Fields)
		} else {
			result[field.Name] = doc.Fields[field.Name]
		}
	}
	return result
}

func (resolver *Resolver) resolveMediaField(ctx context.Context, value interface{}) interface{} {
	if value == nil {
		return nil
	}
	if asset, isAsset := value.(*entity.MediaAsset); isAsset {
		return map[string]interface{}{
			"documentId":   asset.DocumentID,
			"url":          asset.URL,
			"thumbnailUrl": asset.ThumbnailURL,
			"fileName":     asset.FileName,
			"width":        asset.Width,
			"height":       asset.Height,
		}
	}
	docID, isString := value.(string)
	if !isString || docID == "" {
		return nil
	}
	if resolver.mediaRepo == nil {
		return nil
	}
	asset, err := resolver.mediaRepo.FindByDocumentID(ctx, docID)
	if err != nil || asset == nil {
		return nil
	}
	return map[string]interface{}{
		"documentId":   asset.DocumentID,
		"url":          asset.URL,
		"thumbnailUrl": asset.ThumbnailURL,
		"fileName":     asset.FileName,
		"width":        asset.Width,
		"height":       asset.Height,
	}
}

func (resolver *Resolver) resolveComponentMedia(ctx context.Context, raw interface{}, subFields []entity.FieldDefinition) interface{} {
	if raw == nil {
		return nil
	}
	switch val := raw.(type) {
	case map[string]interface{}:
		return resolver.resolveComponentMap(ctx, val, subFields)
	case []interface{}:
		arr := make([]map[string]interface{}, 0, len(val))
		for _, item := range val {
			if itemMap, isMap := item.(map[string]interface{}); isMap {
				arr = append(arr, resolver.resolveComponentMap(ctx, itemMap, subFields))
			}
		}
		return arr
	default:
		return raw
	}
}

func (resolver *Resolver) resolveComponentMap(ctx context.Context, compMap map[string]interface{}, subFields []entity.FieldDefinition) map[string]interface{} {
	result := make(map[string]interface{}, len(compMap))
	for _, sub := range subFields {
		if sub.Type == "media" {
			result[sub.Name] = resolver.resolveMediaField(ctx, compMap[sub.Name])
		} else if sub.Type == "component" {
			result[sub.Name] = resolver.resolveComponentMedia(ctx, compMap[sub.Name], sub.Fields)
		} else {
			result[sub.Name] = compMap[sub.Name]
		}
	}
	return result
}
