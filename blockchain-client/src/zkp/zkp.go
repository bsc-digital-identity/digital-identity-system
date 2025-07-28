package zkp

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

const (
	ElipticalCurveID = ecc.BN254
)

func CreateZKP(birthDay, birthMonth, birthYear int) (*ZkpResult, error) {
	// 1. Compile the circuit (constraint system)
	ccs, err := frontend.Compile(
		ElipticalCurveID.ScalarField(),
		r1cs.NewBuilder,
		&IdentityCircuit{},
	)
	if err != nil {
		return nil, err
	}

	// 2. Setup proving/verifying keys
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		return nil, err
	}

	// 3. Assign inputs
	assignment := IdentityCircuit{
		AgeDay:   birthDay,
		AgeMonth: birthMonth,
		AgeYear:  birthYear,
	}

	fullWitness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	if err != nil {
		return nil, err
	}

	// 4. Create the proof
	proof, err := groth16.Prove(ccs, pk, fullWitness)
	if err != nil {
		return nil, err
	}

	// 5. Get the public witness
	publicWitness, err := fullWitness.Public()
	if err != nil {
		return nil, err
	}

	return &ZkpResult{
		Proof:         proof,
		VerifyingKey:  vk,
		PublicWitness: publicWitness,
		TxHash:        "",
	}, nil
}
