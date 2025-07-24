package model

type Identity struct {
	Id            int        `gorm:"primaryKey;autoIncrement"`
	IdentityId    string     `gorm:"uniqueIndex"` // public/business ID (e.g., UUID)
	IdentityName  string     `gorm:"uniqueIndex"` // human-readable name
	ParentId      *int       // references Id of parent identity (nullable)
	Parent        *Identity  `gorm:"foreignKey:ParentId"` // Parent entity
	SubIdentities []Identity `gorm:"foreignKey:ParentId"` // Child entities
}
