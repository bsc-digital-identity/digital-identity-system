package zkp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ZkpHandler struct {
	service ZkpService
}

func NewZkpHandler(service ZkpService) *ZkpHandler {
	return &ZkpHandler{service: service}
}

func (h *ZkpHandler) AddVerifiedIdentity(c *gin.Context) {
	var zkpProof ZKPProof
	if err := c.ShouldBindJSON(&zkpProof); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.service.AddNew(zkpProof)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save identity in database: " + err.Error()})
		return
	}
}

func (h *ZkpHandler) UpdateVerifiedIdentity(c *gin.Context) {
	var req struct {
		SuperIdnenityId     uuid.UUID `json:"super_identity_id"`
		IdentitySchemaId    uuid.UUID `json:"identity_schema_id"`
		BlockchainReference string    `json:"blockcahin_ref"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.service.UpdateBlockchainRef(req.SuperIdnenityId, req.IdentitySchemaId, req.BlockchainReference)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update blockachain reference: " + err.Error()})
		return
	}
}

func (h *ZkpHandler) AuthorizeIdentity(c *gin.Context) {
	superIdnenityId := uuid.MustParse(c.Query("superId"))
	identitySchemaId := uuid.MustParse(c.Query("schemaId"))

	authResult, err := h.service.AuthUser(superIdnenityId, identitySchemaId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authorization unsucessfull: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, authResult)
}
