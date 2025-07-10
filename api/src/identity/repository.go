package identity

import (
	"gorm.io/gorm"
)

func Create(db *gorm.DB, identity *SuperIdentity) error {
	return db.Create(identity).Error
}

func GetById(db *gorm.DB, id string) (*SuperIdentity, error) {
	var identity SuperIdentity
	err := db.Where("identity_id = ?", id).First(&identity).Error
	return &identity, err
}

func GetByName(db *gorm.DB, name string) (*SuperIdentity, error) {
	var identity SuperIdentity
	err := db.Where("identity_name = ?", name).First(&identity).Error
	return &identity, err
}
