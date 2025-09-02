package dtocommon

import (
	reasoncodes "pkg-common/reason_codes"
	"pkg-common/utilities"
)

type ZkpProofDtoFactory interface {
	CreateErrorDto(error, reasoncodes.ReasonCode) utilities.Serializable
	CreateInfoDto(reasoncodes.ReasonCode) utilities.Serializable
}

type zkpProofFailureDtoFactory struct {
	EventId    string
	ReqestBody []byte
}

func NewZkpProofFailureFactory(eventId string, requestBody []byte) ZkpProofDtoFactory {
	return zkpProofFailureDtoFactory{
		EventId:    eventId,
		ReqestBody: requestBody,
	}
}

func (zpfdf zkpProofFailureDtoFactory) CreateErrorDto(
	err error,
	reasonCode reasoncodes.ReasonCode) utilities.Serializable {
	return ZkpProofFailureDto{
		EventId:    zpfdf.EventId,
		ReqestBody: zpfdf.ReqestBody,
		Error:      err.Error(),
		ReasonCode: reasonCode,
	}
}

func (zpfdf zkpProofFailureDtoFactory) CreateInfoDto(reasonCode reasoncodes.ReasonCode) utilities.Serializable {
	return ZkpProofFailureDto{
		EventId:    zpfdf.EventId,
		ReqestBody: zpfdf.ReqestBody,
		ReasonCode: reasonCode,
	}
}
