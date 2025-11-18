package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func PublicAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// tutaj docelowo np. auth na podstawie nagłówków,
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// DEV: konkretny frontend (wallet) – lepiej niż "*"
		origin := "http://192.168.8.107:8081"

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
