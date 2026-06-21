package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowAll := len(allowedOrigins) == 0
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowAll = true
			break
		}
		originSet[origin] = struct{}{}
	}

	return func(ctx *gin.Context) {
		header := ctx.Writer.Header()
		header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		header.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		origin := ctx.GetHeader("Origin")
		if allowAll {
			header.Set("Access-Control-Allow-Origin", "*")
		} else if _, ok := originSet[origin]; ok {
			header.Set("Access-Control-Allow-Origin", origin)
		}

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
