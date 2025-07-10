package database

type SuperIdentity struct {
	Id           int    `gorm:"primaryKey;autoIncrement"`
	IdentityId   string `gorm:"uniqueIndex"`
	IdentityName string `gorm:"uniqueIndex"`
}

type Attribute struct {
	Id                int `gorm:"primaryKey;autoIncrement"`
	DigitalIdentityId int // FK to SuperIdentity.Id
	Key               string
	Value             string
}

type ZKPProof struct {
	Id                int    `gorm:"primaryKey;autoIncrement"`
	DigitalIdentityId int    // FK to SuperIdentity.Id
	ProofReference    string // e.g. blockchain tx hash or IPFS reference
	Description       string
}
