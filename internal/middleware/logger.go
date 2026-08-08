package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        log.Printf("[%d] %s %s | user=%s | latency=%s",
            c.Writer.Status(),
            c.Request.Method,
            c.Request.URL.Path,
            c.GetString(UserIDKey),
            time.Since(start),
        )
    }
}