package config_test

import (
	"reflect"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/config"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"PORT", "MONGODB_URI", "MONGODB_DB", "JWT_SECRET",
		"CLOUDINARY_CLOUD_NAME", "CLOUDINARY_API_KEY", "CLOUDINARY_API_SECRET",
		"CONTENT_TYPES_DIR", "STORAGE_PROVIDER", "S3_BUCKET", "S3_REGION",
		"SUPPORTED_LOCALES", "MEDIA_AUTO_THUMBNAIL", "GRAPHQL_PATH",
	} {
		t.Setenv(k, "")
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", "test-secret")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := &config.Config{
		Port:               "8080",
		MongoDBURI:         "mongodb://localhost:27017",
		MongoDBDB:          "cms",
		JWTSecret:          "test-secret",
		ContentTypesDir:    "content-types",
		StorageProvider:    "cloudinary",
		SupportedLocales:   []string{"en", "vi"},
		MediaAutoThumbnail: true,
		GraphQLPath:        "/graphql",
	}
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("Load() = %+v, want %+v", cfg, want)
	}
}

func TestLoad_OverridesAndLocaleSplit(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", "s")
	t.Setenv("PORT", "9090")
	t.Setenv("STORAGE_PROVIDER", "s3")
	t.Setenv("S3_BUCKET", "my-bucket")
	t.Setenv("S3_REGION", "us-east-1")
	t.Setenv("SUPPORTED_LOCALES", " en , vi ,fr,, ")
	t.Setenv("MEDIA_AUTO_THUMBNAIL", "false")
	t.Setenv("GRAPHQL_PATH", "/api/graphql")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9090")
	}
	if cfg.StorageProvider != "s3" {
		t.Errorf("StorageProvider = %q, want %q", cfg.StorageProvider, "s3")
	}
	if cfg.S3Bucket != "my-bucket" || cfg.S3Region != "us-east-1" {
		t.Errorf("S3Bucket/S3Region = %q/%q, want my-bucket/us-east-1", cfg.S3Bucket, cfg.S3Region)
	}
	wantLocales := []string{"en", "vi", "fr"}
	if !reflect.DeepEqual(cfg.SupportedLocales, wantLocales) {
		t.Errorf("SupportedLocales = %v, want %v", cfg.SupportedLocales, wantLocales)
	}
	if cfg.MediaAutoThumbnail != false {
		t.Errorf("MediaAutoThumbnail = %v, want false", cfg.MediaAutoThumbnail)
	}
	if cfg.GraphQLPath != "/api/graphql" {
		t.Errorf("GraphQLPath = %q, want %q", cfg.GraphQLPath, "/api/graphql")
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	clearEnv(t)

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for missing JWT_SECRET")
	}
}

func TestLoad_InvalidStorageProvider(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", "s")
	t.Setenv("STORAGE_PROVIDER", "dropbox")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for invalid STORAGE_PROVIDER")
	}
}

func TestLoad_InvalidMediaAutoThumbnail(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", "s")
	t.Setenv("MEDIA_AUTO_THUMBNAIL", "not-a-bool")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for invalid MEDIA_AUTO_THUMBNAIL")
	}
}
