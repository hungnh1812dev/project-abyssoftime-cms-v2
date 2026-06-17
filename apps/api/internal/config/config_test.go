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
		"SUPPORTED_LOCALES", "MEDIA_AUTO_THUMBNAIL", "GRAPHQL_PATH", "DB_DRIVER",
		"POSTGRES_URI",
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
		Port:             "8080",
		JWTSecret:        "test-secret",
		ContentTypeDir:   "content-types",
		SupportedLocales: []string{"en", "vi"},
		DB: config.DBConfig{
			Driver: "mongo",
			Mongo:  config.MongoConfig{URI: "mongodb://localhost:27017", Name: "cms"},
		},
		Media: config.MediaConfig{
			Driver:            "cloudinary",
			GenerateThumbnail: true,
		},
		GraphQL: config.GraphQLConfig{Path: "/graphql"},
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
	if cfg.Media.Driver != "s3" {
		t.Errorf("Media.Driver = %q, want %q", cfg.Media.Driver, "s3")
	}
	if cfg.Media.S3.Bucket != "my-bucket" || cfg.Media.S3.Region != "us-east-1" {
		t.Errorf("Media.S3.Bucket/Region = %q/%q, want my-bucket/us-east-1", cfg.Media.S3.Bucket, cfg.Media.S3.Region)
	}
	wantLocales := []string{"en", "vi", "fr"}
	if !reflect.DeepEqual(cfg.SupportedLocales, wantLocales) {
		t.Errorf("SupportedLocales = %v, want %v", cfg.SupportedLocales, wantLocales)
	}
	if cfg.Media.GenerateThumbnail != false {
		t.Errorf("Media.GenerateThumbnail = %v, want false", cfg.Media.GenerateThumbnail)
	}
	if cfg.GraphQL.Path != "/api/graphql" {
		t.Errorf("GraphQL.Path = %q, want %q", cfg.GraphQL.Path, "/api/graphql")
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	clearEnv(t)

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for missing JWT_SECRET")
	}
}

func TestLoad_InvalidMediaDriver(t *testing.T) {
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

func TestLoad_CloudinaryMissingCreds(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", "s")
	t.Setenv("STORAGE_PROVIDER", "cloudinary")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for missing Cloudinary credentials")
	}
}

func TestLoad_S3MissingBucket(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", "s")
	t.Setenv("STORAGE_PROVIDER", "s3")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for missing S3_BUCKET")
	}
}

func TestLoad_S3MissingRegion(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", "s")
	t.Setenv("STORAGE_PROVIDER", "s3")
	t.Setenv("S3_BUCKET", "my-bucket")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for missing S3_REGION")
	}
}
