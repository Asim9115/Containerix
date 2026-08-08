package middleware

import (
	"net/http"
	"strings"

	"github.com/asim9115/containerix/internal/auth"
	"github.com/asim9115/containerix/internal/repository"
	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"
const UserKey = "user"


// takes api key from request, validates the key with db and sets the user id and user 
func APIKeyAuth(repos *repository.Repos) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawKey := c.GetHeader("X-API-Key")
		if rawKey == "" {
			bearer := c.GetHeader("Authorization")
			if strings.HasPrefix(bearer, "Bearer") {
				rawKey = strings.TrimPrefix(bearer, "Bearer ")
			}
		}

		if rawKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":"missing API key — provide X-API-Key header",
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
        c.Next() //executes the pending hanlders like (from auth to deploy)
	}
}