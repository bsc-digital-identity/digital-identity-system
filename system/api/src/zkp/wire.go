package zkp

import (
	"api/src/queues"

	"gorm.io/gorm"
)

func Build(db *gorm.DB, consumer *queues.RabbitConsumer) (*ZeroKnowledgeProofHandler, error) {
	zkpRepo := NewZkpRepository(db)
	zkpService := NewZkpService(zkpRepo)
	zkpHandler := NewZeroKnowledgeProofHandler(zkpService)
	return zkpHandler, nil
}
