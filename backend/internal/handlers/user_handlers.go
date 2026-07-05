package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"quota/internal/auth"
	"quota/internal/models"
)

// UserHandler manages team members within an org.
type UserHandler struct {
	DB *gorm.DB
}

// List returns all users in the caller's org (managers/admins only).
func (h *UserHandler) List(c *gin.Context) {
	var users []models.User
	if err := h.DB.Where("org_id = ?", auth.OrgID(c)).Order("name").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

type createUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required"`
}

// Create adds a new team member (rep/manager) to the caller's org.
func (h *UserHandler) Create(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role != models.RoleRep && req.Role != models.RoleManager && req.Role != models.RoleAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))

	var existing models.User
	if err := h.DB.Where("email = ?", email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := models.User{
		OrgID:        auth.OrgID(c),
		Name:         req.Name,
		Email:        email,
		Role:         req.Role,
		PasswordHash: string(hash),
	}
	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, user)
}
