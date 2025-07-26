package zkp

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

func Build(db *gorm.DB, resultsChannel *amqp.Channel) (*ZeroKnowledgeProofHandler, error) {
	zkpRepo := NewZkpRepository(db)
	zkpService := NewZkpService(zkpRepo)
	zkpHandler := NewZeroKnowledgeProofHandler(zkpService, resultsChannel, "identity.verified.results")
	return zkpHandler, nil
}
