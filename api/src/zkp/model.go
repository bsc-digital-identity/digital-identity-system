package zkp

import (
	authschemas "api/src/auth_schemas"
	"api/src/identity"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"
)

type ZKPProof[T comparable] struct {
	Id                      int                    `gorm:"primaryKey;autoIncrement"`
	DigitalIdentitySchemaId int                    // foreign key
	IdentitySchema          authschemas.Schema[T]  `gorm:"foreignKey:DigitalIdentitySchemaId;references:Id"`
	SuperIdentityId         int                    // foreign key
	SuperIdentity           identity.SuperIdentity `gorm:"foreignKey:SuperIdentityId;references:Id"`
	ProofReference          string
}

// TODO: Naming
type ZkpResponse struct {
	IsProofValid   bool
	ProofReference string
}

type ZkpResult struct {
	Proof         groth16.Proof
	VerifyingKey  groth16.VerifyingKey
	PublicWitness witness.Witness
	TxHash        string
}
