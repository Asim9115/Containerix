package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)


type ipWindow struct {
	count   int
	resetAt time.Time
}

type windowLimiter struct {
	mu      sync.Mutex
	entries map[string]*ipWindow
	limit   int
	window  time.Duration
}

func newWindowLimiter(limit int, window time.Duration) *windowLimiter {
	if limit < 1 {
		limit = 1
	}
	if window < time.Second {
		window = time.Minute
	}
	return &windowLimiter{
		entries: make(map[string]*ipWindow),
		limit:   limit,
		window:  window,
	}
}

func (l *windowLimiter) allow(key string) (ok bool, retryAfter time.Duration) {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	for k, entry := range l.entries {
		if now.After(entry.resetAt) {
			delete(l.entries, k)
		}
	}

	entry, exists := l.entries[key]
	if !exists || now.After(entry.resetAt) {
		l.entries[key] = &ipWindow{count: 1, resetAt: now.Add(l.window)}
		return true, 0
	}

	if entry.count >= l.limit {
		return false, entry.resetAt.Sub(now)
	}

	entry.count++
	return true, 0
}

func abortRateLimit(c *gin.Context, message string, retryAfter time.Duration) {
	seconds := int(retryAfter.Seconds()) + 1
	if seconds < 1 {
		seconds = 1
	}
	c.Header("Retry-After", strconv.Itoa(seconds))
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": message})
}

// GlobalRateLimit applies a per-IP limit to all routes. Set limit <= 0 to disable.
// /health and /ready are excluded so orchestrator probes are not throttled.
func GlobalRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	if limit <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	limiter := newWindowLimiter(limit, window)

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/health" || path == "/ready" {
			c.Next()
			return
		}

		ok, retryAfter := limiter.allow(c.ClientIP())
		if ok {
			c.Next()
			return
		}

		abortRateLimit(c, "rate limit exceeded", retryAfter)
	}
}

// RegistrationRateLimit is a stricter per-IP limit for POST /users only.
func RegistrationRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	if limit <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	limiter := newWindowLimiter(limit, window)

	return func(c *gin.Context) {
		ok, retryAfter := limiter.allow(c.ClientIP())
		if ok {
			c.Next()
			return
		}

		abortRateLimit(c, "rate limit exceeded — too many registration attempts", retryAfter)
	}
}
