package model

import (
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"
)

type ZeroKnowledgeProof struct {
	Id                      int            `gorm:"primaryKey;autoIncrement"`
	DigitalIdentitySchemaId int            // foreign key to VerifiedSchema
	DigitalIdentitySchema   VerifiedSchema `gorm:"foreignKey:DigitalIdentitySchemaId;references:Id"`
	SuperIdentityId         int            // foreign key
	SuperIdentity           Identity       `gorm:"foreignKey:SuperIdentityId;references:Id"`
	ProofReference          string
}

type ZeroKnowledgeProofVerificationRequest struct {
	IdentityId string `json:"identity_id"`
	Schema     string `json:"schema"`
}

// TODO: Naming
type ZeroKnowledgeProofVerificationResponse struct {
	IdentityId     string `json:"identity_id"`
	IsProofValid   bool   `json:"is_proof_valid"`
	ProofReference string `json:"proof_reference"`
	Schema         string `json:"schema"` // echo back what was used
	Error          string `json:"error,omitempty"`
}

type ZeroKnowledgeProofResult struct {
	Proof         groth16.Proof
	VerifyingKey  groth16.VerifyingKey
	PublicWitness witness.Witness
	TxHash        string
}
