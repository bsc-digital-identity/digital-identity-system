package model

import "gorm.io/gorm"

type OutboxEvent struct {
	Id             int    `gorm:"primaryKey;autoIncrement"`
	EventId        string `gorm:"uniqueIndex"`
	IdentityId     string // convert to FK
	SchemaId       string // convert to FK
	Retry          int
	ToProcess      bool
	RequestMessage string
	ProcessedAt    gorm.DeletedAt
	CreatedAt      string
}

func (oe OutboxEvent) MapToZkpVerifcationRequest() ZeroKnowledgeProofToVerification {
	return ZeroKnowledgeProofToVerification{
		EventId: oe.EventId,
		Data:    oe.RequestMessage, // might need to strip some data here
	}
}
