package main

import (
	"context"
	"log"
	"net/http"

	"project-abyssoftime-cms-v2/api/graphql/codegen"
	"project-abyssoftime-cms-v2/api/graphql/generated"
	"project-abyssoftime-cms-v2/api/graphql/resolver"
	"project-abyssoftime-cms-v2/api/internal/config"
	deliveryhttp "project-abyssoftime-cms-v2/api/internal/delivery/http"
	deliveryhandler "project-abyssoftime-cms-v2/api/internal/delivery/http/handler"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	cloudinaryadapter "project-abyssoftime-cms-v2/api/internal/infrastructure/cloudinary"
	"project-abyssoftime-cms-v2/api/internal/infrastructure/gormdb"
	"project-abyssoftime-cms-v2/api/internal/infrastructure/mongodb"
	s3adapter "project-abyssoftime-cms-v2/api/internal/infrastructure/s3"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
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

	// --- database connections ---
	needsMongo := cfg.DB.EntityDB.User == "mongo" ||
		cfg.DB.EntityDB.ContentType == "mongo" ||
		cfg.DB.EntityDB.Document == "mongo" ||
		cfg.DB.EntityDB.Media == "mongo"
	needsSQL := cfg.DB.EntityDB.User == "sql" ||
		cfg.DB.EntityDB.ContentType == "sql" ||
		cfg.DB.EntityDB.Document == "sql" ||
		cfg.DB.EntityDB.Media == "sql"

	var mongoDB *mongo.Database
	if needsMongo {
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

		mongoDB = mongodb.Database(mongoClient, cfg.DB.Mongo.Name)

		if err := mongodb.EnsureIndexes(ctx, mongoDB); err != nil {
			log.Fatalf("ensure indexes: %v", err)
		}
		log.Println("indexes ensured")
	}

	var sqlDB *gorm.DB
	if needsSQL {
		var err error
		sqlDB, err = gormdb.NewClient(cfg.DB.SQL.Driver, cfg.DB.SQL.DSN)
		if err != nil {
			log.Fatalf("gorm connect: %v", err)
		}
		if err := gormdb.AutoMigrate(sqlDB); err != nil {
			log.Fatalf("gorm auto-migrate: %v", err)
		}
		log.Printf("connected to sql (%s)", cfg.DB.SQL.Driver)
	}

	// --- repository factory ---
	var userRepo repository.UserRepository
	if cfg.DB.EntityDB.User == "sql" {
		userRepo = gormdb.NewUserRepository(sqlDB)
	} else {
		userRepo = mongodb.NewUserRepository(mongoDB)
	}

	var ctRepo repository.ContentTypeRepository
	if cfg.DB.EntityDB.ContentType == "sql" {
		ctRepo = gormdb.NewContentTypeRepository(sqlDB)
	} else {
		ctRepo = mongodb.NewContentTypeRepository(mongoDB)
	}

	var docRepo repository.DocumentRepository
	if cfg.DB.EntityDB.Document == "sql" {
		docRepo = gormdb.NewDocumentRepository(sqlDB)
	} else {
		docRepo = mongodb.NewDocumentRepository(mongoDB)
	}

	var mediaRepo repository.MediaAssetRepository
	if cfg.DB.EntityDB.Media == "sql" {
		mediaRepo = gormdb.NewMediaAssetRepository(sqlDB)
	} else {
		mediaRepo = mongodb.NewMediaAssetRepository(mongoDB)
	}

	storage, err := newStorageAdapter(ctx, cfg)
	if err != nil {
		log.Fatalf("%s init: %v", cfg.Media.Driver, err)
	}
	log.Printf("storage provider: %s", cfg.Media.Driver)

	authUC := auth.New(userRepo)
	ctUC := contenttype.New(ctRepo)
	documentUC := docuc.New(docRepo, mediaRepo, cfg.SupportedLocales)
	mediaUC := mediauc.New(mediaRepo, storage, cfg.Media.GenerateThumbnail)

	defsDir := cfg.ContentTypeDir
	defs, err := contenttype.LoadDefinitions(defsDir)
	if err != nil {
		log.Fatalf("load content-type definitions: %v", err)
	}
	if err := contenttype.NewSyncer(ctUC, documentUC, docRepo).Sync(ctx, defs); err != nil {
		log.Fatalf("sync content types: %v", err)
	}
	log.Printf("synced %d content-type definitions from %s", len(defs), defsDir)

	// GraphQL
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
	gqlHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := resolver.WithRequest(r.Context(), r)
		gqlSrv.ServeHTTP(w, r.WithContext(ctx))
	})

	// Gin router
	router := deliveryhttp.SetupRouter(deliveryhttp.RouterConfig{
		AuthHandler:    deliveryhandler.NewAuthHandler(authUC),
		CTHandler:      deliveryhandler.NewContentTypeHandler(ctUC),
		DocHandler:     deliveryhandler.NewDocumentHandler(documentUC, ctUC),
		MediaHandler:   deliveryhandler.NewMediaHandler(mediaUC),
		LocaleHandler:  deliveryhandler.NewLocaleHandler(cfg.SupportedLocales),
		GraphQLHandler: gqlHandler,
		GraphQLPath:    cfg.GraphQL.Path,
	})

	addr := ":" + cfg.Port
	log.Printf("server listening on %s", addr)
	log.Printf("graphql endpoint: %s", cfg.GraphQL.Path)
	if err := router.Run(addr); err != nil {
		log.Fatal(err)
	}
}
