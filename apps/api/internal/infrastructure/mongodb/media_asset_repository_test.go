//go:build !integration

package mongodb_test

import (
	"context"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/infrastructure/mongodb"
)

// compile-time: NewMediaAssetRepository must exist and satisfy the interface.
var _ = mongodb.NewMediaAssetRepository

func TestMediaAssetRepository_Create_SetsID(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestMediaAssetRepository_FindByID_NotFound(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestMediaAssetRepository_FindByDocumentRef_Empty(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestMediaAssetRepository_DeleteByDocumentRef(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

func TestMediaAssetRepository_Delete_NotFound(t *testing.T) {
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}

var mediaAssetRepoTests = []struct {
	name        string
	documentRef string
	url         string
}{
	{"profile image", "doc-1", "https://res.cloudinary.com/demo/image/upload/sample.jpg"},
	{"banner", "doc-2", "https://res.cloudinary.com/demo/image/upload/banner.png"},
}

func TestMediaAssetRepository_TableDriven(t *testing.T) {
	_ = context.Background
	_ = entity.MediaAsset{}
	_ = mediaAssetRepoTests
	t.Skip("integration test: requires live MongoDB — run with -tags integration")
}
