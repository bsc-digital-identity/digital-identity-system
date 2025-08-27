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

// CreateIdentity godoc
// @Summary      Create a new identity
// @Description  Creates an identity with optional parent_id
// @Tags         Identity
// @Accept       json
// @Produce      json
// @Param        body  body      object{identity_name=string,parent_id=int}  true  "Identity info"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/identity [post]
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

// GetIdentity godoc
// @Summary      Get identity by ID
// @Description  Returns identity info
// @Tags         Identity
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Identity ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Router       /v1/identity/{id} [get]
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

// QueueVerification godoc
// @Summary      Queue ZKP verification
// @Description  Queue a zero-knowledge proof verification request
// @Tags         Verification
// @Accept       json
// @Produce      json
// @Param        body  body      model.ZeroKnowledgeProofVerificationRequest  true  "Verification request"
// @Success      202  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /v1/identity/verify [post]
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
