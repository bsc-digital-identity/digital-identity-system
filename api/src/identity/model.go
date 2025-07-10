package identity

type SuperIdentity struct {
	Id           int    `gorm:"primaryKey;autoIncrement"`
	IdentityId   string `gorm:"uniqueIndex"`
	IdentityName string `gorm:"uniqueIndex"`
}
