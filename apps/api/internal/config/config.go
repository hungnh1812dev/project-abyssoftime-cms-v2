// Package config centralizes environment-variable loading. Every other
// package receives configuration through a Config value built here by
// Load() — no os.Getenv calls anywhere else in the codebase.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port             string
	JWTSecret        string
	ContentTypeDir   string // was ContentTypesDir; env var CONTENT_TYPES_DIR unchanged
	SupportedLocales []string
	DB               DBConfig
	Media            MediaConfig
	GraphQL          GraphQLConfig
}

type DBConfig struct {
	Driver     string // DB_DRIVER, default "mongo"
	Mongo      MongoConfig
	Postgresql PostgresConfig
}

type MongoConfig struct {
	URI  string // MONGODB_URI, default "mongodb://localhost:27017"
	Name string // MONGODB_DB,  default "cms"
}

type PostgresConfig struct {
	URI string // POSTGRES_URI — placeholder, not validated in v1
}

type MediaConfig struct {
	Driver            string // STORAGE_PROVIDER, default "cloudinary"
	GenerateThumbnail bool   // MEDIA_AUTO_THUMBNAIL, default true
	Cloudinary        CloudinaryConfig
	S3                S3Config
}

type CloudinaryConfig struct {
	CloudName string // CLOUDINARY_CLOUD_NAME
	APIKey    string // CLOUDINARY_API_KEY
	APISecret string // CLOUDINARY_API_SECRET
}

type S3Config struct {
	Bucket string // S3_BUCKET
	Region string // S3_REGION
}

type GraphQLConfig struct {
	Path string // GRAPHQL_PATH, default "/graphql"
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitLocales(raw string) []string {
	parts := strings.Split(raw, ",")
	locales := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			locales = append(locales, p)
		}
	}
	return locales
}

// Load reads, defaults, and validates every environment variable the
// application needs, returning a single typed Config or the first
// validation error encountered.
func Load() (*Config, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	dbDriver := getenv("DB_DRIVER", "mongo")
	if dbDriver != "mongo" && dbDriver != "postgresql" {
		return nil, fmt.Errorf("unknown DB_DRIVER %q (want %q or %q)", dbDriver, "mongo", "postgresql")
	}

	mediaDriver := getenv("STORAGE_PROVIDER", "cloudinary")
	if mediaDriver != "s3" && mediaDriver != "cloudinary" {
		return nil, fmt.Errorf("unknown STORAGE_PROVIDER %q (want %q or %q)", mediaDriver, "s3", "cloudinary")
	}

	generateThumbnail := true
	if raw := os.Getenv("MEDIA_AUTO_THUMBNAIL"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid MEDIA_AUTO_THUMBNAIL %q: %w", raw, err)
		}
		generateThumbnail = v
	}

	cloudinaryCloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	cloudinaryAPIKey := os.Getenv("CLOUDINARY_API_KEY")
	cloudinaryAPISecret := os.Getenv("CLOUDINARY_API_SECRET")

	s3Bucket := os.Getenv("S3_BUCKET")
	s3Region := os.Getenv("S3_REGION")

	return &Config{
		Port:             getenv("PORT", "8080"),
		JWTSecret:        jwtSecret,
		ContentTypeDir:   getenv("CONTENT_TYPES_DIR", "content-types"),
		SupportedLocales: splitLocales(getenv("SUPPORTED_LOCALES", "en,vi")),
		DB: DBConfig{
			Driver: dbDriver,
			Mongo: MongoConfig{
				URI:  getenv("MONGODB_URI", "mongodb://localhost:27017"),
				Name: getenv("MONGODB_DB", "cms"),
			},
			Postgresql: PostgresConfig{
				URI: os.Getenv("POSTGRES_URI"),
			},
		},
		Media: MediaConfig{
			Driver:            mediaDriver,
			GenerateThumbnail: generateThumbnail,
			Cloudinary: CloudinaryConfig{
				CloudName: cloudinaryCloudName,
				APIKey:    cloudinaryAPIKey,
				APISecret: cloudinaryAPISecret,
			},
			S3: S3Config{
				Bucket: s3Bucket,
				Region: s3Region,
			},
		},
		GraphQL: GraphQLConfig{
			Path: getenv("GRAPHQL_PATH", "/graphql"),
		},
	}, nil
}
