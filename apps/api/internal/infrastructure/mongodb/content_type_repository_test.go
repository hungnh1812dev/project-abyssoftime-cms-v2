//go:build !integration

package mongodb_test

import (
	"context"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/infrastructure/mongodb"
)

// compile-time: NewContentTypeRepository must exist and return the interface.
var _ = mongodb.NewContentTypeRepository

func TestContentTypeRepository_Create_SetsID(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestContentTypeRepository_FindByID_NotFound(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestContentTypeRepository_FindBySlug_NotFound(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestContentTypeRepository_FindAll_Empty(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestContentTypeRepository_Update_NotFound(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestContentTypeRepository_Delete_NotFound(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

var contentTypeRepoTests = []struct {
	name string
	slug string
	kind entity.ContentKind
}{
	{"blog post", "blog-post", entity.KindCollection},
	{"homepage", "homepage", entity.KindSingle},
}

func TestContentTypeRepository_TableDriven(t *testing.T) {
	_ = context.Background
	_ = contentTypeRepoTests
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}
