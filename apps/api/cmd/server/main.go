package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	deliveryhandler "project-abyssoftime-cms-v2/api/internal/delivery/http/handler"
	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	cloudinaryadapter "project-abyssoftime-cms-v2/api/internal/infrastructure/cloudinary"
	"project-abyssoftime-cms-v2/api/internal/infrastructure/mongodb"
	s3adapter "project-abyssoftime-cms-v2/api/internal/infrastructure/s3"
	"project-abyssoftime-cms-v2/api/internal/usecase/auth"
	contenttype "project-abyssoftime-cms-v2/api/internal/usecase/content_type"
	docuc "project-abyssoftime-cms-v2/api/internal/usecase/document"
	mediauc "project-abyssoftime-cms-v2/api/internal/usecase/media"
)

// resolveStorageProvider validates and defaults the STORAGE_PROVIDER env
// var. Pure function — kept separate from adapter construction so it's
// unit-testable without touching AWS or Cloudinary.
func resolveStorageProvider(envValue string) (string, error) {
	if envValue == "" {
		return "cloudinary", nil
	}
	if envValue != "s3" && envValue != "cloudinary" {
		return "", fmt.Errorf("unknown STORAGE_PROVIDER %q (want %q or %q)", envValue, "s3", "cloudinary")
	}
	return envValue, nil
}

func newStorageAdapter(ctx context.Context, provider string) (repository.StorageAdapter, error) {
	switch provider {
	case "s3":
		return s3adapter.New(ctx, os.Getenv("S3_BUCKET"), os.Getenv("S3_REGION"))
	default:
		return cloudinaryadapter.NewCloudinaryAdapter(
			os.Getenv("CLOUDINARY_CLOUD_NAME"),
			os.Getenv("CLOUDINARY_API_KEY"),
			os.Getenv("CLOUDINARY_API_SECRET"),
		)
	}
}

func main() {
	ctx := context.Background()

	mongoClient, err := mongodb.NewClient(ctx)
	if err != nil {
		log.Fatalf("mongodb connect: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("mongodb disconnect: %v", err)
		}
	}()
	log.Println("connected to mongodb")

	db := mongodb.Database(mongoClient)

	if err := mongodb.EnsureIndexes(ctx, db); err != nil {
		log.Fatalf("ensure indexes: %v", err)
	}
	log.Println("indexes ensured")

	// repositories
	userRepo := mongodb.NewUserRepository(db)
	ctRepo := mongodb.NewContentTypeRepository(db)
	docRepo := mongodb.NewDocumentRepository(db)
	mediaRepo := mongodb.NewMediaAssetRepository(db)

	// storage adapter: env-selected, S3 or Cloudinary, behind the same interface
	storageProvider, err := resolveStorageProvider(os.Getenv("STORAGE_PROVIDER"))
	if err != nil {
		log.Fatal(err)
	}
	storage, err := newStorageAdapter(ctx, storageProvider)
	if err != nil {
		log.Fatalf("%s init: %v", storageProvider, err)
	}
	log.Printf("storage provider: %s", storageProvider)

	// usecases
	authUC := auth.New(userRepo)
	ctUC := contenttype.New(ctRepo)
	documentUC := docuc.New(docRepo, mediaRepo)
	mediaUC := mediauc.New(mediaRepo, storage)

	// content-type schema-as-code sync: JSON definitions are the source of
	// truth, reconciled into Mongo before the server starts accepting traffic.
	defsDir := os.Getenv("CONTENT_TYPES_DIR")
	if defsDir == "" {
		defsDir = "content-types"
	}
	defs, err := contenttype.LoadDefinitions(defsDir)
	if err != nil {
		log.Fatalf("load content-type definitions: %v", err)
	}
	if err := contenttype.NewSyncer(ctUC, documentUC).Sync(ctx, defs); err != nil {
		log.Fatalf("sync content types: %v", err)
	}
	log.Printf("synced %d content-type definitions from %s", len(defs), defsDir)

	// handlers
	authHandler := deliveryhandler.NewAuthHandler(authUC)
	ctHandler := deliveryhandler.NewContentTypeHandler(ctUC)
	docHandler := deliveryhandler.NewDocumentHandler(documentUC)
	mediaHandler := deliveryhandler.NewMediaHandler(mediaUC)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	mux.HandleFunc("GET /auth/setup", authHandler.SetupStatus)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)

	adminOnly := func(h http.HandlerFunc) http.Handler {
		return middleware.Auth(middleware.RequireRole("admin", h))
	}
	mux.Handle("GET /api/content-types", adminOnly(ctHandler.List))
	mux.Handle("GET /api/content-types/{id}", adminOnly(ctHandler.GetByID))

	authRequired := func(h http.HandlerFunc) http.Handler {
		return middleware.Auth(http.HandlerFunc(h))
	}
	mux.Handle("GET /api/documents", authRequired(docHandler.List))
	mux.Handle("GET /api/documents/{id}", authRequired(docHandler.GetByID))
	mux.Handle("POST /api/documents", adminOnly(docHandler.Create))
	mux.Handle("PUT /api/documents/{id}", adminOnly(docHandler.Update))
	mux.Handle("DELETE /api/documents/{id}", adminOnly(docHandler.Delete))
	mux.Handle("POST /api/documents/{id}/publish", adminOnly(docHandler.Publish))
	mux.Handle("POST /api/documents/{id}/unpublish", adminOnly(docHandler.Unpublish))

	// Public/content read path: no auth, resolves the published record only.
	mux.HandleFunc("GET /api/public/documents/{id}", docHandler.GetPublic)

	mux.Handle("POST /api/media/upload", adminOnly(mediaHandler.Upload))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
