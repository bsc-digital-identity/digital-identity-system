package zkpresult

import (
	"api/src/model"
	"api/src/outbox"
	dtocommon "pkg-common/dto_common"
	"pkg-common/logger"

	"github.com/google/uuid"
)

// Service interface
type ZkpService interface {
	ProcessVerificationResult(resp dtocommon.ZkpProofResultDto) error
}

// Implementation with repo dependency
type zkpService struct {
	identityRepo ZkpRepository
	outboxRepo   outbox.OutboxRepository
}

// Constructor (injects the repo)
func NewZkpService() ZkpService {
	return &zkpService{
		identityRepo: NewZkpRepository(),
		outboxRepo:   outbox.NewRepo(),
	}
}

// ProcessVerificationResult saves ZKP result into DB
func (s *zkpService) ProcessVerificationResult(resp dtocommon.ZkpProofResultDto) error {
	zkpLogger := logger.Default()

	event, err := s.outboxRepo.GetEvent(uuid.MustParse(resp.EventId))

	// 1. Get identity (by UUID string)
	identity, err := s.identityRepo.GetIdentityByUUID(event.IdentityId)
	if err != nil {
		return err
	}

	// 2. Find or create the verified schema
	schema, err := s.identityRepo.FindOrCreateVerifiedSchema(event.SchemaId, identity.Id)
	if err != nil {
		return err
	}

	// 3. Save the proof
	zkp := &model.ZeroKnowledgeProof{
		DigitalIdentitySchemaId: schema.Id,
		SuperIdentityId:         identity.Id,
		ProofReference:          resp.Signature,
		AccountId:               resp.AccountId,
	}
	if err := s.identityRepo.SaveZeroKnowledgeProof(zkp); err != nil {
		return err
	}

	zkpLogger.Infof("Saved ZKP proof for identity: %s, schema: %s", event.IdentityId, schema.SchemaId)
	zkpLogger.Infof("Resolved and deleted event: %s", event.EventId)
	return s.outboxRepo.MarkEventAsProcessed(uuid.MustParse(event.EventId))
}
