package identity

import (
	"api/src/model"
	"api/src/outbox"
	"fmt"
	"pkg-common/rabbitmq"
	"pkg-common/utilities"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	IdentityRepo      Repository
	OutboxRepo        outbox.OutboxRepository
	RabbitmqPublisher rabbitmq.IRabbitmqPublisher
}

func NewService() *Service {
	// TODO: remove hardcoding
	return &Service{
		IdentityRepo:      NewRepository(),
		OutboxRepo:        outbox.NewRepo(),
		RabbitmqPublisher: rabbitmq.GetPublisher("VerifiersUnverifiedPublisher")}
}

func (s *Service) CreateIdentity(name string, parentId *int) (*model.Identity, error) {
	if parentId != nil {
		parent, err := s.IdentityRepo.GetById(fmt.Sprint(*parentId))
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
	if err := s.IdentityRepo.Create(identity); err != nil {
		return nil, err
	}

	return identity, nil
}

func (s *Service) GetIdentityById(id string) (*model.Identity, error) {
	return s.IdentityRepo.GetById(id)
}

func (s *Service) QueueVerification(req model.ZeroKnowledgeProofVerificationRequest) error {
	schemaId, err := uuid.Parse(req.SchemaId)
	if err != nil {
		return err
	}

	identityId, err := uuid.Parse(req.SchemaId)
	if err != nil {
		return err
	}

	// do validation if types are correct before procedding
	schema, err := s.IdentityRepo.GetSchemaById(schemaId)
	if err != nil {
		return err
	}
	if !validateSchema(schema, req) {
		return fmt.Errorf("Schema validation failed, types are incorrect")
	}

	zkpStringFieds := utilities.Map(req.Fields, func(field model.ZkpField) string {
		val, _ := field.Serialize()
		return string(val)
	})

	msgBody := strings.Join(zkpStringFieds, ", ")

	eventId, err := s.OutboxRepo.NewEvent(
		identityId,
		schemaId,
		msgBody,
	)
	if err != nil {
		return err
	}

	return s.RabbitmqPublisher.Publish(model.ZeroKnowledgeProofToVerification{
		EventId: eventId.String(),
		Data:    msgBody,
	})
}
