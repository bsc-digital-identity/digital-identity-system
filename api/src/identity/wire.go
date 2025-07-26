package identity

import (
	"api/src/queues"
	"gorm.io/gorm"
)

func Build(db *gorm.DB, rabbit *queues.RabbitPublisher) (*Handler, error) {
	repo := NewRepository(db)
	service := NewService(repo, rabbit)
	handler := NewHandler(service)
	return handler, nil
}
