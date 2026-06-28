package main

import (
	"strings"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	contenttype "project-abyssoftime-cms-v2/api/internal/usecase/content_type"
)

func TestBuildBaseSchema(t *testing.T) {
	sdl := buildBaseSchema()

	for _, want := range []string{
		"directive @auth on FIELD_DEFINITION",
		"scalar JSON",
		"scalar Time",
		"enum SortOrder {",
		"input IDFilter {",
		"input StringFilter {",
		"input NumberFilter {",
		"input BooleanFilter {",
		"input TimeFilter {",
		"input PaginationInput {",
		"start: Int",
		"limit: Int",
		"page: Int",
		"pageSize: Int",
		"type PaginationMeta {",
		"page: Int!",
		"pageSize: Int!",
		"total: Int!",
		"type ListMeta {",
		"pagination: PaginationMeta!",
		"type MediaAsset {",
		"type ContentType {",
		"contentTypes: [ContentType!]!",
		"type Mutation {",
		"_empty: Boolean",
	} {
		if !strings.Contains(sdl, want) {
			t.Errorf("base schema missing %q", want)
		}
	}
}

func TestBuildContentTypeSDL_Collection(t *testing.T) {
	def := contenttype.ContentTypeDefinition{
		Slug: "blog-posts",
		Name: "Blog Posts",
		Kind: "collection",
		Fields: []entity.FieldDefinition{
			{Name: "title", Type: "text"},
			{Name: "featured", Type: "boolean"},
			{Name: "readingTime", Type: "number"},
		},
	}
	sdl := buildContentTypeSDL(def)

	for _, want := range []string{
		"type BlogPosts {",
		"title: String",
		"featured: Boolean",
		"readingTime: Float",
		"documentId: ID!",
		"locale: String!",
		"input BlogPostsInput {",
		"input BlogPostsFilter {",
		"documentId: IDFilter",
		"title: StringFilter",
		"featured: BooleanFilter",
		"readingTime: NumberFilter",
		"and: [BlogPostsFilter!]",
		"or: [BlogPostsFilter!]",
		"not: BlogPostsFilter",
		"input BlogPostsOrderBy {",
		"title: SortOrder",
		"createdAt: SortOrder",
		"blogPosts(documentId: ID!, locale: String, status: String): BlogPosts",
		"type BlogPostsList {",
		"items: [BlogPosts!]!",
		"meta: ListMeta!",
		"blogPostses(pagination: PaginationInput, filters: [BlogPostsFilter!], orderBy: BlogPostsOrderBy, locale: String, status: String): BlogPostsList!",
		"createBlogPosts(data: BlogPostsInput!): BlogPosts! @auth",
		"updateBlogPosts(documentId: ID!, data: BlogPostsInput!): BlogPosts! @auth",
		"deleteBlogPosts(documentId: ID!): Boolean! @auth",
		"publishBlogPosts(documentId: ID!, locale: String): BlogPosts! @auth",
		"unpublishBlogPosts(documentId: ID!, locale: String): BlogPosts! @auth",
	} {
		if !strings.Contains(sdl, want) {
			t.Errorf("collection SDL missing %q\n\nFull SDL:\n%s", want, sdl)
		}
	}
}

func TestBuildContentTypeSDL_Single(t *testing.T) {
	def := contenttype.ContentTypeDefinition{
		Slug: "about-page",
		Name: "About Page",
		Kind: "single",
		Fields: []entity.FieldDefinition{
			{Name: "headline", Type: "text"},
			{Name: "openToWork", Type: "boolean"},
		},
	}
	sdl := buildContentTypeSDL(def)

	for _, want := range []string{
		"type AboutPage {",
		"headline: String",
		"aboutPage(locale: String, status: String): AboutPage",
		"saveAboutPage(data: AboutPageInput!, locale: String): AboutPage! @auth",
		"publishAboutPage(locale: String): AboutPage! @auth",
		"unpublishAboutPage(locale: String): AboutPage! @auth",
	} {
		if !strings.Contains(sdl, want) {
			t.Errorf("single SDL missing %q\n\nFull SDL:\n%s", want, sdl)
		}
	}

	if strings.Contains(sdl, "deleteAboutPage") {
		t.Error("single-type should not have delete mutation")
	}
	if strings.Contains(sdl, "Filter") {
		t.Error("single-type should not have filter type")
	}
	if strings.Contains(sdl, "OrderBy") {
		t.Error("single-type should not have orderBy type")
	}
}

