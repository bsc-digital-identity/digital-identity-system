package identity

import (
	"api/src/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func NewHandler() *Handler {
	return &Handler{Service: NewService()}
}

func (h *Handler) CreateIdentity(c *gin.Context) {
	var req struct {
		IdentityName string `json:"identity_name"`
		ParentId     *int   `json:"parent_id,omitempty"` // Accept parent_id as optional
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.IdentityName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	identity, err := h.Service.CreateIdentity(req.IdentityName, req.ParentId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create identity: " + err.Error()})
		return
	}

	var _ *model.Identity
	if identity.ParentId != nil {
		_, _ = h.Service.GetIdentityById(identity.Parent.IdentityId)
	}

	c.JSON(http.StatusCreated, gin.H{
		"identity_id":   identity.IdentityId,
		"identity_name": identity.IdentityName,
		"parent_id":     identity.ParentId,
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
	var req model.ZeroKnowledgeProofVerificationRequest
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
