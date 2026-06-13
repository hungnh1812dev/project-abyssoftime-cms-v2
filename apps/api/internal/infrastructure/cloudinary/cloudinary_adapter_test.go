//go:build !integration

package cloudinary_test

import (
	"testing"

	cloudinaryadapter "project-abyssoftime-cms-v2/api/internal/infrastructure/cloudinary"
)

// compile-time: NewCloudinaryAdapter must exist and satisfy StorageAdapter.
var _ = cloudinaryadapter.NewCloudinaryAdapter

func TestCloudinaryAdapter_Upload_ReturnsSecureURL(t *testing.T) {
	t.Skip("integration test: requires Cloudinary credentials — run with -tags integration")
}

func TestCloudinaryAdapter_Delete_RemovesAsset(t *testing.T) {
	t.Skip("integration test: requires Cloudinary credentials — run with -tags integration")
}
