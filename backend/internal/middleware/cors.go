package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// CORS allows the Next.js dev server on localhost:3000 to call the API.
func CORS() gin.HandlerFunc {
	allowed := map[string]struct{}{
		"http://localhost:3000":  {},
		"http://127.0.0.1:3000": {},
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowed[origin]; ok {
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
