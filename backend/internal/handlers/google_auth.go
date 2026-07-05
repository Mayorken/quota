package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"quota/internal/models"
)

// googleTokenInfo is the subset of Google's tokeninfo response we use.
type googleTokenInfo struct {
	Aud           string `json:"aud"`
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"` // returned as "true"/"false" strings
	Name          string `json:"name"`
	Exp           string `json:"exp"`
	Error         string `json:"error_description"`
}

// verifyGoogleIDToken validates a Google ID token and returns its claims.
// It uses Google's tokeninfo endpoint so no signing-key dependency is needed.
func verifyGoogleIDToken(credential, clientID string) (*googleTokenInfo, error) {
	endpoint := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(credential)

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("contacting Google: %w", err)
	}
	defer resp.Body.Close()

	var info googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decoding Google response: %w", err)
	}
	if resp.StatusCode != http.StatusOK || info.Error != "" {
		return nil, errors.New("invalid Google token")
	}
	// The token's audience must match our configured client ID.
	if clientID != "" && info.Aud != clientID {
		return nil, errors.New("token audience mismatch")
	}
	if info.EmailVerified != "true" {
		return nil, errors.New("Google email is not verified")
	}
	if info.Email == "" {
		return nil, errors.New("no email in Google token")
	}
	return &info, nil
}

type googleAuthRequest struct {
	Credential string `json:"credential" binding:"required"`
	OrgName    string `json:"org_name"` // used only when creating a brand-new org
}

// GoogleAuth signs a user in (or provisions them on first sign-in) via a
// Google ID token obtained on the client by Google Identity Services.
func (h *AuthHandler) GoogleAuth(c *gin.Context) {
	if h.GoogleClientID == "" {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Google sign-in is not configured"})
		return
	}
	var req googleAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	info, err := verifyGoogleIDToken(req.Credential, h.GoogleClientID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	email := strings.ToLower(strings.TrimSpace(info.Email))

	// Existing user → sign in, linking the Google ID if not already set.
	var user models.User
	err = h.DB.Where("email = ?", email).First(&user).Error
	if err == nil {
		if user.GoogleID == "" {
			h.DB.Model(&user).Update("google_id", info.Sub)
		}
		h.issueAndRespond(c, user)
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		return
	}

	// New user → provision an org and an admin account with no password.
	orgName := strings.TrimSpace(req.OrgName)
	if orgName == "" {
		orgName = defaultOrgName(info.Name, email)
	}
	name := info.Name
	if name == "" {
		name = email
	}

	org := models.Organization{Name: orgName, PlanTier: "free"}
	user = models.User{
		Email:    email,
		Name:     name,
		Role:     models.RoleAdmin,
		GoogleID: info.Sub,
	}
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

// defaultOrgName derives a friendly org name from the user's name or email domain.
func defaultOrgName(name, email string) string {
	if name != "" {
		return name + "'s Team"
	}
	if at := strings.LastIndex(email, "@"); at >= 0 {
		domain := email[at+1:]
		if dot := strings.Index(domain, "."); dot > 0 {
			domain = domain[:dot]
		}
		if domain != "" {
			return strings.Title(domain) + " Team"
		}
	}
	return "My Team"
}
