package dynamic

import (
	"fmt"
	"strings"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
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

enum SortOrder {
  ASC
  DESC
}

type MediaAsset {
  documentId: ID!
  url: String!
  thumbnailUrl: String
  fileName: String
  width: Int
  height: Int
}

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

	// Component types
	for _, f := range def.Fields {
		if f.Type == "component" {
			writeComponentType(&sb, typeName, f)
		}
	}

	// Object type
	fmt.Fprintf(&sb, "type %s {\n", typeName)
	sb.WriteString("  documentId: ID!\n")
	for _, f := range def.Fields {
		if f.Type == "layout" {
			continue
		}
		if f.Type == "component" {
			compType := typeName + slugToPascalCase(f.Name)
			if f.Repeatable {
				fmt.Fprintf(&sb, "  %s: [%s!]\n", f.Name, compType)
			} else {
				fmt.Fprintf(&sb, "  %s: %s\n", f.Name, compType)
			}
			continue
		}
		fmt.Fprintf(&sb, "  %s: %s\n", f.Name, fieldTypeToGraphQL(f.Type))
	}
	sb.WriteString("  locale: String!\n")
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

	// Filter type
	writeFilterType(&sb, typeName, def.Fields)

	// OrderBy type
	writeOrderByType(&sb, typeName, def.Fields)

	if def.Kind == "collection" {
		// Queries
		sb.WriteString("extend type Query {\n")
		fmt.Fprintf(&sb, "  %s(%sId: ID!, locale: String): %s\n", camel, camel, typeName)
		fmt.Fprintf(&sb, "  %sList(where: %sFilter, orderBy: %sOrderBy, start: Int, size: Int, locale: String): [%s!]!\n", camel, typeName, typeName, typeName)
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

func writeComponentType(sb *strings.Builder, parentType string, f entity.FieldDefinition) {
	compType := parentType + slugToPascalCase(f.Name)

	for _, sub := range f.Fields {
		if sub.Type == "component" {
			writeComponentType(sb, compType, sub)
		}
	}

	fmt.Fprintf(sb, "type %s {\n", compType)
	for _, sub := range flattenLayoutFieldsDef(f.Fields) {
		if sub.Type == "component" {
			nestedType := compType + slugToPascalCase(sub.Name)
			if sub.Repeatable {
				fmt.Fprintf(sb, "  %s: [%s!]\n", sub.Name, nestedType)
			} else {
				fmt.Fprintf(sb, "  %s: %s\n", sub.Name, nestedType)
			}
			continue
		}
		fmt.Fprintf(sb, "  %s: %s\n", sub.Name, fieldTypeToGraphQL(sub.Type))
	}
	sb.WriteString("}\n\n")
}

func flattenLayoutFieldsDef(fields []entity.FieldDefinition) []entity.FieldDefinition {
	var result []entity.FieldDefinition
	for _, f := range fields {
		if f.Type == "layout" {
			result = append(result, f.Fields...)
		} else {
			result = append(result, f)
		}
	}
	return result
}

func writeFilterType(sb *strings.Builder, typeName string, fields []entity.FieldDefinition) {
	fmt.Fprintf(sb, "input %sFilter {\n", typeName)
	for _, f := range fields {
		if f.Type == "layout" {
			continue
		}
		switch f.Type {
		case "text", "richtext":
			fmt.Fprintf(sb, "  %s: StringFilter\n", f.Name)
		case "number":
			fmt.Fprintf(sb, "  %s: NumberFilter\n", f.Name)
		case "boolean":
			fmt.Fprintf(sb, "  %s: BooleanFilter\n", f.Name)
		}
	}
	fmt.Fprintf(sb, "  AND: [%sFilter!]\n", typeName)
	fmt.Fprintf(sb, "  OR: [%sFilter!]\n", typeName)
	fmt.Fprintf(sb, "  NOT: %sFilter\n", typeName)
	sb.WriteString("}\n\n")
}

func writeOrderByType(sb *strings.Builder, typeName string, fields []entity.FieldDefinition) {
	fmt.Fprintf(sb, "input %sOrderBy {\n", typeName)
	for _, f := range fields {
		switch f.Type {
		case "text", "richtext", "number", "boolean":
			fmt.Fprintf(sb, "  %s: SortOrder\n", f.Name)
		}
	}
	sb.WriteString("  createdAt: SortOrder\n")
	sb.WriteString("  updatedAt: SortOrder\n")
	sb.WriteString("  publishedAt: SortOrder\n")
	sb.WriteString("}\n\n")
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
		return "MediaAsset"
	case "json":
		return "JSON"
	default:
		return "String"
	}
}
