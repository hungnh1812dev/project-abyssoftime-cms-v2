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
	Port                string
	MongoDBURI          string
	MongoDBDB           string
	JWTSecret           string
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
	ContentTypesDir     string
	StorageProvider     string
	S3Bucket            string
	S3Region            string
	SupportedLocales    []string
	MediaAutoThumbnail  bool
	GraphQLPath         string
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

	storageProvider := getenv("STORAGE_PROVIDER", "cloudinary")
	if storageProvider != "s3" && storageProvider != "cloudinary" {
		return nil, fmt.Errorf("unknown STORAGE_PROVIDER %q (want %q or %q)", storageProvider, "s3", "cloudinary")
	}

	mediaAutoThumbnail := true
	if raw := os.Getenv("MEDIA_AUTO_THUMBNAIL"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid MEDIA_AUTO_THUMBNAIL %q: %w", raw, err)
		}
		mediaAutoThumbnail = v
	}

	return &Config{
		Port:                getenv("PORT", "8080"),
		MongoDBURI:          getenv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBDB:           getenv("MONGODB_DB", "cms"),
		JWTSecret:           jwtSecret,
		CloudinaryCloudName: os.Getenv("CLOUDINARY_CLOUD_NAME"),
		CloudinaryAPIKey:    os.Getenv("CLOUDINARY_API_KEY"),
		CloudinaryAPISecret: os.Getenv("CLOUDINARY_API_SECRET"),
		ContentTypesDir:     getenv("CONTENT_TYPES_DIR", "content-types"),
		StorageProvider:     storageProvider,
		S3Bucket:            os.Getenv("S3_BUCKET"),
		S3Region:            os.Getenv("S3_REGION"),
		SupportedLocales:    splitLocales(getenv("SUPPORTED_LOCALES", "en,vi")),
		MediaAutoThumbnail:  mediaAutoThumbnail,
		GraphQLPath:         getenv("GRAPHQL_PATH", "/graphql"),
	}, nil
}
