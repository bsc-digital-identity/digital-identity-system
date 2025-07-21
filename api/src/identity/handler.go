package identity

import (
	"api/src/queues"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) CreateIdentity(c *gin.Context) {
	var req struct {
		IdentityName string `json:"identity_name"`
		BirthDay     int    `json:"birth_day"`
		BirthMonth   int    `json:"birth_month"`
		BirthYear    int    `json:"birth_year"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.IdentityName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	identity, err := h.Service.CreateIdentity(req.IdentityName, req.BirthDay, req.BirthMonth, req.BirthYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create identity: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"identity_id":   identity.IdentityId,
		"identity_name": identity.IdentityName,
	})
}

func (h *Handler) GetIdentity(c *gin.Context) {
	id := c.Param("id")
	identity, err := h.Service.GetIdentityById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Identity not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"identity_id":   identity.IdentityId,
		"identity_name": identity.IdentityName,
	})
}

func (h *Handler) QueueVerification(c *gin.Context) {
	var req queues.ZkpVerifiedMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	if err := h.Service.QueueVerification(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue verification"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"status": "queued"})
}

type VerifyResult struct {
	IdentityID string `json:"identity_id" binding:"required"`
	TxHash     string `json:"tx_hash" binding:"required"`
}

func HandleVerifyResult(c *gin.Context) {
	var result VerifyResult
	if err := c.ShouldBindJSON(&result); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.String(200, result.TxHash)
}
