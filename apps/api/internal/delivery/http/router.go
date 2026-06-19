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
	GraphQLHandler http.Handler
	GraphQLPath    string
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

	// Content type routes (admin only)
	ctGroup := r.Group("/api/content-types", middleware.GinAuth(), middleware.GinRequireRole("admin"))
	{
		ctGroup.GET("", cfg.CTHandler.ListSummary)
		ctGroup.GET("/:identifier", cfg.CTHandler.Get)
	}

	// Document routes — single-type
	stGroup := r.Group("/api/document-manager/single-type", middleware.GinAuth())
	{
		stGroup.GET("/:slug", cfg.DocHandler.GetSingleType)
		stGroup.PUT("/:slug", middleware.GinRequireRole("admin"), cfg.DocHandler.SaveSingleType)
		stGroup.POST("/:slug/publish", middleware.GinRequireRole("admin"), cfg.DocHandler.PublishSingleType)
		stGroup.POST("/:slug/unpublish", middleware.GinRequireRole("admin"), cfg.DocHandler.UnpublishSingleType)
	}

	// Document routes — collection-type
	colGroup := r.Group("/api/document-manager/collection-type", middleware.GinAuth())
	{
		colGroup.GET("/:slug", cfg.DocHandler.ListCollection)
		colGroup.GET("/:slug/:documentId", cfg.DocHandler.GetCollection)
		colGroup.POST("/:slug", middleware.GinRequireRole("admin"), cfg.DocHandler.CreateCollection)
		colGroup.PUT("/:slug/:documentId", middleware.GinRequireRole("admin"), cfg.DocHandler.UpdateCollection)
		colGroup.DELETE("/:slug/:documentId", middleware.GinRequireRole("admin"), cfg.DocHandler.DeleteCollection)
		colGroup.POST("/:slug/:documentId/publish", middleware.GinRequireRole("admin"), cfg.DocHandler.PublishCollection)
		colGroup.POST("/:slug/:documentId/unpublish", middleware.GinRequireRole("admin"), cfg.DocHandler.UnpublishCollection)
	}

	// Public document route (no auth)
	r.GET("/api/public/document-manager/:slug/:documentId", cfg.DocHandler.GetPublic)

	// Media routes (admin only)
	mediaGroup := r.Group("/api/media", middleware.GinAuth(), middleware.GinRequireRole("admin"))
	{
		mediaGroup.GET("", cfg.MediaHandler.List)
		mediaGroup.POST("/upload", cfg.MediaHandler.Upload)
		mediaGroup.DELETE("/:id", cfg.MediaHandler.Delete)
	}

	// Locales (public)
	r.GET("/api/locales", cfg.LocaleHandler.List)

	// GraphQL endpoint — wrap the existing gqlgen http.Handler
	if cfg.GraphQLHandler != nil {
		r.Any(cfg.GraphQLPath, gin.WrapH(cfg.GraphQLHandler))
	}

	return r
}
