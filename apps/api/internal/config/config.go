// Package config centralizes environment-variable loading. Every other
// package receives configuration through a Config value built here by
// Load() — no os.Getenv calls anywhere else in the codebase.
package config

import (
	"fmt"
	"net/url"
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
	Driver   string // DB_DRIVER: "mongo" | "postgres", default "mongo"
	Host     string // DB_HOST
	Port     string // DB_PORT (default: 27017 for mongo, 5432 for postgres)
	Name     string // DB_NAME, default "cms"
	Username string // DB_USERNAME
	Password string // DB_PASSWORD
	SSLMode  string // DB_SSL_MODE: "disable" | "require", default "disable"
	EntityDB EntityDBConfig
}

// MongoURI builds the MongoDB connection string from generic DB_* vars.
// Username and password are URL-encoded to handle special characters.
func (d DBConfig) MongoURI() string {
	host := d.Host
	if host == "" {
		host = "localhost"
	}
	port := d.Port
	if port == "" {
		port = "27017"
	}
	u := &url.URL{
		Scheme: "mongodb",
		Host:   fmt.Sprintf("%s:%s", host, port),
	}
	if d.Username != "" && d.Password != "" {
		u.User = url.UserPassword(d.Username, d.Password)
	}
	return u.String()
}

// PostgresDSN builds the PostgreSQL connection string from generic DB_* vars.
// Username and password are URL-encoded to handle special characters (@, ?, etc.).
func (d DBConfig) PostgresDSN() string {
	host := d.Host
	if host == "" {
		host = "localhost"
	}
	port := d.Port
	if port == "" {
		port = "5432"
	}
	name := d.Name
	if name == "" {
		name = "cms"
	}
	sslMode := d.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	u := &url.URL{
		Scheme:   "postgres",
		Host:     fmt.Sprintf("%s:%s", host, port),
		Path:     name,
		RawQuery: fmt.Sprintf("sslmode=%s", sslMode),
	}
	if d.Username != "" {
		u.User = url.UserPassword(d.Username, d.Password)
	}
	return u.String()
}

type EntityDBConfig struct {
	User        string // DB_ENTITY_USER, default: DB_DRIVER
	ContentType string // DB_ENTITY_CONTENT_TYPE, default: DB_DRIVER
	Document    string // DB_ENTITY_DOCUMENT, default: DB_DRIVER
	Media       string // DB_ENTITY_MEDIA, default: DB_DRIVER
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
		} else {
			generateThumbnail = v
		}
	}

	driver := getenv("DB_DRIVER", "mongo")

	defaultPort := "27017"
	if driver == "postgres" {
		defaultPort = "5432"
	}

	return &Config{
		Port:             getenv("PORT", "8080"),
		GRPCPort:         getenv("GRPC_PORT", "9090"),
		JWTSecret:        getenv("JWT_SECRET", ""),
		ContentTypeDir:   getenv("CONTENT_TYPES_DIR", "content-types"),
		SupportedLocales: splitLocales(getenv("SUPPORTED_LOCALES", "en,vi")),
		DB: DBConfig{
			Driver:   driver,
			Host:     getenv("DB_HOST", "localhost"),
			Port:     getenv("DB_PORT", defaultPort),
			Name:     getenv("DB_NAME", "cms"),
			Username: getenv("DB_USERNAME", ""),
			Password: getenv("DB_PASSWORD", ""),
			SSLMode:  getenv("DB_SSL_MODE", "disable"),
			EntityDB: EntityDBConfig{
				User:        getenv("DB_ENTITY_USER", driver),
				ContentType: getenv("DB_ENTITY_CONTENT_TYPE", driver),
				Document:    getenv("DB_ENTITY_DOCUMENT", driver),
				Media:       getenv("DB_ENTITY_MEDIA", driver),
			},
		},
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
