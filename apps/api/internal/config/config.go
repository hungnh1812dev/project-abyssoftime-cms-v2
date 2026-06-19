// Package config centralizes environment-variable loading. Every other
// package receives configuration through a Config value built here by
// Load() — no os.Getenv calls anywhere else in the codebase.
package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port             string
	GRPCPort         string
	JWTSecret        string
	ContentTypeDir   string
	SupportedLocales []string
	DB               DBConfig
	Media            MediaConfig
	GraphQL          GraphQLConfig
}

type DBConfig struct {
	Driver   string // DB_DRIVER, default "mongo"
	Mongo    MongoConfig
	SQL      SQLConfig
	EntityDB EntityDBConfig
}

type MongoConfig struct {
	URI  string // MONGODB_URI, default "mongodb://localhost:27017"
	Name string // MONGODB_DB,  default "cms"
}

type SQLConfig struct {
	Driver string // SQL_DRIVER: "postgres" (default when any entity uses sql)
	DSN    string // SQL_DSN: full connection string
}

type EntityDBConfig struct {
	User        string // DB_USER, default: DB_DRIVER
	ContentType string // DB_CONTENT_TYPE, default: DB_DRIVER
	Document    string // DB_DOCUMENT, default: DB_DRIVER
	Media       string // DB_MEDIA, default: DB_DRIVER
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
	generateThumbnail := true
	if raw := os.Getenv("MEDIA_AUTO_THUMBNAIL"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			generateThumbnail = false
		} else{
		generateThumbnail = v}
	}

	return &Config{
		Port:             getenv("PORT", "8080"),
		GRPCPort:         getenv("GRPC_PORT", "9090"),
		JWTSecret:        getenv("JWT_SECRET", ""),
		ContentTypeDir:   getenv("CONTENT_TYPES_DIR", "content-types"),
		SupportedLocales: splitLocales(getenv("SUPPORTED_LOCALES", "en,vi")),
		DB: func() DBConfig {
			driver := getenv("DB_DRIVER", "mongo")
			return DBConfig{
				Driver: driver,
				Mongo: MongoConfig{
					URI:  getenv("MONGODB_URI", "mongodb://localhost:27017"),
					Name: getenv("MONGODB_DB", "cms"),
				},
				SQL: SQLConfig{
					Driver: getenv("SQL_DRIVER", "postgres"),
					DSN:    getenv("SQL_DSN", ""),
				},
				EntityDB: EntityDBConfig{
					User:        getenv("DB_USER", driver),
					ContentType: getenv("DB_CONTENT_TYPE", driver),
					Document:    getenv("DB_DOCUMENT", driver),
					Media:       getenv("DB_MEDIA", driver),
				},
			}
		}(),
		Media: MediaConfig{
			Driver:            getenv("STORAGE_PROVIDER", "cloudinary"),
			GenerateThumbnail: generateThumbnail,
			Cloudinary: CloudinaryConfig{
				CloudName: getenv("CLOUDINARY_CLOUD_NAME", ""),
				APIKey:    getenv("CLOUDINARY_API_KEY", ""),
				APISecret: getenv("CLOUDINARY_API_SECRET", ""),
			},
			S3: S3Config{
				Bucket: getenv("S3_BUCKET", ""),
				Region: getenv("S3_REGION", ""),
			},
		},
		GraphQL: GraphQLConfig{
			Path: getenv("GRAPHQL_PATH", "/graphql"),
		},
	}, nil
}
