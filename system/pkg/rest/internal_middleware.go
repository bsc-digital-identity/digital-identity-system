package rest

import (
	"github.com/gin-gonic/gin"
)

func InternalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: better internal auth
		// Skip for now
		/*
			token := c.GetHeader("Authorization")
			if token == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No Authorization Header Provided"})
				return
			}

			if token != "internal_token_admin_123" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Wrong auth token"})
			}
		*/
		c.Next()
	}
}
