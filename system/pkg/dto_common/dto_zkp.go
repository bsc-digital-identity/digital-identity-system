package dtocommon

import "pkg-common/utilities"

type ZeroKnowledgeProofVerificationResultDto struct {
	IdentityId     string `json:"identity_id"`
	ProofReference string `json:"proof_reference"`
	Schema         string `json:"schema"`
}

func (zkpr ZeroKnowledgeProofVerificationResultDto) Serialize() ([]byte, error) {
	return utilities.Serialize[ZeroKnowledgeProofVerificationResultDto](zkpr)
}

type ZeroKnowledgeProofVerificationFailureDto struct {
	IdentityId     string `json:"identity_id"`
	ProofReference string `json:"proof_reference"`
	Schema         string `json:"schema"`
	Error          string `json:"error"`
}

func (zkpf ZeroKnowledgeProofVerificationFailureDto) Serialize() ([]byte, error) {
	return utilities.Serialize[ZeroKnowledgeProofVerificationResultDto](zkpf)
}
