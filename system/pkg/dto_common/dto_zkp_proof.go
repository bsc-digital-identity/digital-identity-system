package dtocommon

import "pkg-common/utilities"

type ZkpProofResultDto struct {
	IdentityId     string `json:"identity_id"`
	ProofReference string `json:"proof_reference"`
	Schema         string `json:"schema"`
}

func (zkpr ZkpProofResultDto) Serialize() ([]byte, error) {
	return utilities.Serialize[ZkpProofResultDto](zkpr)
}

type ZkpProofFailureDto struct {
	IdentityId string `json:"identity_id"`
	Schema     string `json:"schema"`
	ReqestBody []byte `json:"request_body"`
	Error      string `json:"error"`
}

func (zkpf ZkpProofFailureDto) Serialize() ([]byte, error) {
	return utilities.Serialize[ZkpProofFailureDto](zkpf)
}
