package identity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateIdentity(db *gorm.DB, name string) (*SuperIdentity, error) {
	id := uuid.New().String()
	identity := &SuperIdentity{
		IdentityId:   id,
		IdentityName: name,
	}
	err := Create(db, identity)
	return identity, err
}

func GetIdentityById(db *gorm.DB, id string) (*SuperIdentity, error) {
	return GetById(db, id)
}
