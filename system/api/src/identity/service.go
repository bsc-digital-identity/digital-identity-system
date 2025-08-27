package identity

import (
	"api/src/model"
	"fmt"
	"pkg-common/rabbitmq"

	"github.com/google/uuid"
)

type Service struct {
	Repo              Repository
	RabbitmqPublisher rabbitmq.IRabbitmqPublisher
}

func NewService() *Service {
	// TODO: remove hardcoding
	return &Service{Repo: NewRepository(), RabbitmqPublisher: rabbitmq.GetPublisher("VerifiersUnverifiedPublisher")}
}

func (s *Service) CreateIdentity(name string, parentId *int) (*model.Identity, error) {
	if parentId != nil {
		parent, err := s.Repo.GetById(fmt.Sprint(*parentId))
		if err != nil {
			return nil, fmt.Errorf("parent identity with Id %d does not exist", *parentId)
		}
		_ = parent
	}

	id := uuid.New().String()
	identity := &model.Identity{
		IdentityId:   id,
		IdentityName: name,
		ParentId:     parentId,
	}
	if err := s.Repo.Create(identity); err != nil {
		return nil, err
	}

	return identity, nil
}

func (s *Service) GetIdentityById(id string) (*model.Identity, error) {
	return s.Repo.GetById(id)
}

func (s *Service) QueueVerification(req model.ZeroKnowledgeProofVerificationRequest) error {
	return s.RabbitmqPublisher.Publish(req)
}
