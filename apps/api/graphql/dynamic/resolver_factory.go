package dynamic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	gqlhandler "github.com/graphql-go/handler"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"

	contenttype "project-abyssoftime-cms-v2/api/internal/usecase/content_type"
)

type DocumentUseCase interface {
	Save(ctx context.Context, contentTypeSlug string, doc *entity.Document, fields []entity.FieldDefinition, userID string) (*entity.Document, error)
	GetForEdit(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition) (*entity.Document, string, error)
	GetPublished(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition) (*entity.Document, error)
	Publish(ctx context.Context, contentTypeSlug, documentID, locale string, fields []entity.FieldDefinition, userID string) error
	Unpublish(ctx context.Context, contentTypeSlug, documentID, locale string) error
	Delete(ctx context.Context, contentTypeSlug, documentID string, fields []entity.FieldDefinition) error
	GetSingleType(ctx context.Context, contentTypeSlug, locale string, fields []entity.FieldDefinition) (*entity.Document, string, error)
	SaveSingleType(ctx context.Context, contentTypeSlug string, data map[string]any, locale string, fields []entity.FieldDefinition, userID string) (*entity.Document, error)
	PublishSingleType(ctx context.Context, contentTypeSlug, locale string, fields []entity.FieldDefinition, userID string) error
	UnpublishSingleType(ctx context.Context, contentTypeSlug, locale string) error
	GetAllPaginated(ctx context.Context, contentTypeSlug string, start, size int, locale string, fields []entity.FieldDefinition, orderBy string, sortDir int) ([]*entity.Document, []string, int64, error)
	GetPublishedPaginated(ctx context.Context, contentTypeSlug string, start, size int, locale string, fields []entity.FieldDefinition) ([]*entity.Document, int64, error)
	GetPublishedSingleType(ctx context.Context, contentTypeSlug, locale string, fields []entity.FieldDefinition) (*entity.Document, error)
}

type ContentTypeUseCase interface {
	FindAll(ctx context.Context) ([]*entity.ContentType, error)
}

type ResolverFactory struct {
	docUC     DocumentUseCase
	ctUC      ContentTypeUseCase
	mediaRepo repository.MediaAssetRepository
}

func NewResolverFactory(docUC DocumentUseCase, ctUC ContentTypeUseCase, mediaRepo repository.MediaAssetRepository) *ResolverFactory {
	return &ResolverFactory{docUC: docUC, ctUC: ctUC, mediaRepo: mediaRepo}
}

func (f *ResolverFactory) BuildHandler(defs []contenttype.ContentTypeDefinition) (http.Handler, error) {
	jsonScalar := graphql.NewScalar(graphql.ScalarConfig{
		Name: "JSON",
		Serialize: func(value any) any {
			return value
		},
		ParseValue: func(value any) any {
			return value
		},
		ParseLiteral: func(valueAST ast.Value) any {
			return valueAST.GetValue()
		},
	})
	timeScalar := graphql.DateTime

	mediaAssetType := graphql.NewObject(graphql.ObjectConfig{
		Name: "MediaAsset",
		Fields: graphql.Fields{
			"documentId":   &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"url":          &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"thumbnailUrl": &graphql.Field{Type: graphql.String},
			"fileName":     &graphql.Field{Type: graphql.String},
			"width":        &graphql.Field{Type: graphql.Int},
			"height":       &graphql.Field{Type: graphql.Int},
		},
	})

	contentTypeObj := graphql.NewObject(graphql.ObjectConfig{
		Name: "ContentType",
		Fields: graphql.Fields{
			"id":        &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"name":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"slug":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"kind":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"createdAt": &graphql.Field{Type: graphql.NewNonNull(timeScalar)},
			"updatedAt": &graphql.Field{Type: graphql.NewNonNull(timeScalar)},
		},
	})

	queryFields := graphql.Fields{
		"contentTypes": &graphql.Field{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(contentTypeObj))),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				cts, err := f.ctUC.FindAll(p.Context)
				if err != nil {
					return nil, err
				}
				result := make([]map[string]any, len(cts))
				for i, ct := range cts {
					result[i] = map[string]any{
						"id": ct.ID, "name": ct.Name, "slug": ct.Slug,
						"kind": string(ct.Kind), "createdAt": ct.CreatedAt, "updatedAt": ct.UpdatedAt,
					}
				}
				return result, nil
			},
		},
	}

	mutationFields := graphql.Fields{}

	for _, def := range defs {
		objType := f.buildObjectType(def, timeScalar, jsonScalar, mediaAssetType)
		inputType := f.buildInputType(def, jsonScalar)

		if def.Kind == "collection" {
			f.addCollectionFields(def, objType, inputType, timeScalar, queryFields, mutationFields)
		} else {
			f.addSingleFields(def, objType, inputType, queryFields, mutationFields)
		}
	}

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Query",
		Fields: queryFields,
	})

	var mutationType *graphql.Object
	if len(mutationFields) > 0 {
		mutationType = graphql.NewObject(graphql.ObjectConfig{
			Name:   "Mutation",
			Fields: mutationFields,
		})
	}

	schemaCfg := graphql.SchemaConfig{Query: queryType}
	if mutationType != nil {
		schemaCfg.Mutation = mutationType
	}

	schema, err := graphql.NewSchema(schemaCfg)
	if err != nil {
		return nil, fmt.Errorf("graphql schema: %w", err)
	}

	h := gqlhandler.New(&gqlhandler.Config{
		Schema:     &schema,
		Pretty:     false,
		Playground: false,
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = injectAuthFromRequest(ctx, r)
		h.ContextHandler(ctx, w, r)
	}), nil
}

