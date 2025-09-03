package zkpfailed

import (
	"api/src/model"
	"api/src/outbox"
	dtocommon "pkg-common/dto_common"

	"github.com/google/uuid"
)

type ZkpFailedService interface {
	SaveFailedAndUpdateOutbox(dtocommon.ZkpProofFailureDto) error
}

type zkpFailedService struct {
	zkpRepo    ZkpFailedRepository
	outboxRepo outbox.OutboxRepository
}

// Constructor (injects the repo)
func newFailedZkpService() ZkpFailedService {
	return &zkpFailedService{
		zkpRepo:    newFailedZkpRepository(),
		outboxRepo: outbox.NewRepo(),
	}
}

func (zfs *zkpFailedService) SaveFailedAndUpdateOutbox(data dtocommon.ZkpProofFailureDto) error {
	entity := model.ZkpProofFailure{
		EventId:     data.EventId,
		RequestBody: data.ReqestBody,
		Error:       data.Error,
		ReasonCode:  string(data.ReasonCode),
	}

	err := zfs.zkpRepo.SaveFailedProof(&entity)
	if err != nil {
		return err
	}

	return zfs.outboxRepo.UpdateRetryValue(uuid.MustParse(data.EventId))
}
