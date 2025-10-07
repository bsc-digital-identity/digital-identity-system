package model

import (
	"time"

	"gorm.io/gorm"
)

type OutboxEvent struct {
	Id             uint           `gorm:"primaryKey;autoIncrement"`
	EventId        string         `gorm:"uniqueIndex;type:uuid;not null"`
	IdentityId     string         `gorm:"type:uuid;not null"`
	SchemaId       string         `gorm:"type:uuid;not null"`
	Retry          int            `gorm:"default:0"`
	ToProcess      bool           `gorm:"default:false;index"`
	RequestMessage string         `gorm:"type:text;not null"`
	ProcessedAt    gorm.DeletedAt `gorm:"index"`
	CreatedAt      time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// Foreign key relationships
	Identity Identity       `gorm:"foreignKey:IdentityId;references:IdentityId"`
	Schema   VerifiedSchema `gorm:"foreignKey:SchemaId;references:SchemaId"`
}

func (oe OutboxEvent) MapToZkpVerifcationRequest() ZeroKnowledgeProofToVerification {
	return ZeroKnowledgeProofToVerification{
		EventId: oe.EventId,
		Data:    oe.RequestMessage, // might need to strip some data here
	}
}
