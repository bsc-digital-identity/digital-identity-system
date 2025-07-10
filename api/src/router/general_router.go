package router

import (
	"api/src/identity"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PrepareAppRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	identityGroup := r.Group("/identity")
	identity.RegisterIdentityRoutes(identityGroup, db)

	// Root handler
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello from Go Docker multistage")
	})

	return r
}
