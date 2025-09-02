package dtocommon

import (
	reasoncodes "pkg-common/reason_codes"
	"pkg-common/utilities"
)

type ZkpProofResultDto struct {
	EventId   string `json:"event_id"`
	Signature string `json:"signature"`
	AccountId string `json:"account_id"`
}

func (zkpr ZkpProofResultDto) Serialize() ([]byte, error) {
	return utilities.Serialize[ZkpProofResultDto](zkpr)
}

type ZkpProofFailureDto struct {
	EventId    string                 `json:"event_id"`
	ReqestBody []byte                 `json:"request_body"`
	Error      string                 `json:"error"`
	ReasonCode reasoncodes.ReasonCode `json:"reason_code"`
}

func (zkpf ZkpProofFailureDto) Serialize() ([]byte, error) {
	return utilities.Serialize[ZkpProofFailureDto](zkpf)
}
