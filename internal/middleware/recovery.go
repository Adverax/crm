package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/adverax/crm/internal/pkg/apperror"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered",
					"error", r,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"request_id", c.GetString("request_id"),
				)
				apperror.Respond(c, apperror.Internal("internal server error"))
				c.Abort()
			}
		}()
		c.Next()
	}
}

func Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
