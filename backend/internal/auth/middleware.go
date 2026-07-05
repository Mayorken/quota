package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Context keys for values set by the auth middleware.
const (
	CtxUserID = "userID"
	CtxOrgID  = "orgID"
	CtxRole   = "role"
)

// Middleware validates the bearer token and stores identity in the context.
func Middleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := ParseToken(secret, tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxOrgID, claims.OrgID)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}

// RequireRole ensures the caller has one of the allowed roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role := c.GetString(CtxRole)
		if !allowed[role] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

// OrgID / UserID / Role helpers read identity from the request context.
func OrgID(c *gin.Context) string  { return c.GetString(CtxOrgID) }
func UserID(c *gin.Context) string { return c.GetString(CtxUserID) }
func Role(c *gin.Context) string   { return c.GetString(CtxRole) }

// IsManager reports whether the caller can see team-wide data and edit config.
func IsManager(c *gin.Context) bool {
	r := Role(c)
	return r == "manager" || r == "admin"
}
