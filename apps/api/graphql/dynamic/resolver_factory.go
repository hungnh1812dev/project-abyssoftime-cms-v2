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
	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"

	contenttype "project-abyssoftime-cms-v2/api/internal/usecase/content_type"
)

type DocumentUseCase interface {
	Save(ctx context.Context, contentTypeSlug string, doc *entity.Document, userID string) (*entity.Document, error)
	GetForEdit(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, string, error)
	GetPublished(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error)
	Publish(ctx context.Context, contentTypeSlug, documentID, locale, userID string) error
	Unpublish(ctx context.Context, contentTypeSlug, documentID, locale string) error
	Delete(ctx context.Context, contentTypeSlug, documentID string) error
	GetSingleType(ctx context.Context, contentTypeSlug, locale string) (*entity.Document, string, error)
	SaveSingleType(ctx context.Context, contentTypeSlug string, data map[string]any, locale, userID string) (*entity.Document, error)
	PublishSingleType(ctx context.Context, contentTypeSlug, locale, userID string) error
	UnpublishSingleType(ctx context.Context, contentTypeSlug, locale string) error
	GetAllPaginated(ctx context.Context, contentTypeSlug string, start, size int, locale string) ([]*entity.Document, []string, int64, error)
}

type ContentTypeUseCase interface {
	FindAll(ctx context.Context) ([]*entity.ContentType, error)
}

type ResolverFactory struct {
	docUC DocumentUseCase
	ctUC  ContentTypeUseCase
}

