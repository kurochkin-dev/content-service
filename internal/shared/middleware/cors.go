package middleware

import (
	"net/http"
	"os"

	"content-service/internal/shared/config"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		allowOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
		if allowOrigin == "" {
			if cfg.IsProduction() {
				allowOrigin = ""
			} else {
				allowOrigin = "*"
			}
		}

		if allowOrigin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")

			if allowOrigin != "*" {
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
