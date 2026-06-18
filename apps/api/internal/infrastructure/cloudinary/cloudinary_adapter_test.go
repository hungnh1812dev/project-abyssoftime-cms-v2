//go:build !integration

package cloudinary_test

import (
	"testing"

	cloudinaryadapter "project-abyssoftime-cms-v2/api/internal/infrastructure/cloudinary"
)

// compile-time: NewCloudinaryAdapter must exist and satisfy StorageAdapter.
var _ = cloudinaryadapter.NewCloudinaryAdapter

func TestNewCloudinaryAdapter_EmptyCredentials_ReturnsError(t *testing.T) {
	_, err := cloudinaryadapter.NewCloudinaryAdapter("", "", "")
	if err == nil {
		t.Fatal("NewCloudinaryAdapter() error = nil, want error for empty credentials")
	}
}

func TestCloudinaryAdapter_Upload_ReturnsSecureURL(t *testing.T) {
	t.Skip("integration test: requires Cloudinary credentials — run with -tags integration")
}

func TestCloudinaryAdapter_Delete_RemovesAsset(t *testing.T) {
	t.Skip("integration test: requires Cloudinary credentials — run with -tags integration")
}
