package model

import (
	"pkg-common/utilities"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"
)

type ZeroKnowledgeProof struct {
	Id                      int `gorm:"primaryKey;autoIncrement"`
	DigitalIdentitySchemaId int // foreign key to VerifiedSchema
	SuperIdentityId         int // foreign key
	ProofReference          string
	AccountId               string
}

type ZeroKnowledgeProofVerificationRequest struct {
	IdentityId string     `json:"identity_id"`
	SchemaId   string     `json:"schema_id"`
	Fields     []ZkpField `json:"data"`
}

func (req ZeroKnowledgeProofVerificationRequest) Serialize() ([]byte, error) {
	return utilities.Serialize[ZeroKnowledgeProofVerificationRequest](req)
}

type ZkpField struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

func (req ZkpField) Serialize() ([]byte, error) {
	return utilities.Serialize[ZkpField](req)
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

type ZeroKnowledgeProofToVerification struct {
	EventId string `json:"event_id"`
	Data    string `json:"data"`
}

func (req ZeroKnowledgeProofToVerification) Serialize() ([]byte, error) {
	return utilities.Serialize[ZeroKnowledgeProofToVerification](req)
}
