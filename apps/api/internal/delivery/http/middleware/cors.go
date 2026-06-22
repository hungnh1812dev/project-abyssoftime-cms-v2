package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[strings.TrimRight(origin, "/")] = true
	}

	return func(ginCtx *gin.Context) {
		origin := ginCtx.GetHeader("Origin")
		if origin == "" {
			ginCtx.Next()
			return
		}

		if !allowed[strings.TrimRight(origin, "/")] {
			ginCtx.Next()
			return
		}

		ginCtx.Header("Access-Control-Allow-Origin", origin)
		ginCtx.Header("Access-Control-Allow-Credentials", "true")
		ginCtx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ginCtx.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		ginCtx.Header("Access-Control-Max-Age", "86400")

		if ginCtx.Request.Method == http.MethodOptions {
			ginCtx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ginCtx.Next()
	}
}
