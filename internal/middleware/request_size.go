package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MaxBody limits the size of incoming request bodies.
func MaxBody(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}
		c.Next()
	}
}
