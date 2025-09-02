package model

type ZkpProofFailure struct {
	Id          int `gorm:"primaryKey;autoIncrement"`
	EventId     string
	RequestBody []byte
	Error       string
	ReasonCode  string
}
