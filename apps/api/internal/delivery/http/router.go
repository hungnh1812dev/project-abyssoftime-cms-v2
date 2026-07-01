package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/handler"
	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
)

type RouterConfig struct {
	AuthHandler        *handler.AuthHandler
	CTHandler          *handler.ContentTypeHandler
	DocHandler         *handler.DocumentHandler
	MediaHandler       *handler.MediaHandler
	LocaleHandler      *handler.LocaleHandler
	UserHandler        *handler.UserHandler
	InviteHandler      *handler.InviteHandler
	AccessTokenHandler *handler.AccessTokenHandler
	RoleHandler        *handler.RoleHandler
	RoleCache          *middleware.RoleCache
	GraphQLHandler     http.Handler
	GraphQLPath        string
	CORSOrigins        []string
	EnableDebug        bool
}

func SetupRouter(cfg RouterConfig) *gin.Engine {
	if !cfg.EnableDebug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(cfg.CORSOrigins))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes (public, no middleware)
	authGroup := r.Group("/auth")
	{
		authGroup.GET("/setup", cfg.AuthHandler.SetupStatus)
		authGroup.POST("/register", cfg.AuthHandler.Register)
		authGroup.POST("/login", cfg.AuthHandler.Login)
		authGroup.POST("/refresh", cfg.AuthHandler.Refresh)
		authGroup.POST("/logout", cfg.AuthHandler.Logout)
	}

	cache := cfg.RoleCache

	// Content type routes
	ctGroup := r.Group("/api/content-types", middleware.GinAuth(), middleware.GinRequirePermission(cache, "content_types:read"))
	{
		ctGroup.GET("", cfg.CTHandler.ListSummary)
		ctGroup.GET("/:identifier", cfg.CTHandler.Get)
		ctGroup.PATCH("/:slug/list-fields", cfg.CTHandler.UpdateListFields)
	}

	// Document routes — single-type
	stGroup := r.Group("/api/document-manager/single-type", middleware.GinAuth())
	{
		stGroup.GET("/:slug", middleware.GinRequirePermission(cache, "content:read"), cfg.DocHandler.GetSingleType)
		stGroup.PUT("/:slug", middleware.GinRequirePermission(cache, "content:update"), cfg.DocHandler.SaveSingleType)
		stGroup.POST("/:slug/publish", middleware.GinRequirePermission(cache, "content:publish"), cfg.DocHandler.PublishSingleType)
		stGroup.POST("/:slug/unpublish", middleware.GinRequirePermission(cache, "content:unpublish"), cfg.DocHandler.UnpublishSingleType)
	}

	// Document routes — collection-type
	colGroup := r.Group("/api/document-manager/collection-type", middleware.GinAuth())
	{
		colGroup.GET("/:slug", middleware.GinRequirePermission(cache, "content:read"), cfg.DocHandler.ListCollection)
		colGroup.GET("/:slug/:documentId", middleware.GinRequirePermission(cache, "content:read"), cfg.DocHandler.GetCollection)
		colGroup.POST("/:slug", middleware.GinRequirePermission(cache, "content:create"), cfg.DocHandler.CreateCollection)
		colGroup.POST("/:slug/bulk", middleware.GinRequirePermission(cache, "content:create"), middleware.GinRequirePermission(cache, "content:publish"), cfg.DocHandler.BulkCreateCollection)
		colGroup.PUT("/:slug/:documentId", middleware.GinRequirePermission(cache, "content:update"), cfg.DocHandler.UpdateCollection)
		colGroup.DELETE("/:slug/:documentId", middleware.GinRequirePermission(cache, "content:delete"), cfg.DocHandler.DeleteCollection)
		colGroup.POST("/:slug/:documentId/publish", middleware.GinRequirePermission(cache, "content:publish"), cfg.DocHandler.PublishCollection)
		colGroup.POST("/:slug/:documentId/unpublish", middleware.GinRequirePermission(cache, "content:unpublish"), cfg.DocHandler.UnpublishCollection)
		colGroup.POST("/:slug/:documentId/duplicate", middleware.GinRequirePermission(cache, "content:create"), cfg.DocHandler.DuplicateCollection)
	}

	// Public document route (no auth)
	r.GET("/api/public/document-manager/:slug/:documentId", cfg.DocHandler.GetPublic)

	// Media routes
	mediaGroup := r.Group("/api/media", middleware.GinAuth())
	{
		mediaGroup.GET("", middleware.GinRequirePermission(cache, "media:read"), cfg.MediaHandler.List)
		mediaGroup.POST("/upload", middleware.GinRequirePermission(cache, "media:upload"), cfg.MediaHandler.Upload)
		mediaGroup.DELETE("/:id", middleware.GinRequirePermission(cache, "media:delete"), cfg.MediaHandler.Delete)
	}

	// User management routes
	userGroup := r.Group("/api/users", middleware.GinAuth(), middleware.GinRequirePermission(cache, "users:manage"))
	{
		userGroup.GET("", cfg.UserHandler.List)
		userGroup.GET("/:id", cfg.UserHandler.Get)
		userGroup.PUT("/:id/role", cfg.UserHandler.UpdateRole)
		userGroup.DELETE("/:id", cfg.UserHandler.Delete)
	}

	// Invite routes
	inviteGroup := r.Group("/api/invites", middleware.GinAuth(), middleware.GinRequirePermission(cache, "users:manage"))
	{
		inviteGroup.POST("", cfg.InviteHandler.Create)
		inviteGroup.GET("", cfg.InviteHandler.List)
		inviteGroup.DELETE("/:id", cfg.InviteHandler.Revoke)
	}

	// Public invite accept
	authGroup.POST("/invite/:token", cfg.InviteHandler.Accept)

	// Role routes
	roleGroup := r.Group("/api/roles", middleware.GinAuth())
	{
		roleGroup.GET("", cfg.RoleHandler.List)
		roleGroup.GET("/:id", cfg.RoleHandler.Get)
		roleGroup.POST("", middleware.GinRequirePermission(cache, "roles:manage"), cfg.RoleHandler.Create)
		roleGroup.PUT("/:id", middleware.GinRequirePermission(cache, "roles:manage"), cfg.RoleHandler.Update)
		roleGroup.DELETE("/:id", middleware.GinRequirePermission(cache, "roles:manage"), cfg.RoleHandler.Delete)
	}

	// Access token routes
	tokenGroup := r.Group("/api/access-tokens", middleware.GinAuth(), middleware.GinRequirePermission(cache, "access_tokens:manage"))
	{
		tokenGroup.POST("", cfg.AccessTokenHandler.Create)
		tokenGroup.GET("", cfg.AccessTokenHandler.List)
		tokenGroup.DELETE("/:id", cfg.AccessTokenHandler.Delete)
	}

	// Locale routes
	localeGroup := r.Group("/api/locales")
	{
		localeGroup.GET("", cfg.LocaleHandler.List)
		localeGroup.POST("", middleware.GinAuth(), middleware.GinRequirePermission(cache, "locales:manage"), cfg.LocaleHandler.Create)
		localeGroup.PUT("/:code", middleware.GinAuth(), middleware.GinRequirePermission(cache, "locales:manage"), cfg.LocaleHandler.Update)
		localeGroup.DELETE("/:code", middleware.GinAuth(), middleware.GinRequirePermission(cache, "locales:manage"), cfg.LocaleHandler.Delete)
	}

	// GraphQL endpoint — wrap the existing gqlgen http.Handler
	if cfg.GraphQLHandler != nil {
		r.Any(cfg.GraphQLPath, gin.WrapH(cfg.GraphQLHandler))
	}

	return r
}