func TestBuildContentTypeSDL_ComponentRepeatable(t *testing.T) {
	def := contenttype.ContentTypeDefinition{
		Slug: "cv-page",
		Kind: "collection",
		Fields: []entity.FieldDefinition{
			{Name: "position", Type: "text"},
			{Name: "skills", Type: "component", Repeatable: true, Fields: []entity.FieldDefinition{
				{Name: "level", Type: "text"},
				{Name: "skill", Type: "text"},
			}},
			{Name: "banner", Type: "component", Repeatable: false, Fields: []entity.FieldDefinition{
				{Name: "title", Type: "text"},
			}},
		},
	}
	sdl := buildContentTypeSDL(def)

	for _, want := range []string{
		"type CvPageSkills {",
		"skills: [CvPageSkills!]",
		"type CvPageBanner {",
		"banner: CvPageBanner",
	} {
		if !strings.Contains(sdl, want) {
			t.Errorf("component SDL missing %q\n\nFull SDL:\n%s", want, sdl)
		}
	}

	if strings.Contains(sdl, "banner: [CvPageBanner") {
		t.Errorf("non-repeatable banner should not be a list type")
	}
}

func TestBuildContentTypeSDL_MediaInput(t *testing.T) {
	def := contenttype.ContentTypeDefinition{
		Slug: "contact",
		Kind: "single",
		Fields: []entity.FieldDefinition{
			{Name: "avatar", Type: "media"},
			{Name: "name", Type: "text"},
		},
	}
	sdl := buildContentTypeSDL(def)

	if !strings.Contains(sdl, "avatar: MediaAsset") {
		t.Error("output type for media should be MediaAsset")
	}
	if !strings.Contains(sdl, "avatar: String") {
		t.Error("input type for media should be String (documentId)")
	}
}

func TestSlugToPascalCase(t *testing.T) {
	tests := []struct {
		slug string
		want string
	}{
		{"blog-posts", "BlogPosts"},
		{"about-page", "AboutPage"},
		{"homepage", "Homepage"},
		{"common-text", "CommonText"},
	}
	for _, test := range tests {
		got := slugToPascalCase(test.slug)
		if got != test.want {
			t.Errorf("slugToPascalCase(%q) = %q, want %q", test.slug, got, test.want)
		}
	}
}

func TestSlugToCamelCase(t *testing.T) {
	tests := []struct {
		slug string
		want string
	}{
		{"blog-posts", "blogPosts"},
		{"about-page", "aboutPage"},
		{"homepage", "homepage"},
	}
	for _, test := range tests {
		got := slugToCamelCase(test.slug)
		if got != test.want {
			t.Errorf("slugToCamelCase(%q) = %q, want %q", test.slug, got, test.want)
		}
	}
}

func TestFieldTypeToGraphQL(t *testing.T) {
	tests := []struct {
		fieldType string
		want      string
	}{
		{"text", "String"},
		{"richtext", "String"},
		{"number", "Float"},
		{"boolean", "Boolean"},
		{"media", "MediaAsset"},
		{"json", "JSON"},
	}
	for _, test := range tests {
		got := fieldTypeToGraphQL(test.fieldType)
		if got != test.want {
			t.Errorf("fieldTypeToGraphQL(%q) = %q, want %q", test.fieldType, got, test.want)
		}
	}
}

func TestFieldTypeToInputGraphQL_MediaIsString(t *testing.T) {
	got := fieldTypeToInputGraphQL("media")
	if got != "String" {
		t.Errorf("fieldTypeToInputGraphQL(media) = %q, want String", got)
	}
}
