package zkp

import (
	"api/src/model"
	"pkg-common/logger"
)

// Service interface
type ZkpService interface {
	ProcessVerificationResult(resp model.ZeroKnowledgeProofVerificationResponse) error
}

// Implementation with repo dependency
type zkpService struct {
	repo ZkpRepository
}

// Constructor (injects the repo)
func NewZkpService() ZkpService {
	return &zkpService{repo: NewZkpRepository()}
}

// ProcessVerificationResult saves ZKP result into DB
func (s *zkpService) ProcessVerificationResult(resp model.ZeroKnowledgeProofVerificationResponse) error {
	zkpLogger := logger.Default()
	if !resp.IsProofValid {
		zkpLogger.Warnf("Invalid proof, not saving: %s (error: %s)", resp.IdentityId, resp.Error)
		return nil
	}

	// 1. Get identity (by UUID string)
	identity, err := s.repo.GetIdentityByUUID(resp.IdentityId)
	if err != nil {
		return err
	}

	// 2. Find or create the verified schema
	schema, err := s.repo.FindOrCreateVerifiedSchema(resp.Schema, identity.Id)
	if err != nil {
		return err
	}

	// 3. Save the proof
	zkp := &model.ZeroKnowledgeProof{
		DigitalIdentitySchemaId: schema.Id,
		SuperIdentityId:         identity.Id,
		ProofReference:          resp.ProofReference,
	}
	if err := s.repo.SaveZeroKnowledgeProof(zkp); err != nil {
		return err
	}

	zkpLogger.Infof("Saved ZKP proof for identity: %s, schema: %s", resp.IdentityId, schema.SchemaId)
	return nil
}
