package zkp

import (
	"gorm.io/gorm"
)

import "api/src/queues"

func Build(db *gorm.DB, consumer *queues.RabbitConsumer) (*ZeroKnowledgeProofHandler, error) {
	zkpRepo := NewZkpRepository(db)
	zkpService := NewZkpService(zkpRepo)
	zkpHandler := NewZeroKnowledgeProofHandler(zkpService, consumer)
	return zkpHandler, nil
}
