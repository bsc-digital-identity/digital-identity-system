package zkp

//import (
//	"api/src/middleware"
//
//	"github.com/gin-gonic/gin"
//	"gorm.io/gorm"
//)
//
//func RegisterZkpRoutes(rg *gin.RouterGroup, db *gorm.DB) {
//	service := NewZkpService(db)
//	handler := NewZeroKnowledgeProofHandler(service)
//
//	internal := rg.Group("internal/identity")
//	internal.Use(middleware.InternalAuthMiddleware())
//	{
//		internal.POST("/create", handler.AddVerifiedIdentity)
//		internal.PATCH("/update", handler.UpdateVerifiedIdentity)
//	}
//
//	public := rg.Group("identity")
//	public.Use(middleware.PublicAuthMiddleware())
//	{
//		public.GET("/auth", handler.AuthorizeIdentity)
//	}
//}