func (f *ResolverFactory) buildComponentType(parentType string, fd entity.FieldDefinition, jsonSc *graphql.Scalar, mediaType *graphql.Object) *graphql.Object {
	compFields := graphql.Fields{}
	for _, sub := range fd.Fields {
		if sub.Type == "media" {
			compFields[sub.Name] = &graphql.Field{Type: mediaType}
		} else {
			compFields[sub.Name] = &graphql.Field{Type: gqlScalarFor(sub.Type, jsonSc)}
		}
	}
	return graphql.NewObject(graphql.ObjectConfig{
		Name:   parentType + slugToPascalCase(fd.Name),
		Fields: compFields,
	})
}

func (f *ResolverFactory) buildObjectType(def contenttype.ContentTypeDefinition, timeSc, jsonSc *graphql.Scalar, mediaType *graphql.Object) *graphql.Object {
	typeName := slugToPascalCase(def.Slug)
	fields := graphql.Fields{
		"documentId":  &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"locale":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"createdAt":   &graphql.Field{Type: graphql.NewNonNull(timeSc)},
		"updatedAt":   &graphql.Field{Type: graphql.NewNonNull(timeSc)},
		"publishedAt": &graphql.Field{Type: timeSc},
	}
	for _, fd := range def.Fields {
		if fd.Type == "layout" {
			continue
		}
		if fd.Type == "component" {
			compType := f.buildComponentType(typeName, fd, jsonSc, mediaType)
			fields[fd.Name] = &graphql.Field{Type: compType}
			continue
		}
		if fd.Type == "media" {
			fields[fd.Name] = &graphql.Field{Type: mediaType}
			continue
		}
		fields[fd.Name] = &graphql.Field{Type: gqlScalarFor(fd.Type, jsonSc)}
	}
	return graphql.NewObject(graphql.ObjectConfig{
		Name:   typeName,
		Fields: fields,
	})
}

func (f *ResolverFactory) buildInputType(def contenttype.ContentTypeDefinition, jsonSc *graphql.Scalar) *graphql.InputObject {
	fields := graphql.InputObjectConfigFieldMap{}
	for _, fd := range def.Fields {
		if fd.Type == "layout" || fd.Type == "component" {
			continue
		}
		fields[fd.Name] = &graphql.InputObjectFieldConfig{Type: gqlScalarFor(fd.Type, jsonSc)}
	}
	return graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   slugToPascalCase(def.Slug) + "Input",
		Fields: fields,
	})
}

