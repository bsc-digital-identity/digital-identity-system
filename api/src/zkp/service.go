package zkp

import (
	"api/src/model"
	"encoding/json"
	"gorm.io/gorm"
	"log"
)

// Service interface
type ZkpService interface {
	ProcessVerificationResult(resp model.ZeroKnowledgeProofVerificationResponse) error
}

// Implementation
type ZkpBlockchainService struct {
	db *gorm.DB
}

func NewZkpService(db *gorm.DB) *ZkpBlockchainService {
	return &ZkpBlockchainService{db: db}
}

// Just log the result (later, you could insert into DB if needed)
func (s *ZkpBlockchainService) ProcessVerificationResult(resp model.ZeroKnowledgeProofVerificationResponse) error {
	// Log everything for now
	data, _ := json.Marshal(resp)
	log.Printf("Received ZKP Verification Result: %s", data)
	return nil
}

// -- Placeholders for future blockchain interaction --
type BlockhainInterfacePlaceholder interface {
	GetZkp(proofRef string) model.ZeroKnowledgeProofResult
}
