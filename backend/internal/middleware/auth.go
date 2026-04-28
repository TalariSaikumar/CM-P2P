package middleware

import (
	"net/http"
	"strings"

	"carmanage/backend/internal/auth"
	"carmanage/backend/internal/config"
	"carmanage/backend/internal/httpx"

	"github.com/gin-gonic/gin"
)

const (
	ctxUserID   = "user_id"
	ctxUserRole = "user_role"
	ctxUserEmail = "user_email"
)

// AuthRequired validates Bearer JWT and sets user_id, user_role, user_email on the context.
func AuthRequired(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			httpx.Abort(c, http.StatusUnauthorized, "UNAUTHORIZED", "You need to sign in to continue.")
			return
		}
		raw := strings.TrimSpace(h[7:])
		claims, err := auth.ParseAccessToken(cfg.JWTSecret, raw)
		if err != nil {
			httpx.Abort(c, http.StatusUnauthorized, "UNAUTHORIZED", "Your session is invalid or has expired. Please sign in again.")
			return
		}
		uid, err := claims.UserID()
		if err != nil {
			httpx.Abort(c, http.StatusUnauthorized, "UNAUTHORIZED", "Your session is invalid. Please sign in again.")
			return
		}
		c.Set(ctxUserID, uid.String())
		c.Set(ctxUserRole, claims.Role)
		c.Set(ctxUserEmail, claims.Email)
		c.Next()
	}
}

// UserID returns the authenticated user id string from Gin context.
func UserID(c *gin.Context) (string, bool) {
	v, ok := c.Get(ctxUserID)
	if !ok {
		return "", false
	}
	s, _ := v.(string)
	return s, s != ""
}

// UserRole returns the authenticated user's role.
func UserRole(c *gin.Context) (string, bool) {
	v, ok := c.Get(ctxUserRole)
	if !ok {
		return "", false
	}
	s, _ := v.(string)
	return s, s != ""
}
