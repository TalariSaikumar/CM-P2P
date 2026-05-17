package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORS allows browser clients (local Next.js dev and deployed Vercel origins).
func CORS() gin.HandlerFunc {
	allowed := map[string]struct{}{
		"http://localhost:3000":  {},
		"http://127.0.0.1:3000": {},
	}
	for _, o := range strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			allowed[o] = struct{}{}
		}
	}
	allowVercel := strings.ToLower(strings.TrimSpace(os.Getenv("CORS_ALLOW_VERCEL"))) != "false"

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if originAllowed(origin, allowed, allowVercel) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func originAllowed(origin string, allowed map[string]struct{}, allowVercel bool) bool {
	if origin == "" {
		return false
	}
	if _, ok := allowed[origin]; ok {
		return true
	}
	if allowVercel && strings.HasPrefix(origin, "https://") && strings.HasSuffix(origin, ".vercel.app") {
		return true
	}
	return false
}

// RequestID attaches a simple correlation id for logs (optional hook for later).
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = time.Now().UTC().Format("20060102150405.000000000")
		}
		c.Writer.Header().Set("X-Request-ID", id)
		c.Next()
	}
}
