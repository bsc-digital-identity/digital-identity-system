package zkp

import (
	"api/src/model"
	"log"
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
func NewZkpService(repo ZkpRepository) ZkpService {
	return &zkpService{repo: repo}
}

// ProcessVerificationResult saves ZKP result into DB
func (s *zkpService) ProcessVerificationResult(resp model.ZeroKnowledgeProofVerificationResponse) error {
	if !resp.IsProofValid {
		log.Printf("Invalid proof, not saving: %s (error: %s)", resp.IdentityId, resp.Error)
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

	log.Printf("Saved ZKP proof for identity: %s, schema: %s", resp.IdentityId, schema.SchemaId)
	return nil
}
