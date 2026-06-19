package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	pkgjwt "project-abyssoftime-cms-v2/api/pkg/jwt"
)

// GinAuth validates the Bearer token and injects userID + role into both
// the Gin context (c.Set) and the request context (context.WithValue).
// The dual injection keeps the GraphQL @auth directive and usecases working
// unchanged — they read from context.Context, not from gin.Context.
func GinAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		claims, err := pkgjwt.ValidateToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)

		ctx := context.WithValue(c.Request.Context(), ContextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, ContextKeyRole, claims.Role)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GinRequireRole aborts with 403 if the role in context does not match.
func GinRequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}
