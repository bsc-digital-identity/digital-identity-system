package identity

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

// RegisterIdentityRoutes registers endpoints on the Gin router group
func RegisterIdentityRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	// POST /identity
	rg.POST("", func(c *gin.Context) {
		var req struct {
			IdentityName string `json:"identity_name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.IdentityName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		identity, err := CreateIdentity(db, req.IdentityName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create identity: " + err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"identity_id":   identity.IdentityId,
			"identity_name": identity.IdentityName,
		})
	})

	// GET /identity/:id
	rg.GET("/:id", func(c *gin.Context) {
		id := c.Param("id")
		identity, err := GetIdentityById(db, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Identity not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"identity_id":   identity.IdentityId,
			"identity_name": identity.IdentityName,
		})
	})
}
