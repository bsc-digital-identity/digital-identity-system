package zkpfailed

import "api/src/model"

type ZkpFailedService interface {
	Todo() error
}

type zkpFailedService struct {
	repo ZkpFailedRepository
}

// Constructor (injects the repo)
func newFailedZkpService() ZkpFailedService {
	return &zkpFailedService{repo: newFailedZkpRepository()}
}

func (zfs *zkpFailedService) Todo() error {
	return zfs.repo.SaveFailedProof(&model.ZkpProofFailure{})
}
