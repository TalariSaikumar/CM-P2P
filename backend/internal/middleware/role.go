package middleware

import (
	"net/http"

	"carmanage/backend/internal/httpx"
	"carmanage/backend/internal/models"

	"github.com/gin-gonic/gin"
)

// RequireRole blocks the request unless the JWT role matches one of the allowed roles.
func RequireRole(allowed ...models.UserRole) gin.HandlerFunc {
	set := make(map[string]struct{}, len(allowed))
	for _, r := range allowed {
		set[string(r)] = struct{}{}
	}
	return func(c *gin.Context) {
		role, ok := UserRole(c)
		if !ok {
			httpx.Abort(c, http.StatusUnauthorized, "UNAUTHORIZED", "You need to sign in to continue.")
			return
		}
		if _, ok := set[role]; !ok {
			httpx.Abort(c, http.StatusForbidden, "FORBIDDEN", "This action is not available for your account type.")
			return
		}
		c.Next()
	}
}
