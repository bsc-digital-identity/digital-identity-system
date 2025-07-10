package zkp

type ZKPProof struct {
	Id                int    `gorm:"primaryKey;autoIncrement"`
	DigitalIdentityId int    // FK to SuperIdentity.Id
	ProofReference    string // e.g. blockchain tx hash or IPFS reference
	Description       string
}
