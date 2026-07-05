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

// AuthHandler holds dependencies for auth routes.
type AuthHandler struct {
	DB             *gorm.DB
	JWTSecret      string
	GoogleClientID string
}

// Config exposes public, non-secret client configuration (e.g. the Google
// client ID) so the frontend can enable Google sign-in when it's set.
func (h *AuthHandler) Config(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"google_client_id":    h.GoogleClientID,
		"google_sign_in":      h.GoogleClientID != "",
	})
}

type signupRequest struct {
	OrgName  string `json:"org_name" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// Signup creates a new organization and its first admin user.
func (h *AuthHandler) Signup(c *gin.Context) {
	var req signupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	org := models.Organization{Name: req.OrgName, PlanTier: "free"}
	user := models.User{Email: email, Name: req.Name, Role: models.RoleAdmin, PasswordHash: string(hash)}

	// Create org and admin user atomically.
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&org).Error; err != nil {
			return err
		}
		user.OrgID = org.ID
		return tx.Create(&user).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}

	h.issueAndRespond(c, user)
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login authenticates a user and returns a JWT.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))

	var user models.User
	if err := h.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	h.issueAndRespond(c, user)
}

// Me returns the currently authenticated user.
func (h *AuthHandler) Me(c *gin.Context) {
	var user models.User
	if err := h.DB.First(&user, "id = ?", auth.UserID(c)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) issueAndRespond(c *gin.Context, user models.User) {
	token, err := auth.IssueToken(h.JWTSecret, user.ID, user.OrgID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}