func (f *ResolverFactory) addCollectionFields(
	def contenttype.ContentTypeDefinition,
	objType *graphql.Object,
	inputType *graphql.InputObject,
	timeSc *graphql.Scalar,
	qf, mf graphql.Fields,
) {
	camel := slugToCamelCase(def.Slug)
	pascal := slugToPascalCase(def.Slug)
	slug := def.Slug
	fields := def.Fields

	qf[camel] = &graphql.Field{
		Type: objType,
		Args: graphql.FieldConfigArgument{
			camel + "Id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			"locale":     &graphql.ArgumentConfig{Type: graphql.String},
			"status":     &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			docID := p.Args[camel+"Id"].(string)
			locale, _ := p.Args["locale"].(string)
			statusFilter, _ := p.Args["status"].(string)
			if statusFilter == "draft" && middleware.UserID(p.Context) != "" {
				doc, _, err := f.docUC.GetForEdit(p.Context, slug, docID, locale, fields)
				if err != nil {
					return nil, err
				}
				return f.docToMap(p.Context, doc, fields), nil
			}
			doc, err := f.docUC.GetPublished(p.Context, slug, docID, locale, fields)
			if err != nil {
				return nil, err
			}
			return f.docToMap(p.Context, doc, fields), nil
		},
	}

	qf[camel+"List"] = &graphql.Field{
		Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(objType))),
		Args: graphql.FieldConfigArgument{
			"start":  &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 0},
			"size":   &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 20},
			"locale": &graphql.ArgumentConfig{Type: graphql.String},
			"status": &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			start, _ := p.Args["start"].(int)
			size, _ := p.Args["size"].(int)
			locale, _ := p.Args["locale"].(string)
			statusFilter, _ := p.Args["status"].(string)
			if statusFilter == "draft" && middleware.UserID(p.Context) != "" {
				docs, _, _, err := f.docUC.GetAllPaginated(p.Context, slug, start, size, locale, fields, "createdAt", -1)
				if err != nil {
					return nil, err
				}
				items := make([]map[string]any, len(docs))
				for i, d := range docs {
					items[i] = f.docToMap(p.Context, d, fields)
				}
				return items, nil
			}
			docs, _, err := f.docUC.GetPublishedPaginated(p.Context, slug, start, size, locale, fields)
			if err != nil {
				return nil, err
			}
			items := make([]map[string]any, len(docs))
			for i, d := range docs {
				items[i] = f.docToMap(p.Context, d, fields)
			}
			return items, nil
		},
	}

	mf["create"+pascal] = &graphql.Field{
		Type: graphql.NewNonNull(objType),
		Args: graphql.FieldConfigArgument{
			"data": &graphql.ArgumentConfig{Type: graphql.NewNonNull(inputType)},
		},
		Resolve: f.authRequired(func(p graphql.ResolveParams) (any, error) {
			data := inputToMap(p.Args["data"])
			doc := &entity.Document{Fields: data}
			saved, err := f.docUC.Save(p.Context, slug, doc, fields, middleware.UserID(p.Context))
			if err != nil {
				return nil, err
			}
			return f.docToMap(p.Context, saved, fields), nil
		}),
	}

	mf["update"+pascal] = &graphql.Field{
		Type: graphql.NewNonNull(objType),
		Args: graphql.FieldConfigArgument{
			camel + "Id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			"data":       &graphql.ArgumentConfig{Type: graphql.NewNonNull(inputType)},
		},
		Resolve: f.authRequired(func(p graphql.ResolveParams) (any, error) {
			docID := p.Args[camel+"Id"].(string)
			data := inputToMap(p.Args["data"])
			doc := &entity.Document{DocumentID: docID, Fields: data}
			saved, err := f.docUC.Save(p.Context, slug, doc, fields, middleware.UserID(p.Context))
			if err != nil {
				return nil, err
			}
			return f.docToMap(p.Context, saved, fields), nil
		}),
	}

	mf["delete"+pascal] = &graphql.Field{
		Type: graphql.NewNonNull(graphql.Boolean),
		Args: graphql.FieldConfigArgument{
			camel + "Id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
		},
		Resolve: f.authRequired(func(p graphql.ResolveParams) (any, error) {
			docID := p.Args[camel+"Id"].(string)
			return true, f.docUC.Delete(p.Context, slug, docID, fields)
		}),
	}

	mf["publish"+pascal] = &graphql.Field{
		Type: graphql.NewNonNull(objType),
		Args: graphql.FieldConfigArgument{
			camel + "Id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			"locale":     &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: f.authRequired(func(p graphql.ResolveParams) (any, error) {
			docID := p.Args[camel+"Id"].(string)
			locale, _ := p.Args["locale"].(string)
			if err := f.docUC.Publish(p.Context, slug, docID, locale, fields, middleware.UserID(p.Context)); err != nil {
				return nil, err
			}
			doc, err := f.docUC.GetPublished(p.Context, slug, docID, locale, fields)
			if err != nil {
				return nil, err
			}
			return f.docToMap(p.Context, doc, fields), nil
		}),
	}

	mf["unpublish"+pascal] = &graphql.Field{
		Type: graphql.NewNonNull(objType),
		Args: graphql.FieldConfigArgument{
			camel + "Id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			"locale":     &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: f.authRequired(func(p graphql.ResolveParams) (any, error) {
			docID := p.Args[camel+"Id"].(string)
			locale, _ := p.Args["locale"].(string)
			if err := f.docUC.Unpublish(p.Context, slug, docID, locale); err != nil {
				return nil, err
			}
			doc, _, err := f.docUC.GetForEdit(p.Context, slug, docID, locale, fields)
			if err != nil {
				return nil, err
			}
			return f.docToMap(p.Context, doc, fields), nil
		}),
	}
}

