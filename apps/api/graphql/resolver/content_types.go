package resolver

import (
	"context"
)

func (resolver *queryResolver) ContentTypes(ctx context.Context) ([]map[string]interface{}, error) {
	contentTypes, err := resolver.ctUC.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, len(contentTypes))
	for idx, contentType := range contentTypes {
		result[idx] = map[string]interface{}{
			"id":        contentType.ID,
			"name":      contentType.Name,
			"slug":      contentType.Slug,
			"kind":      string(contentType.Kind),
			"createdAt": contentType.CreatedAt,
			"updatedAt": contentType.UpdatedAt,
		}
	}
	return result, nil
}

func (resolver *mutationResolver) Empty(ctx context.Context) (*bool, error) {
	return nil, nil
}
