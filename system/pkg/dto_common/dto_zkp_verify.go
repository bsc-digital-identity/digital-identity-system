package dtocommon

import "pkg-common/utilities"

type ZkpVerificationResponseDto struct {
	Siganture      string `json:"identity_id"`
	ProofReference string `json:"proof_reference"`
}

func (zkpr ZkpVerificationResponseDto) Serialize() ([]byte, error) {
	return utilities.Serialize[ZkpProofResultDto](zkpr)
}
