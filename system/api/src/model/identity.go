package model

type Identity struct {
	Id           int       `gorm:"primaryKey;autoIncrement"`
	IdentityId   string    `gorm:"uniqueIndex;type:uuid;not null"` // public/business ID (e.g., UUID)
	IdentityName string    `gorm:"not null"` // human-readable name
	ParentId     *int      // references Id of parent identity (nullable)
	Parent       *Identity `gorm:"-"` // ignored by GORM for migration
}
