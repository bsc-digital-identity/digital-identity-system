package identity

import (
	"api/src/model"
	"api/src/queues"
	"fmt"
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

func (s *Service) CreateIdentity(name string, parentId *int) (*model.Identity, error) {
	if parentId != nil {
		var parent model.Identity
		if err := s.DB.First(&parent, *parentId).Error; err != nil {
			return nil, fmt.Errorf("parent identity with Id %d does not exist", *parentId)
		}
	}

	id := uuid.New().String()
	identity := &model.Identity{
		IdentityId:   id,
		IdentityName: name,
		ParentId:     parentId,
	}
	if err := Create(s.DB, identity); err != nil {
		return nil, err
	}
	//
	//// Queue for blockchain verification
	//msg := queues.ZkpVerifiedMessage{
	//	IdentityId: id,
	//	BirthDay:   birthDay,
	//	BirthMonth: birthMonth,
	//	BirthYear:  birthYear,
	//}
	//_ = s.Rabbit.PublishZkpVerificationRequest(msg) // Optionally, handle/log error

	return identity, nil
}

func (s *Service) GetIdentityById(id string) (*model.Identity, error) {
	return GetById(s.DB, id)
}

// New: queue a ZKP verification message
func (s *Service) QueueVerification(req model.ZeroKnowledgeProofVerificationRequest) error {
	return s.Rabbit.PublishZkpVerificationRequest(req)
}
