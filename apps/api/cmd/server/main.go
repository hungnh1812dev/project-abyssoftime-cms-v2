package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"project-abyssoftime-cms-v2/api/graphql/codegen"
	"project-abyssoftime-cms-v2/api/graphql/generated"
	"project-abyssoftime-cms-v2/api/graphql/resolver"
	"project-abyssoftime-cms-v2/api/internal/config"
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
	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"

	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
)

func newStorageAdapter(ctx context.Context, cfg *config.Config) (repository.StorageAdapter, error) {
	switch cfg.Media.Driver {
	case "s3":
		return s3adapter.New(ctx, cfg.Media.S3.Bucket, cfg.Media.S3.Region)
	default:
		return cloudinaryadapter.NewCloudinaryAdapter(
			cfg.Media.Cloudinary.CloudName,
			cfg.Media.Cloudinary.APIKey,
			cfg.Media.Cloudinary.APISecret,
		)
	}
}

func main() {
	const schemaPath = "graphql/schema.graphqls"
	const sentinelPath = "graphql/.graphql.hash"
	if err := codegen.EnsureUpToDate(schemaPath, sentinelPath, codegen.DefaultRunner()); err != nil {
		log.Fatalf("graphql codegen: %v", err)
	}
	log.Println("graphql: generated code up to date")

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	pkgjwt.SetSecret(cfg.JWTSecret)

	mongoClient, err := mongodb.NewClient(ctx, cfg.DB.Mongo.URI)
	if err != nil {
		log.Fatalf("mongodb connect: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("mongodb disconnect: %v", err)
		}
	}()
	log.Println("connected to mongodb")

	db := mongodb.Database(mongoClient, cfg.DB.Mongo.Name)

	if err := mongodb.EnsureIndexes(ctx, db); err != nil {
		log.Fatalf("ensure indexes: %v", err)
	}
	log.Println("indexes ensured")

	// repositories
	userRepo := mongodb.NewUserRepository(db)
	ctRepo := mongodb.NewContentTypeRepository(db)
	docRepo := mongodb.NewDocumentRepository(db)
	mediaRepo := mongodb.NewMediaAssetRepository(db)

	// storage adapter: config-selected, S3 or Cloudinary, behind the same interface
	storage, err := newStorageAdapter(ctx, cfg)
	if err != nil {
		log.Fatalf("%s init: %v", cfg.Media.Driver, err)
	}
	log.Printf("storage provider: %s", cfg.Media.Driver)

	// usecases
	authUC := auth.New(userRepo)
	ctUC := contenttype.New(ctRepo)
	documentUC := docuc.New(docRepo, mediaRepo, cfg.SupportedLocales)
	mediaUC := mediauc.New(mediaRepo, storage, cfg.Media.GenerateThumbnail)

	// content-type schema-as-code sync: JSON definitions are the source of
	// truth, reconciled into Mongo before the server starts accepting traffic.
	defsDir := cfg.ContentTypeDir
	defs, err := contenttype.LoadDefinitions(defsDir)
	if err != nil {
		log.Fatalf("load content-type definitions: %v", err)
	}
	if err := contenttype.NewSyncer(ctUC, documentUC, docRepo).Sync(ctx, defs); err != nil {
		log.Fatalf("sync content types: %v", err)
	}
	log.Printf("synced %d content-type definitions from %s", len(defs), defsDir)

	// handlers
	authHandler := deliveryhandler.NewAuthHandler(authUC)
	ctHandler := deliveryhandler.NewContentTypeHandler(ctUC)
	docHandler := deliveryhandler.NewDocumentHandler(documentUC, ctUC)
	mediaHandler := deliveryhandler.NewMediaHandler(mediaUC)
	localeHandler := deliveryhandler.NewLocaleHandler(cfg.SupportedLocales)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	// auth routes
	mux.HandleFunc("GET /auth/setup", authHandler.SetupStatus)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)

	// content type routes
	adminOnly := func(h http.HandlerFunc) http.Handler {
		return middleware.Auth(middleware.RequireRole("admin", h))
	}
	mux.Handle("GET /api/content-types", adminOnly(ctHandler.ListSummary))
	mux.Handle("GET /api/content-types/{identifier}", adminOnly(ctHandler.Get))

	// document routes — single-type
	authRequired := func(h http.HandlerFunc) http.Handler {
		return middleware.Auth(http.HandlerFunc(h))
	}
	mux.Handle("GET /api/document-manager/single-type/{slug}", authRequired(docHandler.GetSingleType))
	mux.Handle("PUT /api/document-manager/single-type/{slug}", adminOnly(docHandler.SaveSingleType))
	mux.Handle("POST /api/document-manager/single-type/{slug}/publish", adminOnly(docHandler.PublishSingleType))
	mux.Handle("POST /api/document-manager/single-type/{slug}/unpublish", adminOnly(docHandler.UnpublishSingleType))

	// document routes — collection-type
	mux.Handle("GET /api/document-manager/collection-type/{slug}", authRequired(docHandler.ListCollection))
	mux.Handle("GET /api/document-manager/collection-type/{slug}/{documentId}", authRequired(docHandler.GetCollection))
	mux.Handle("POST /api/document-manager/collection-type/{slug}", adminOnly(docHandler.CreateCollection))
	mux.Handle("PUT /api/document-manager/collection-type/{slug}/{documentId}", adminOnly(docHandler.UpdateCollection))
	mux.Handle("DELETE /api/document-manager/collection-type/{slug}/{documentId}", adminOnly(docHandler.DeleteCollection))
	mux.Handle("POST /api/document-manager/collection-type/{slug}/{documentId}/publish", adminOnly(docHandler.PublishCollection))
	mux.Handle("POST /api/document-manager/collection-type/{slug}/{documentId}/unpublish", adminOnly(docHandler.UnpublishCollection))

	// public document route, no auth, only returns published documents
	mux.HandleFunc("GET /api/public/document-manager/{slug}/{documentId}", docHandler.GetPublic)

	// media routes
	mux.Handle("GET /api/media", adminOnly(mediaHandler.List))
	mux.Handle("POST /api/media/upload", adminOnly(mediaHandler.Upload))
	mux.Handle("DELETE /api/media/{id}", adminOnly(mediaHandler.Delete))

	mux.HandleFunc("GET /api/locales", localeHandler.List)

	// GraphQL endpoint alongside REST — same usecases, auth via @auth directive.
	gqlSchema := generated.NewExecutableSchema(generated.Config{
		Resolvers: &resolver.Resolver{
			DocumentUC:    documentUC,
			ContentTypeUC: ctUC,
		},
		Directives: generated.DirectiveRoot{
			Auth: resolver.AuthDirective,
		},
	})
	gqlSrv := gqlhandler.NewDefaultServer(gqlSchema)
	mux.Handle(cfg.GraphQL.Path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := resolver.WithRequest(r.Context(), r)
		gqlSrv.ServeHTTP(w, r.WithContext(ctx))
	}))
	log.Printf("graphql endpoint: %s", cfg.GraphQL.Path)

	addr := ":" + cfg.Port
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
