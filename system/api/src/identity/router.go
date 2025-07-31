package identity

import (
	"github.com/gin-gonic/gin"
)

func RegisterIdentityRoutes(rg *gin.RouterGroup, handler *Handler) {
	// POST /identity         (create identity in DB)
	rg.POST("", handler.CreateIdentity)

	// GET /identity/:id      (get identity by id)
	rg.GET("/:id", handler.GetIdentity)

	// POST /identity/verify  (queue ZKP verification for blockchain client)
	rg.POST("/verify", handler.QueueVerification)

	//rg.POST("/verify-result", HandleVerifyResult)

}
