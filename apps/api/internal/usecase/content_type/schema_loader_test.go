package content_type_test

import (
	"testing"

	contenttype "project-abyssoftime-cms-v2/api/internal/usecase/content_type"
)

func TestLoadDefinitions_Valid(t *testing.T) {
	defs, err := contenttype.LoadDefinitions("testdata/valid")
	if err != nil {
		t.Fatalf("LoadDefinitions() error = %v", err)
	}
	if len(defs) != 2 {
		t.Fatalf("LoadDefinitions() count = %d, want 2", len(defs))
	}

	bySlug := map[string]contenttype.ContentTypeDefinition{}
	for _, d := range defs {
		bySlug[d.Slug] = d
	}

	homepage, ok := bySlug["homepage"]
	if !ok {
		t.Fatalf("expected definition with slug 'homepage', got %+v", bySlug)
	}
	if homepage.Name != "Homepage" {
		t.Errorf("homepage.Name = %q, want %q", homepage.Name, "Homepage")
	}
	if homepage.Kind != "single" {
		t.Errorf("homepage.Kind = %q, want %q", homepage.Kind, "single")
	}
	if len(homepage.Fields) != 2 {
		t.Fatalf("homepage.Fields count = %d, want 2", len(homepage.Fields))
	}
	if homepage.Fields[0].Name != "title" || homepage.Fields[0].Type != "text" {
		t.Errorf("homepage.Fields[0] = %+v, want {title text}", homepage.Fields[0])
	}

	blogPost, ok := bySlug["blog-post"]
	if !ok {
		t.Fatalf("expected definition with slug 'blog-post', got %+v", bySlug)
	}
	if blogPost.Kind != "collection" {
		t.Errorf("blogPost.Kind = %q, want %q", blogPost.Kind, "collection")
	}
	if len(blogPost.ListFields) != 0 {
		t.Errorf("blogPost.ListFields = %v, want []", blogPost.ListFields)
	}
}

func TestLoadDefinitions_ListFieldsIgnored(t *testing.T) {
	defs, err := contenttype.LoadDefinitions("testdata/valid-with-listfields")
	if err != nil {
		t.Fatalf("LoadDefinitions() should not error when listFields is present in JSON, got %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("LoadDefinitions() count = %d, want 1", len(defs))
	}
}

func TestLoadDefinitions_Malformed(t *testing.T) {
	_, err := contenttype.LoadDefinitions("testdata/malformed")
	if err == nil {
		t.Fatal("LoadDefinitions() error = nil, want non-nil for malformed JSON")
	}
}

func TestLoadDefinitions_MissingDir(t *testing.T) {
	_, err := contenttype.LoadDefinitions("testdata/does-not-exist")
	if err == nil {
		t.Fatal("LoadDefinitions() error = nil, want non-nil for missing directory")
	}
}

func TestLoadDefinitions_LayoutEmptyFields(t *testing.T) {
	_, err := contenttype.LoadDefinitions("testdata/invalid/layout-empty-fields")
	if err == nil {
		t.Fatal("LoadDefinitions() error = nil, want error for layout with empty fields")
	}
}

func TestLoadDefinitions_LayoutContainsComponent(t *testing.T) {
	_, err := contenttype.LoadDefinitions("testdata/invalid/layout-contains-component")
	if err == nil {
		t.Fatal("LoadDefinitions() error = nil, want error for layout containing a component")
	}
}

func TestLoadDefinitions_ComponentEmptyName(t *testing.T) {
	_, err := contenttype.LoadDefinitions("testdata/invalid/component-empty-name")
	if err == nil {
		t.Fatal("LoadDefinitions() error = nil, want error for component with empty name")
	}
}

func TestLoadDefinitions_ComponentDepthExceeded(t *testing.T) {
	_, err := contenttype.LoadDefinitions("testdata/invalid/component-depth-exceeded")
	if err == nil {
		t.Fatal("LoadDefinitions() error = nil, want error for component depth > 2")
	}
}
