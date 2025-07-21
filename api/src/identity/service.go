package identity

import (
	"api/src/queues"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	DB     *gorm.DB
	Rabbit *queues.RabbitPublisher
}

func NewService(db *gorm.DB, rabbit *queues.RabbitPublisher) *Service {
	return &Service{DB: db, Rabbit: rabbit}
}

func (s *Service) CreateIdentity(name string, birthDay, birthMonth, birthYear int) (*SuperIdentity, error) {
	id := uuid.New().String()
	identity := &SuperIdentity{
		IdentityId:   id,
		IdentityName: name,
	}
	err := Create(s.DB, identity)
	if err != nil {
		return nil, err
	}

	// Queue for blockchain verification
	msg := queues.ZkpVerifiedMessage{
		IdentityId: id,
		BirthDay:   birthDay,
		BirthMonth: birthMonth,
		BirthYear:  birthYear,
	}
	_ = s.Rabbit.PublishZkpVerified(msg) // Optionally, handle/log error

	return identity, nil
}

func (s *Service) GetIdentityById(id string) (*SuperIdentity, error) {
	return GetById(s.DB, id)
}

// New: queue a ZKP verification message
func (s *Service) QueueVerification(req queues.ZkpVerifiedMessage) error {
	return s.Rabbit.PublishZkpVerified(req)
}
