package config_test

import (
	"reflect"
	"testing"

	"project-abyssoftime-cms-v2/api/internal/config"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"PORT", "JWT_SECRET",
		"CLOUDINARY_CLOUD_NAME", "CLOUDINARY_API_KEY", "CLOUDINARY_API_SECRET",
		"CONTENT_TYPES_DIR", "STORAGE_PROVIDER", "S3_BUCKET", "S3_REGION",
		"SUPPORTED_LOCALES", "MEDIA_AUTO_THUMBNAIL", "GRAPHQL_PATH",
		"DB_DRIVER", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USERNAME", "DB_PASSWORD", "DB_SSL_MODE",
		"DB_ENTITY_USER", "DB_ENTITY_CONTENT_TYPE", "DB_ENTITY_DOCUMENT", "DB_ENTITY_MEDIA",
		"GRPC_PORT",
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
		GRPCPort:         "9090",
		JWTSecret:        "test-secret",
		ContentTypeDir:   "content-types",
		SupportedLocales: []string{"en", "vi"},
		DB: config.DBConfig{
			Driver:  "mongo",
			Host:    "localhost",
			Port:    "27017",
			Name:    "cms",
			SSLMode: "disable",
			EntityDB: config.EntityDBConfig{
				User: "mongo", ContentType: "mongo", Document: "mongo", Media: "mongo",
			},
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

func TestLoad_PostgresDefaults(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", "s")
	t.Setenv("DB_DRIVER", "postgres")
	t.Setenv("DB_USERNAME", "cms_user")
	t.Setenv("DB_PASSWORD", "secret")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DB.Port != "5432" {
		t.Errorf("DB.Port = %q, want %q (postgres default)", cfg.DB.Port, "5432")
	}
	if cfg.DB.EntityDB.User != "postgres" {
		t.Errorf("EntityDB.User = %q, want %q", cfg.DB.EntityDB.User, "postgres")
	}

	wantDSN := "postgres://cms_user:secret@localhost:5432/cms?sslmode=disable"
	if cfg.DB.PostgresDSN() != wantDSN {
		t.Errorf("PostgresDSN() = %q, want %q", cfg.DB.PostgresDSN(), wantDSN)
	}
}

func TestLoad_MongoURI(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", "s")
	t.Setenv("DB_HOST", "mongo.example.com")
	t.Setenv("DB_PORT", "27018")
	t.Setenv("DB_USERNAME", "admin")
	t.Setenv("DB_PASSWORD", "pass")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	wantURI := "mongodb://admin:pass@mongo.example.com:27018"
	if cfg.DB.MongoURI() != wantURI {
		t.Errorf("MongoURI() = %q, want %q", cfg.DB.MongoURI(), wantURI)
	}
}

func TestLoad_MongoURI_NoAuth(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", "s")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	wantURI := "mongodb://localhost:27017"
	if cfg.DB.MongoURI() != wantURI {
		t.Errorf("MongoURI() = %q, want %q", cfg.DB.MongoURI(), wantURI)
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