func (f *ResolverFactory) addSingleFields(
	def contenttype.ContentTypeDefinition,
	objType *graphql.Object,
	inputType *graphql.InputObject,
	qf, mf graphql.Fields,
) {
	camel := slugToCamelCase(def.Slug)
	pascal := slugToPascalCase(def.Slug)
	slug := def.Slug
	fields := def.Fields

	qf[camel] = &graphql.Field{
		Type: objType,
		Args: graphql.FieldConfigArgument{
			"locale": &graphql.ArgumentConfig{Type: graphql.String},
			"status": &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			locale, _ := p.Args["locale"].(string)
			statusFilter, _ := p.Args["status"].(string)
			if statusFilter == "draft" && middleware.UserID(p.Context) != "" {
				doc, _, err := f.docUC.GetSingleType(p.Context, slug, locale, fields)
				if err != nil {
					return nil, nil
				}
				return f.docToMap(p.Context, doc, fields), nil
			}
			doc, err := f.docUC.GetPublishedSingleType(p.Context, slug, locale, fields)
			if err != nil {
				return nil, nil
			}
			return f.docToMap(p.Context, doc, fields), nil
		},
	}

	mf["save"+pascal] = &graphql.Field{
		Type: graphql.NewNonNull(objType),
		Args: graphql.FieldConfigArgument{
			"data":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(inputType)},
			"locale": &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: f.authRequired(func(p graphql.ResolveParams) (any, error) {
			data := inputToMap(p.Args["data"])
			locale, _ := p.Args["locale"].(string)
			saved, err := f.docUC.SaveSingleType(p.Context, slug, data, locale, fields, middleware.UserID(p.Context))
			if err != nil {
				return nil, err
			}
			doc, _, err := f.docUC.GetSingleType(p.Context, slug, saved.Locale, fields)
			if err != nil {
				return nil, err
			}
			return f.docToMap(p.Context, doc, fields), nil
		}),
	}

	mf["publish"+pascal] = &graphql.Field{
		Type: graphql.NewNonNull(objType),
		Args: graphql.FieldConfigArgument{
			"locale": &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: f.authRequired(func(p graphql.ResolveParams) (any, error) {
			locale, _ := p.Args["locale"].(string)
			if err := f.docUC.PublishSingleType(p.Context, slug, locale, fields, middleware.UserID(p.Context)); err != nil {
				return nil, err
			}
			doc, _, err := f.docUC.GetSingleType(p.Context, slug, locale, fields)
			if err != nil {
				return nil, err
			}
			return f.docToMap(p.Context, doc, fields), nil
		}),
	}

	mf["unpublish"+pascal] = &graphql.Field{
		Type: graphql.NewNonNull(objType),
		Args: graphql.FieldConfigArgument{
			"locale": &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: f.authRequired(func(p graphql.ResolveParams) (any, error) {
			locale, _ := p.Args["locale"].(string)
			if err := f.docUC.UnpublishSingleType(p.Context, slug, locale); err != nil {
				return nil, err
			}
			doc, _, err := f.docUC.GetSingleType(p.Context, slug, locale, fields)
			if err != nil {
				return nil, err
			}
			return f.docToMap(p.Context, doc, fields), nil
		}),
	}
}

func (f *ResolverFactory) authRequired(resolve graphql.FieldResolveFn) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (any, error) {
		if middleware.UserID(p.Context) == "" {
			return nil, fmt.Errorf("unauthorized")
		}
		return resolve(p)
	}
}

func (f *ResolverFactory) docToMap(ctx context.Context, d *entity.Document, fields []entity.FieldDefinition) map[string]any {
	m := map[string]any{
		"documentId":  d.DocumentID,
		"locale":      d.Locale,
		"createdAt":   d.CreatedAt,
		"updatedAt":   d.UpdatedAt,
		"publishedAt": nil,
	}
	if d.PublishedAt != nil {
		m["publishedAt"] = *d.PublishedAt
	}
	for _, fd := range fields {
		if fd.Type == "media" {
			m[fd.Name] = f.resolveMediaField(ctx, d.Fields[fd.Name])
		} else if fd.Type == "component" {
			raw := d.Fields[fd.Name]
			m[fd.Name] = f.resolveComponentMedia(ctx, raw, fd.Fields)
		} else if fd.Type != "layout" {
			m[fd.Name] = d.Fields[fd.Name]
		}
	}
	return m
}

func (f *ResolverFactory) resolveMediaField(ctx context.Context, value any) any {
	if value == nil {
		return nil
	}
	docID, ok := value.(string)
	if !ok || docID == "" {
		return nil
	}
	if f.mediaRepo == nil {
		return nil
	}
	asset, err := f.mediaRepo.FindByDocumentID(ctx, docID)
	if err != nil || asset == nil {
		return nil
	}
	return map[string]any{
		"documentId":   asset.DocumentID,
		"url":          asset.URL,
		"thumbnailUrl": asset.ThumbnailURL,
		"fileName":     asset.FileName,
		"width":        asset.Width,
		"height":       asset.Height,
	}
}

func (f *ResolverFactory) resolveComponentMedia(ctx context.Context, raw any, subFields []entity.FieldDefinition) any {
	if raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case map[string]any:
		return f.resolveComponentMap(ctx, v, subFields)
	case []any:
		arr := make([]map[string]any, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				arr = append(arr, f.resolveComponentMap(ctx, m, subFields))
			}
		}
		return arr
	default:
		return raw
	}
}

