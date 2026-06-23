package dynamic

import (
	"strings"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	contenttype "project-abyssoftime-cms-v2/api/internal/usecase/content_type"
)

func TestBuildBaseSchema(t *testing.T) {
	b := NewSchemaBuilder(nil)
	sdl := b.BuildBaseSchema()

	for _, want := range []string{
		"scalar JSON",
		"scalar Time",
		"enum SortOrder {",
		"ASC",
		"DESC",
		"type ContentType {",
		"id: ID!",
		"name: String!",
		"slug: String!",
		"kind: String!",
		"createdAt: Time!",
		"updatedAt: Time!",
		"type Query {",
		"contentTypes: [ContentType!]!",
	} {
		if !strings.Contains(sdl, want) {
			t.Errorf("base schema missing %q", want)
		}
	}
}

func TestBuildBaseSchema_DoesNotContainMutationType(t *testing.T) {
	b := NewSchemaBuilder(nil)
	sdl := b.BuildBaseSchema()
	if strings.Contains(sdl, "type Mutation") {
		t.Error("base schema should not define type Mutation (added per content-type)")
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
	for _, tt := range tests {
		got := slugToPascalCase(tt.slug)
		if got != tt.want {
			t.Errorf("slugToPascalCase(%q) = %q, want %q", tt.slug, got, tt.want)
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
	for _, tt := range tests {
		got := slugToCamelCase(tt.slug)
		if got != tt.want {
			t.Errorf("slugToCamelCase(%q) = %q, want %q", tt.slug, got, tt.want)
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
		{"unknown", "String"},
	}
	for _, tt := range tests {
		got := fieldTypeToGraphQL(tt.fieldType)
		if got != tt.want {
			t.Errorf("fieldTypeToGraphQL(%q) = %q, want %q", tt.fieldType, got, tt.want)
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
	b := NewSchemaBuilder(nil)
	sdl := b.BuildContentTypeSDL(def)

	for _, want := range []string{
		"type BlogPosts {",
		"title: String",
		"featured: Boolean",
		"readingTime: Float",
		"documentId: ID!",
		"locale: String!",
		"createdAt: Time!",
		"updatedAt: Time!",
		"input BlogPostsInput {",
		"input BlogPostsFilter {",
		"title: StringFilter",
		"featured: BooleanFilter",
		"readingTime: NumberFilter",
		"AND: [BlogPostsFilter!]",
		"OR: [BlogPostsFilter!]",
		"NOT: BlogPostsFilter",
		"input BlogPostsOrderBy {",
		"title: SortOrder",
		"createdAt: SortOrder",
		"extend type Query {",
		"blogPosts(blogPostsId: ID!, locale: String): BlogPosts",
		"blogPostsList(where: BlogPostsFilter, orderBy: BlogPostsOrderBy, start: Int, size: Int, locale: String): [BlogPosts!]!",
		"extend type Mutation {",
		"createBlogPosts(data: BlogPostsInput!): BlogPosts!",
		"updateBlogPosts(blogPostsId: ID!, data: BlogPostsInput!): BlogPosts!",
		"deleteBlogPosts(blogPostsId: ID!): Boolean!",
		"publishBlogPosts(blogPostsId: ID!, locale: String): BlogPosts!",
		"unpublishBlogPosts(blogPostsId: ID!, locale: String): BlogPosts!",
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
	b := NewSchemaBuilder(nil)
	sdl := b.BuildContentTypeSDL(def)

	for _, want := range []string{
		"type AboutPage {",
		"headline: String",
		"openToWork: Boolean",
		"aboutPage(locale: String): AboutPage",
		"saveAboutPage(data: AboutPageInput!, locale: String): AboutPage!",
		"publishAboutPage(locale: String): AboutPage!",
		"unpublishAboutPage(locale: String): AboutPage!",
	} {
		if !strings.Contains(sdl, want) {
			t.Errorf("single SDL missing %q\n\nFull SDL:\n%s", want, sdl)
		}
	}

	if strings.Contains(sdl, "deleteAboutPage") {
		t.Error("single-type should not have delete mutation")
	}
}

func TestBuildContentTypeSDL_Component(t *testing.T) {
	def := contenttype.ContentTypeDefinition{
		Slug: "blog-posts",
		Name: "Blog Posts",
		Kind: "collection",
		Fields: []entity.FieldDefinition{
			{Name: "title", Type: "text"},
			{Name: "banner", Type: "component", Fields: []entity.FieldDefinition{
				{Name: "background", Type: "media"},
				{Name: "title", Type: "text"},
			}},
		},
	}
	b := NewSchemaBuilder(nil)
	sdl := b.BuildContentTypeSDL(def)

	for _, want := range []string{
		"type BlogPostsBanner {",
		"background: MediaAsset",
		"title: String",
		"banner: BlogPostsBanner",
	} {
		if !strings.Contains(sdl, want) {
			t.Errorf("component SDL missing %q\n\nFull SDL:\n%s", want, sdl)
		}
	}
}

func TestBuildContentTypeSDL_RepeatableComponent(t *testing.T) {
	def := contenttype.ContentTypeDefinition{
		Slug: "cv-page",
		Name: "CV Page",
		Kind: "collection",
		Fields: []entity.FieldDefinition{
			{Name: "position", Type: "text"},
			{Name: "skills", Type: "component", Repeatable: true, Fields: []entity.FieldDefinition{
				{Name: "category", Type: "text"},
				{Name: "skill", Type: "text"},
			}},
			{Name: "banner", Type: "component", Repeatable: false, Fields: []entity.FieldDefinition{
				{Name: "title", Type: "text"},
			}},
		},
	}
	b := NewSchemaBuilder(nil)
	sdl := b.BuildContentTypeSDL(def)

	for _, want := range []string{
		"type CvPageSkills {",
		"category: String",
		"skills: [CvPageSkills!]",
		"type CvPageBanner {",
		"banner: CvPageBanner",
	} {
		if !strings.Contains(sdl, want) {
			t.Errorf("repeatable component SDL missing %q\n\nFull SDL:\n%s", want, sdl)
		}
	}

	if strings.Contains(sdl, "banner: [CvPageBanner") {
		t.Errorf("non-repeatable banner should not be a list type\n\nFull SDL:\n%s", sdl)
	}
}

func TestBuildSDL_MergesAllDefinitions(t *testing.T) {
	defs := []contenttype.ContentTypeDefinition{
		{
			Slug: "about-page", Name: "About Page", Kind: "single",
			Fields: []entity.FieldDefinition{{Name: "headline", Type: "text"}},
		},
		{
			Slug: "blog-posts", Name: "Blog Posts", Kind: "collection",
			Fields: []entity.FieldDefinition{{Name: "title", Type: "text"}},
		},
		{
			Slug: "common-text", Name: "Common Text", Kind: "single",
			Fields: []entity.FieldDefinition{{Name: "text", Type: "json"}},
		},
	}
	b := NewSchemaBuilder(defs)
	sdl := b.BuildSDL()

	for _, want := range []string{
		"scalar JSON",
		"type MediaAsset {",
		"type ContentType {",
		"contentTypes: [ContentType!]!",
		"type AboutPage {",
		"aboutPage(locale: String): AboutPage",
		"saveAboutPage(",
		"type BlogPosts {",
		"createBlogPosts(",
		"[BlogPosts!]!",
		"type CommonText {",
		"saveCommonText(",
	} {
		if !strings.Contains(sdl, want) {
			t.Errorf("merged SDL missing %q", want)
		}
	}
}
