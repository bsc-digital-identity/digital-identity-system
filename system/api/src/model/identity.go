package model

type Identity struct {
	Id           int       `gorm:"primaryKey;autoIncrement"`
	IdentityId   string    // public/business ID (e.g., UUID)
	IdentityName string    // human-readable name
	ParentId     *int      // references Id of parent identity (nullable)
	Parent       *Identity `gorm:"-"` // ignored by GORM for migration
}
