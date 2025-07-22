package zkp

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ZkpService interface {
	AddNew(proof ZKPProof) error
	UpdateBlockchainRef(superIdnenityId, identitySchemaId uuid.UUID, newRef string) error
	AuthUser(superIdnenityId, identitySchemaId uuid.UUID) (*ZkpResponse, error)
}

type BlockchainPlacheloder struct {
	Acc int
}

type BlockhainInterfacePlaceholder interface {
	GetZkp(proofRef string) ZkpResult
}

type ZkpBlockchainService struct {
	db           *gorm.DB
	blockhainAcc BlockhainInterfacePlaceholder
}

func NewZkpService(db *gorm.DB) *ZkpBlockchainService {
	// TODO: implement this
	return &ZkpBlockchainService{db: db, blockhainAcc: nil}
}

func (s *ZkpBlockchainService) AddNew(proof ZKPProof) error {
	result := s.db.Create(&proof)
	return result.Error
}

func (s *ZkpBlockchainService) UpdateBlockchainRef(superIdnenityId, identitySchemaId uuid.UUID, newRef string) error {
	// TODO: adjust query later
	return s.db.Model(&ZKPProof{}).
		Joins("JOIN super_identitiy ON zkp_proof.super_identity_id = super_identities.id").
		Where("super_identitiy.uuid = ?", superIdnenityId).
		Where("idnenity_schema.uuid = ?", identitySchemaId).
		Update("zkp_proofs.proof_reference", newRef).Error
}

func (s *ZkpBlockchainService) AuthUser(superIdnenityId, identitySchemaId uuid.UUID) (*ZkpResponse, error) {
	var zkpProof ZKPProof
	// TODO: adjust query later
	result := s.db.First(&zkpProof)

	if result.Error != nil {
		return nil, result.Error
	}

	// TODO: implement zkp verification later
	//zkpResult := s.blockhainAcc.GetZkp(zkpProof.ProofReference)

	// verify here
	// verified, err := verifier(zkpResult)

	return &ZkpResponse{true, zkpProof.ProofReference}, nil
}
