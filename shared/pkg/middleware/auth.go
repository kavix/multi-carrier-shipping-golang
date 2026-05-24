package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization"})
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid auth format"})
			return
		}
		c.Set("user_id", parts[1])
		c.Next()
	}
}

// DownstreamContextMiddleware extracts user_id from X-User-ID header (set by API Gateway)
// Used by downstream services to populate the Gin context with user information
func DownstreamContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get user_id from X-User-ID header (forwarded by API Gateway)
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			c.Set("user_id", userID)
		}
		c.Next()
	}
}
