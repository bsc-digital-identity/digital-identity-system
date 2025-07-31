package middleware

import (
	"github.com/gin-gonic/gin"
)

func PublicAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: placholder
		c.Next()
	}
}
