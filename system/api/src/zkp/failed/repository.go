package zkpfailed

import (
	"api/src/database"
	"api/src/model"

	"gorm.io/gorm"
)

type ZkpFailedRepository interface {
	SaveFailedProof(*model.ZkpProofFailure) error
}

func newFailedZkpRepository() ZkpFailedRepository {
	return &zkpFailedRepository{db: database.GetDatabaseConnection()}
}

type zkpFailedRepository struct {
	db *gorm.DB
}

func (zfr *zkpFailedRepository) SaveFailedProof(entity *model.ZkpProofFailure) error {
	return zfr.db.Create(entity).Error
}
