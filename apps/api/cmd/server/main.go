// ci: test change detection
package main

import (
	"context"
	"log"
	"net"

	graphqlpkg "project-abyssoftime-cms-v2/api/graphql"
	"project-abyssoftime-cms-v2/api/graphql/resolver"
	"project-abyssoftime-cms-v2/api/internal/config"
	grpcdelivery "project-abyssoftime-cms-v2/api/internal/delivery/grpc"
	deliveryhttp "project-abyssoftime-cms-v2/api/internal/delivery/http"
	deliveryhandler "project-abyssoftime-cms-v2/api/internal/delivery/http/handler"
	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
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
	accesstokenuc "project-abyssoftime-cms-v2/api/internal/usecase/access_token"
	inviteuc "project-abyssoftime-cms-v2/api/internal/usecase/invite"
	localeuc "project-abyssoftime-cms-v2/api/internal/usecase/locale"
	mediauc "project-abyssoftime-cms-v2/api/internal/usecase/media"
	roleuc "project-abyssoftime-cms-v2/api/internal/usecase/role"
	useruc "project-abyssoftime-cms-v2/api/internal/usecase/user"
	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"
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
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	pkgjwt.SetSecret(cfg.JWTSecret)

	// --- database connections ---
	isPostgres := func(entity string) bool { return entity == "postgres" }

	needsMongo := cfg.DB.EntityDB.User == "mongo" ||
		cfg.DB.EntityDB.ContentType == "mongo" ||
		cfg.DB.EntityDB.Document == "mongo" ||
		cfg.DB.EntityDB.Media == "mongo"
	needsPostgres := isPostgres(cfg.DB.EntityDB.User) ||
		isPostgres(cfg.DB.EntityDB.ContentType) ||
		isPostgres(cfg.DB.EntityDB.Document) ||
		isPostgres(cfg.DB.EntityDB.Media)

	var mongoDB *mongo.Database
	if needsMongo {
		mongoClient, err := mongodb.NewClient(ctx, cfg.DB.MongoURI())
		if err != nil {
			log.Fatalf("mongodb connect: %v", err)
		}
		defer func() {
			if err := mongoClient.Disconnect(ctx); err != nil {
				log.Printf("mongodb disconnect: %v", err)
			}
		}()
		log.Println("connected to mongodb")

		mongoDB = mongodb.Database(mongoClient, cfg.DB.Name)

		if err := mongodb.EnsureIndexes(ctx, mongoDB); err != nil {
			log.Fatalf("ensure indexes: %v", err)
		}
		log.Println("indexes ensured")
	}

	var sqlDB *gorm.DB
	if needsPostgres {
		var err error
		sqlDB, err = gormdb.NewClient("postgres", cfg.DB.PostgresDSN())
		if err != nil {
			log.Fatalf("postgres connect: %v", err)
		}
		if err := gormdb.AutoMigrate(sqlDB); err != nil {
			log.Fatalf("postgres auto-migrate: %v", err)
		}
		log.Printf("connected to postgres (%s:%s/%s)", cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)
	}

	// --- repository factory ---
	var userRepo repository.UserRepository
	if isPostgres(cfg.DB.EntityDB.User) {
		userRepo = gormdb.NewUserRepository(sqlDB)
	} else {
		userRepo = mongodb.NewUserRepository(mongoDB)
	}

	var ctRepo repository.ContentTypeRepository
	if isPostgres(cfg.DB.EntityDB.ContentType) {
		ctRepo = gormdb.NewContentTypeRepository(sqlDB)
	} else {
		ctRepo = mongodb.NewContentTypeRepository(mongoDB)
	}

	var docRepo repository.DocumentRepository
	if isPostgres(cfg.DB.EntityDB.Document) {
		docRepo = gormdb.NewDocumentRepository(sqlDB)
	} else {
		docRepo = mongodb.NewDocumentRepository(mongoDB)
	}

	var mediaRepo repository.MediaAssetRepository
	if isPostgres(cfg.DB.EntityDB.Media) {
		mediaRepo = gormdb.NewMediaAssetRepository(sqlDB)
	} else {
		mediaRepo = mongodb.NewMediaAssetRepository(mongoDB)
	}

	storage, err := newStorageAdapter(ctx, cfg)
	if err != nil {
		log.Fatalf("%s init: %v", cfg.Media.Driver, err)
	}
	log.Printf("storage provider: %s", cfg.Media.Driver)

	// --- role repository ---
	var roleRepo repository.RoleRepository
	if isPostgres(cfg.DB.EntityDB.User) {
		roleRepo = gormdb.NewRoleRepository(sqlDB)
	} else {
		roleRepo = mongodb.NewRoleRepository(mongoDB)
	}

	// --- invite + access token repositories ---
	var inviteRepo repository.InviteRepository
	if isPostgres(cfg.DB.Driver) {
		inviteRepo = gormdb.NewInviteRepository(sqlDB)
	} else {
		inviteRepo = mongodb.NewInviteRepository(mongoDB)
	}

	var accessTokenRepo repository.AccessTokenRepository
	if isPostgres(cfg.DB.Driver) {
		accessTokenRepo = gormdb.NewAccessTokenRepository(sqlDB)
	} else {
		accessTokenRepo = mongodb.NewAccessTokenRepository(mongoDB)
	}

	// --- locale repository ---
	var localeRepo repository.LocaleRepository
	if isPostgres(cfg.DB.Driver) {
		localeRepo = gormdb.NewLocaleRepository(sqlDB)
	} else {
		localeRepo = mongodb.NewLocaleRepository(mongoDB)
	}

	// --- component repository (PostgreSQL only) ---
	var compRepo repository.ComponentRepository
	if isPostgres(cfg.DB.EntityDB.Document) {
		compRepo = gormdb.NewComponentRepository(sqlDB)
	}

	authUC := auth.New(userRepo, roleRepo)
	ctUC := contenttype.New(ctRepo)
	documentUC := docuc.New(docRepo, compRepo, mediaRepo, cfg.SupportedLocales)
	mediaUC := mediauc.New(mediaRepo, storage, cfg.Media.GenerateThumbnail)
	userUC := useruc.New(userRepo, roleRepo)
	inviteUC := inviteuc.New(inviteRepo, userRepo, roleRepo)
	accessTokenUC := accesstokenuc.New(accessTokenRepo)
	roleUC := roleuc.New(roleRepo, userRepo)
	localeUC := localeuc.New(localeRepo, docRepo, ctRepo)

	// Seed default roles on first startup
	if err := roleUC.SeedDefaults(ctx); err != nil {
		log.Fatalf("seed default roles: %v", err)
	}

	// Seed locales from env var on first startup
	if err := localeUC.Seed(ctx, cfg.SupportedLocales); err != nil {
		log.Fatalf("seed locales: %v", err)
	}

	// Role permission cache
	roleCache := middleware.NewRoleCache()
	if allRoles, err := roleUC.FindAll(ctx); err == nil {
		roleCache.Load(allRoles)
	}

	defsDir := cfg.ContentTypeDir
	defs, err := contenttype.LoadDefinitions(defsDir)
	if err != nil {
		log.Fatalf("load content-type definitions: %v", err)
	}
	if err := contenttype.NewSyncer(ctUC, documentUC, docRepo, compRepo).Sync(ctx, defs); err != nil {
		log.Fatalf("sync content types: %v", err)
	}
	log.Printf("synced %d content-type definitions from %s", len(defs), defsDir)

	// gqlgen GraphQL — schema pre-compiled at build time via make graphql-generate
	gqlResolver := resolver.NewResolver(documentUC, ctUC, mediaRepo)
	gqlHandler := graphqlpkg.NewHandler(gqlResolver, accessTokenUC)

	// Gin router (REST + GraphQL)
	router := deliveryhttp.SetupRouter(deliveryhttp.RouterConfig{
		AuthHandler:        deliveryhandler.NewAuthHandler(authUC, cfg.CookieSecure, cfg.CookieSameSite),
		CTHandler:          deliveryhandler.NewContentTypeHandler(ctUC),
		DocHandler:         deliveryhandler.NewDocumentHandler(documentUC, ctUC, userRepo),
		MediaHandler:       deliveryhandler.NewMediaHandler(mediaUC),
		LocaleHandler:      deliveryhandler.NewLocaleHandler(localeUC),
		UserHandler:        deliveryhandler.NewUserHandler(userUC),
		InviteHandler:      deliveryhandler.NewInviteHandler(inviteUC),
		AccessTokenHandler: deliveryhandler.NewAccessTokenHandler(accessTokenUC),
		RoleHandler:        deliveryhandler.NewRoleHandler(roleUC, roleCache),
		RoleCache:          roleCache,
		GraphQLHandler:     gqlHandler,
		GraphQLPath:        cfg.GraphQL.Path,
		CORSOrigins:        cfg.CORSOrigins,
	})

	// gRPC server (opt-in via GRPC_ENABLED=true)
	if cfg.GRPCEnabled {
		grpcSrv := grpcdelivery.NewServer(authUC, ctUC, documentUC, mediaUC)
		go func() {
			grpcAddr := ":" + cfg.GRPCPort
			lis, err := net.Listen("tcp", grpcAddr)
			if err != nil {
				log.Printf("gRPC disabled: %v", err)
				return
			}
			log.Printf("gRPC server listening on %s", grpcAddr)
			if err := grpcSrv.Serve(lis); err != nil {
				log.Printf("gRPC stopped: %v", err)
			}
		}()
	}

	// Start REST + GraphQL (blocks)
	addr := ":" + cfg.Port
	log.Printf("REST server listening on %s", addr)
	log.Printf("graphql endpoint: %s", cfg.GraphQL.Path)
	if err := router.Run(addr); err != nil {
		log.Fatal(err)
	}
}
