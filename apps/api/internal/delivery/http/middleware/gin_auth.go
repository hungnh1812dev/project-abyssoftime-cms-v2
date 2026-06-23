package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"
)

// GinAuth validates the Bearer token and injects userID + role into both
// the Gin context (ginCtx.Set) and the request context (context.WithValue).
// The dual injection keeps the GraphQL @auth directive and usecases working
// unchanged — they read from context.Context, not from gin.Context.
func GinAuth() gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		header := ginCtx.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			ginCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		claims, err := pkgjwt.ValidateToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		ginCtx.Set("userID", claims.UserID)
		ginCtx.Set("role", claims.Role)

		ctx := context.WithValue(ginCtx.Request.Context(), ContextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, ContextKeyRole, claims.Role)
		ginCtx.Request = ginCtx.Request.WithContext(ctx)

		ginCtx.Next()
	}
}

// GinRequireRole aborts with 403 if the role in context does not match.
// Deprecated: use GinRequireMinRole for hierarchy-aware checks.
func GinRequireRole(role string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		if ginCtx.GetString("role") != role {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		ginCtx.Next()
	}
}

// GinRequireMinRole aborts with 403 if the authenticated user's role level
// is below the specified minimum role.
func GinRequireMinRole(minRole string) gin.HandlerFunc {
	minLevel := entity.RoleLevel(entity.Role(minRole))
	return func(ginCtx *gin.Context) {
		actual := entity.RoleLevel(entity.Role(ginCtx.GetString("role")))
		if actual < minLevel {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		ginCtx.Next()
	}
}
