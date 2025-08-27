package identity

import (
	"api/src/database"
	"api/src/model"

	"gorm.io/gorm"
)

type Repository interface {
	Create(identity *model.Identity) error
	GetById(id string) (*model.Identity, error)
	GetByName(name string) (*model.Identity, error)
	GetSubIdentities(parentId int) ([]model.Identity, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository() Repository {
	return &gormRepository{db: database.GetDatabaseConnection()}
}

func (r *gormRepository) Create(identity *model.Identity) error {
	return r.db.Create(identity).Error
}

func (r *gormRepository) GetById(id string) (*model.Identity, error) {
	var identity model.Identity
	err := r.db.Where("identity_id = ?", id).First(&identity).Error
	return &identity, err
}

func (r *gormRepository) GetByName(name string) (*model.Identity, error) {
	var identity model.Identity
	err := r.db.Where("identity_name = ?", name).First(&identity).Error
	return &identity, err
}

func (r *gormRepository) GetSubIdentities(parentId int) ([]model.Identity, error) {
	var subs []model.Identity
	err := r.db.Where("parent_id = ?", parentId).Find(&subs).Error
	return subs, err
}