func NewResolverFactory(docUC DocumentUseCase, ctUC ContentTypeUseCase) *ResolverFactory {
	return &ResolverFactory{docUC: docUC, ctUC: ctUC}
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
		objType := f.buildObjectType(def, timeScalar, jsonScalar)
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

func (f *ResolverFactory) buildObjectType(def contenttype.ContentTypeDefinition, timeSc, jsonSc *graphql.Scalar) *graphql.Object {
	fields := graphql.Fields{
		"documentId":  &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"locale":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"status":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"createdAt":   &graphql.Field{Type: graphql.NewNonNull(timeSc)},
		"updatedAt":   &graphql.Field{Type: graphql.NewNonNull(timeSc)},
		"publishedAt": &graphql.Field{Type: timeSc},
	}
	for _, fd := range def.Fields {
		if fd.Type == "layout" || fd.Type == "component" {
			continue
		}
		fields[fd.Name] = &graphql.Field{Type: gqlScalarFor(fd.Type, jsonSc)}
	}
	return graphql.NewObject(graphql.ObjectConfig{
		Name:   slugToPascalCase(def.Slug),
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

	connType := graphql.NewObject(graphql.ObjectConfig{
		Name: pascal + "Connection",
		Fields: graphql.Fields{
			"items": &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(objType)))},
			"total": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"start": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"size":  &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		},
	})

	qf[camel] = &graphql.Field{
		Type: objType,
		Args: graphql.FieldConfigArgument{
			camel + "Id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			"locale":     &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			docID := p.Args[camel+"Id"].(string)
			locale, _ := p.Args["locale"].(string)
			doc, status, err := f.docUC.GetForEdit(p.Context, slug, docID, locale)
			if err != nil {
				return nil, err
			}
			return docToMap(doc, status), nil
		},
	}

	qf[camel+"List"] = &graphql.Field{
		Type: graphql.NewNonNull(connType),
		Args: graphql.FieldConfigArgument{
			"start":  &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 0},
			"size":   &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 20},
			"locale": &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			start, _ := p.Args["start"].(int)
			size, _ := p.Args["size"].(int)
			locale, _ := p.Args["locale"].(string)
			docs, statuses, total, err := f.docUC.GetAllPaginated(p.Context, slug, start, size, locale)
			if err != nil {
				return nil, err
			}
			items := make([]map[string]any, len(docs))
			for i, d := range docs {
				items[i] = docToMap(d, statuses[i])
			}
			return map[string]any{"items": items, "total": total, "start": start, "size": size}, nil
		},
	}

	mf["create"+pascal] = &graphql.Field{
		Type: graphql.NewNonNull(objType),
		Args: graphql.FieldConfigArgument{
			"data": &graphql.ArgumentConfig{Type: graphql.NewNonNull(inputType)},
		},
		Resolve: f.authRequired(func(p graphql.ResolveParams) (any, error) {
			data := inputToMap(p.Args["data"])
			doc := &entity.Document{Data: data}
			saved, err := f.docUC.Save(p.Context, slug, doc, middleware.UserID(p.Context))
			if err != nil {
				return nil, err
			}
			return docToMap(saved, "draft"), nil
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
			doc := &entity.Document{DocumentID: docID, Data: data}
			saved, err := f.docUC.Save(p.Context, slug, doc, middleware.UserID(p.Context))
			if err != nil {
				return nil, err
			}
			_, status, err := f.docUC.GetForEdit(p.Context, slug, saved.DocumentID, saved.Locale)
			if err != nil {
				return nil, err
			}
			return docToMap(saved, status), nil
		}),
	}

	mf["delete"+pascal] = &graphql.Field{
		Type: graphql.NewNonNull(graphql.Boolean),
		Args: graphql.FieldConfigArgument{
			camel + "Id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
		},
		Resolve: f.authRequired(func(p graphql.ResolveParams) (any, error) {
			docID := p.Args[camel+"Id"].(string)
			return true, f.docUC.Delete(p.Context, slug, docID)
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
			if err := f.docUC.Publish(p.Context, slug, docID, locale, middleware.UserID(p.Context)); err != nil {
				return nil, err
			}
			doc, err := f.docUC.GetPublished(p.Context, slug, docID, locale)
			if err != nil {
				return nil, err
			}
			return docToMap(doc, "published"), nil
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
			doc, status, err := f.docUC.GetForEdit(p.Context, slug, docID, locale)
			if err != nil {
				return nil, err
			}
			return docToMap(doc, status), nil
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

	qf[camel] = &graphql.Field{
		Type: objType,
		Args: graphql.FieldConfigArgument{
			"locale": &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			locale, _ := p.Args["locale"].(string)
			doc, status, err := f.docUC.GetSingleType(p.Context, slug, locale)
			if err != nil {
				return nil, nil
			}
			return docToMap(doc, status), nil
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
			saved, err := f.docUC.SaveSingleType(p.Context, slug, data, locale, middleware.UserID(p.Context))
			if err != nil {
				return nil, err
			}
			doc, status, err := f.docUC.GetSingleType(p.Context, slug, saved.Locale)
			if err != nil {
				return nil, err
			}
			return docToMap(doc, status), nil
		}),
	}

	mf["publish"+pascal] = &graphql.Field{
		Type: graphql.NewNonNull(objType),
		Args: graphql.FieldConfigArgument{
			"locale": &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: f.authRequired(func(p graphql.ResolveParams) (any, error) {
			locale, _ := p.Args["locale"].(string)
			if err := f.docUC.PublishSingleType(p.Context, slug, locale, middleware.UserID(p.Context)); err != nil {
				return nil, err
			}
			doc, status, err := f.docUC.GetSingleType(p.Context, slug, locale)
			if err != nil {
				return nil, err
			}
			return docToMap(doc, status), nil
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
			doc, status, err := f.docUC.GetSingleType(p.Context, slug, locale)
			if err != nil {
				return nil, err
			}
			return docToMap(doc, status), nil
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

func docToMap(d *entity.Document, status string) map[string]any {
	m := map[string]any{
		"documentId":  d.DocumentID,
		"locale":      d.Locale,
		"status":      status,
		"createdAt":   d.CreatedAt,
		"updatedAt":   d.UpdatedAt,
		"publishedAt": nil,
	}
	if !d.PublishedAt.IsZero() {
		m["publishedAt"] = d.PublishedAt
	}
	for k, v := range d.Data {
		m[k] = v
	}
	return m
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
	case "text", "richtext", "media":
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
