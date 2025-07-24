package identity

import (
	"api/src/model"
	"gorm.io/gorm"
)

func Create(db *gorm.DB, identity *model.Identity) error {
	return db.Create(identity).Error
}

func GetById(db *gorm.DB, id string) (*model.Identity, error) {
	var identity model.Identity
	err := db.Where("identity_id = ?", id).First(&identity).Error
	return &identity, err
}

func GetByName(db *gorm.DB, name string) (*model.Identity, error) {
	var identity model.Identity
	err := db.Where("identity_name = ?", name).First(&identity).Error
	return &identity, err
}

func GetSubIdentities(db *gorm.DB, parentId int) ([]model.Identity, error) {
	var subs []model.Identity
	err := db.Where("parent_id = ?", parentId).Find(&subs).Error
	return subs, err
}
