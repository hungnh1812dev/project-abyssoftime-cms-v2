package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/delivery/http/handler"
	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
)

type RouterConfig struct {
	AuthHandler    *handler.AuthHandler
	CTHandler      *handler.ContentTypeHandler
	DocHandler     *handler.DocumentHandler
	MediaHandler   *handler.MediaHandler
	LocaleHandler  *handler.LocaleHandler
	UserHandler        *handler.UserHandler
	InviteHandler      *handler.InviteHandler
	AccessTokenHandler *handler.AccessTokenHandler
	GraphQLHandler     http.Handler
	GraphQLPath        string
}

func SetupRouter(cfg RouterConfig) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

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

	// Content type routes (admin+)
	ctGroup := r.Group("/api/content-types", middleware.GinAuth(), middleware.GinRequireMinRole("admin"))
	{
		ctGroup.GET("", cfg.CTHandler.ListSummary)
		ctGroup.GET("/:identifier", cfg.CTHandler.Get)
	}

	// Document routes — single-type
	stGroup := r.Group("/api/document-manager/single-type", middleware.GinAuth())
	{
		stGroup.GET("/:slug", cfg.DocHandler.GetSingleType)
		stGroup.PUT("/:slug", middleware.GinRequireMinRole("editor"), cfg.DocHandler.SaveSingleType)
		stGroup.POST("/:slug/publish", middleware.GinRequireMinRole("editor"), cfg.DocHandler.PublishSingleType)
		stGroup.POST("/:slug/unpublish", middleware.GinRequireMinRole("editor"), cfg.DocHandler.UnpublishSingleType)
	}

	// Document routes — collection-type
	colGroup := r.Group("/api/document-manager/collection-type", middleware.GinAuth())
	{
		colGroup.GET("/:slug", cfg.DocHandler.ListCollection)
		colGroup.GET("/:slug/:documentId", cfg.DocHandler.GetCollection)
		colGroup.POST("/:slug", middleware.GinRequireMinRole("editor"), cfg.DocHandler.CreateCollection)
		colGroup.PUT("/:slug/:documentId", middleware.GinRequireMinRole("editor"), cfg.DocHandler.UpdateCollection)
		colGroup.DELETE("/:slug/:documentId", middleware.GinRequireMinRole("editor"), cfg.DocHandler.DeleteCollection)
		colGroup.POST("/:slug/:documentId/publish", middleware.GinRequireMinRole("editor"), cfg.DocHandler.PublishCollection)
		colGroup.POST("/:slug/:documentId/unpublish", middleware.GinRequireMinRole("editor"), cfg.DocHandler.UnpublishCollection)
	}

	// Public document route (no auth)
	r.GET("/api/public/document-manager/:slug/:documentId", cfg.DocHandler.GetPublic)

	// Media routes (editor+)
	mediaGroup := r.Group("/api/media", middleware.GinAuth(), middleware.GinRequireMinRole("editor"))
	{
		mediaGroup.GET("", cfg.MediaHandler.List)
		mediaGroup.POST("/upload", cfg.MediaHandler.Upload)
		mediaGroup.DELETE("/:id", cfg.MediaHandler.Delete)
	}

	// User management routes (admin+)
	userGroup := r.Group("/api/users", middleware.GinAuth(), middleware.GinRequireMinRole("admin"))
	{
		userGroup.GET("", cfg.UserHandler.List)
		userGroup.GET("/:id", cfg.UserHandler.Get)
		userGroup.PUT("/:id/role", cfg.UserHandler.UpdateRole)
		userGroup.DELETE("/:id", cfg.UserHandler.Delete)
	}

	// Invite routes (admin+)
	inviteGroup := r.Group("/api/invites", middleware.GinAuth(), middleware.GinRequireMinRole("admin"))
	{
		inviteGroup.POST("", cfg.InviteHandler.Create)
		inviteGroup.GET("", cfg.InviteHandler.List)
		inviteGroup.DELETE("/:id", cfg.InviteHandler.Revoke)
	}

	// Public invite accept
	authGroup.POST("/invite/:token", cfg.InviteHandler.Accept)

	// Access token routes (super_admin only)
	tokenGroup := r.Group("/api/access-tokens", middleware.GinAuth(), middleware.GinRequireMinRole("super_admin"))
	{
		tokenGroup.POST("", cfg.AccessTokenHandler.Create)
		tokenGroup.GET("", cfg.AccessTokenHandler.List)
		tokenGroup.DELETE("/:id", cfg.AccessTokenHandler.Delete)
	}

	// Locales (public)
	r.GET("/api/locales", cfg.LocaleHandler.List)

	// GraphQL endpoint — wrap the existing gqlgen http.Handler
	if cfg.GraphQLHandler != nil {
		r.Any(cfg.GraphQLPath, gin.WrapH(cfg.GraphQLHandler))
	}

	return r
}
