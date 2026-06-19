package dynamic

import (
	"fmt"
	"strings"

	contenttype "project-abyssoftime-cms-v2/api/internal/usecase/content_type"
)

type SchemaBuilder struct {
	defs []contenttype.ContentTypeDefinition
}

func NewSchemaBuilder(defs []contenttype.ContentTypeDefinition) *SchemaBuilder {
	return &SchemaBuilder{defs: defs}
}

func (b *SchemaBuilder) BuildBaseSchema() string {
	return `scalar JSON
scalar Time

type ContentType {
  id: ID!
  name: String!
  slug: String!
  kind: String!
  createdAt: Time!
  updatedAt: Time!
}

type Query {
  contentTypes: [ContentType!]!
}
`
}

func (b *SchemaBuilder) BuildContentTypeSDL(def contenttype.ContentTypeDefinition) string {
	typeName := slugToPascalCase(def.Slug)
	camel := slugToCamelCase(def.Slug)

	var sb strings.Builder

	// Object type
	fmt.Fprintf(&sb, "type %s {\n", typeName)
	sb.WriteString("  documentId: ID!\n")
	for _, f := range def.Fields {
		if f.Type == "layout" || f.Type == "component" {
			continue
		}
		fmt.Fprintf(&sb, "  %s: %s\n", f.Name, fieldTypeToGraphQL(f.Type))
	}
	sb.WriteString("  locale: String!\n")
	sb.WriteString("  status: String!\n")
	sb.WriteString("  createdAt: Time!\n")
	sb.WriteString("  updatedAt: Time!\n")
	sb.WriteString("  publishedAt: Time\n")
	sb.WriteString("}\n\n")

	// Input type
	fmt.Fprintf(&sb, "input %sInput {\n", typeName)
	for _, f := range def.Fields {
		if f.Type == "layout" || f.Type == "component" {
			continue
		}
		fmt.Fprintf(&sb, "  %s: %s\n", f.Name, fieldTypeToGraphQL(f.Type))
	}
	sb.WriteString("}\n\n")

	if def.Kind == "collection" {
		// Connection type
		fmt.Fprintf(&sb, "type %sConnection {\n", typeName)
		fmt.Fprintf(&sb, "  items: [%s!]!\n", typeName)
		sb.WriteString("  total: Int!\n")
		sb.WriteString("  start: Int!\n")
		sb.WriteString("  size: Int!\n")
		sb.WriteString("}\n\n")

		// Queries
		sb.WriteString("extend type Query {\n")
		fmt.Fprintf(&sb, "  %s(%sId: ID!, locale: String): %s\n", camel, camel, typeName)
		fmt.Fprintf(&sb, "  %sList(start: Int, size: Int, locale: String): %sConnection!\n", camel, typeName)
		sb.WriteString("}\n\n")

		// Mutations
		sb.WriteString("extend type Mutation {\n")
		fmt.Fprintf(&sb, "  create%s(data: %sInput!): %s!\n", typeName, typeName, typeName)
		fmt.Fprintf(&sb, "  update%s(%sId: ID!, data: %sInput!): %s!\n", typeName, camel, typeName, typeName)
		fmt.Fprintf(&sb, "  delete%s(%sId: ID!): Boolean!\n", typeName, camel)
		fmt.Fprintf(&sb, "  publish%s(%sId: ID!, locale: String): %s!\n", typeName, camel, typeName)
		fmt.Fprintf(&sb, "  unpublish%s(%sId: ID!, locale: String): %s!\n", typeName, camel, typeName)
		sb.WriteString("}\n")
	} else {
		// Single-type queries
		sb.WriteString("extend type Query {\n")
		fmt.Fprintf(&sb, "  %s(locale: String): %s\n", camel, typeName)
		sb.WriteString("}\n\n")

		// Single-type mutations (no delete)
		sb.WriteString("extend type Mutation {\n")
		fmt.Fprintf(&sb, "  save%s(data: %sInput!, locale: String): %s!\n", typeName, typeName, typeName)
		fmt.Fprintf(&sb, "  publish%s(locale: String): %s!\n", typeName, typeName)
		fmt.Fprintf(&sb, "  unpublish%s(locale: String): %s!\n", typeName, typeName)
		sb.WriteString("}\n")
	}

	return sb.String()
}

func (b *SchemaBuilder) BuildSDL() string {
	var sb strings.Builder
	sb.WriteString(b.BuildBaseSchema())
	for _, def := range b.defs {
		sb.WriteString("\n")
		sb.WriteString(b.BuildContentTypeSDL(def))
	}
	return sb.String()
}

func slugToPascalCase(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

func slugToCamelCase(slug string) string {
	pascal := slugToPascalCase(slug)
	if len(pascal) == 0 {
		return ""
	}
	return strings.ToLower(pascal[:1]) + pascal[1:]
}

func fieldTypeToGraphQL(ft string) string {
	switch ft {
	case "text", "richtext":
		return "String"
	case "number":
		return "Float"
	case "boolean":
		return "Boolean"
	case "media":
		return "String"
	case "json":
		return "JSON"
	default:
		return "String"
	}
}
