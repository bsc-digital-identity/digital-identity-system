package attribute

type Attribute struct {
	Id                int `gorm:"primaryKey;autoIncrement"`
	DigitalIdentityId int // FK to SuperIdentity.Id
	Key               string
	Value             string
}
