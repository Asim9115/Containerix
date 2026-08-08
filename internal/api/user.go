package api

import (
	"net/http"
	"time"

	"github.com/asim9115/containerix/internal/auth"
	"github.com/asim9115/containerix/internal/middleware"
	"github.com/asim9115/containerix/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateUserRequest struct {
	Name string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

func (h *GlobalState) CreateUser(c *gin.Context) {
    var body CreateUserRequest
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

	rawKey, hash, err := auth.GenerateAPIKey()
	if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate API key"})
        return
	} 

	user := &repository.User{
		ID: uuid.New().String(),
		Name : body.Name,
		Email: body.Email,
		ApiKeyHash: hash,
		CreatedAt: time.Now(),
	}

	if err := h.Repos.User.Create(user); err != nil {
        c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
        return
	}

	//raw key returned
	c.JSON(http.StatusCreated, gin.H{
		"user_id":user.ID,
		"email":user.Email,
		"api_key":rawKey,
		"warning":"save this API key now..",
	})
}

// POST /users/api-key — rotate API key for the authenticated user
func (h *GlobalState) RotateAPIKey(c *gin.Context) {
    userID := c.GetString(middleware.UserIDKey)
    rawKey, hash, err := auth.GenerateAPIKey()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "key generation failed"})
        return
    }
    if err := h.Repos.User.UpdateApiKeyHash(userID, hash); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rotate key"})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "api_key": rawKey,
        "warning": "Your old key is now invalid. Save this new key immediately.",
    })
}

func (h *GlobalState) GetMe(c *gin.Context) {
	userId := c.GetString(middleware.UserIDKey)
	user, err := h.Repos.User.GetUser(userId)

	if err != nil || user == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "id":         user.ID,
        "name":       user.Name,
        "email":      user.Email,
        "created_at": user.CreatedAt,
    })
}