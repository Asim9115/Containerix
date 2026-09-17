package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/asim9115/containerix/internal/auth"
	"github.com/asim9115/containerix/internal/repository"
	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"
const UserKey = "user"

func extractAPIKey(c *gin.Context) string {
	rawKey := c.GetHeader("X-API-Key")
	if rawKey != "" {
		return rawKey
	}
	bearer := c.GetHeader("Authorization")
	if strings.HasPrefix(bearer, "Bearer ") {
		return strings.TrimPrefix(bearer, "Bearer ")
	}
	if strings.HasPrefix(bearer, "Bearer") {
		return strings.TrimSpace(strings.TrimPrefix(bearer, "Bearer"))
	}
	return ""
}

func secureEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// takes api key from request, validates the key with db and sets the user id and user
func APIKeyAuth(repos *repository.Repos) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawKey := extractAPIKey(c)

		if rawKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing API key — provide X-API-Key header",
			})
			return
		}

		hash := auth.HashApiKey(rawKey)
		user, err := repos.User.GetByApiKeyHash(hash)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "auth lookup failed",
			})
			return
		}
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
			})
			return
		}

		c.Set(UserIDKey, user.ID)
		c.Set(UserKey, user)
		c.Next()
	}
}

// AdminAPIKeyAuth requires X-API-Key (or Bearer) to match CONTAINERIX_ADMIN_API_KEY.
func AdminAPIKeyAuth(adminKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if adminKey == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "admin API key not configured",
			})
			return
		}

		rawKey := extractAPIKey(c)
		if rawKey == "" || !secureEqual(rawKey, adminKey) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or missing admin API key",
			})
			return
		}
		c.Next()
	}
}