func (f *ResolverFactory) resolveComponentMap(ctx context.Context, m map[string]any, subFields []entity.FieldDefinition) map[string]any {
	result := make(map[string]any, len(m))
	for _, sf := range subFields {
		if sf.Type == "media" {
			result[sf.Name] = f.resolveMediaField(ctx, m[sf.Name])
		} else {
			result[sf.Name] = m[sf.Name]
		}
	}
	return result
}

func inputToMap(v any) map[string]any {
	switch val := v.(type) {
	case map[string]any:
		return val
	default:
		b, _ := json.Marshal(v)
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		return m
	}
}

func gqlScalarFor(fieldType string, jsonSc *graphql.Scalar) graphql.Output {
	switch fieldType {
	case "text", "richtext":
		return graphql.String
	case "media":
		return graphql.String
	case "number":
		return graphql.Float
	case "boolean":
		return graphql.Boolean
	case "json":
		return jsonSc
	default:
		return graphql.String
	}
}

func injectAuthFromRequest(ctx context.Context, r *http.Request) context.Context {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return ctx
	}
	claims, err := pkgjwt.ValidateToken(strings.TrimPrefix(header, "Bearer "))
	if err != nil {
		return ctx
	}
	ctx = context.WithValue(ctx, middleware.ContextKeyUserID, claims.UserID)
	ctx = context.WithValue(ctx, middleware.ContextKeyRole, claims.Role)
	return ctx
}
