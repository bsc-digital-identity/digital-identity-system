package zkpresult

import (
	"api/src/database"
	"api/src/model"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ZkpRepository interface {
	GetIdentityByUUID(uuid string) (model.Identity, error)
	FindOrCreateVerifiedSchema(schemaStr string, superIdentityId int) (model.VerifiedSchema, error)
	SaveZeroKnowledgeProof(zkp *model.ZeroKnowledgeProof) error
}

func NewZkpRepository() ZkpRepository {
	return &zkpRepository{db: database.GetDatabaseConnection()}
}

type zkpRepository struct {
	db *gorm.DB
}

func (r *zkpRepository) GetIdentityByUUID(uuid string) (model.Identity, error) {
	var identity model.Identity
	err := r.db.Where("identity_id = ?", uuid).First(&identity).Error
	return identity, err
}

func (r *zkpRepository) FindOrCreateVerifiedSchema(schemaStr string, superIdentityId int) (model.VerifiedSchema, error) {
	var schema model.VerifiedSchema
	err := r.db.Where("schema = ?", schemaStr).First(&schema).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		schema = model.VerifiedSchema{
			SchemaId:        uuid.New().String(),
			SuperIdentityId: superIdentityId,
			Schema:          schemaStr,
		}
		err = r.db.Create(&schema).Error
	}
	return schema, err
}

func (r *zkpRepository) SaveZeroKnowledgeProof(zkp *model.ZeroKnowledgeProof) error {
	return r.db.Create(zkp).Error
}
