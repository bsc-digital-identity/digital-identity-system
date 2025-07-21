package zkp

import (
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

// IdentityCircuit defines the circuit structure
type IdentityCircuit struct {
	AgeDay   frontend.Variable `gnark:",secret"`
	AgeMonth frontend.Variable `gnark:",secret"`
	AgeYear  frontend.Variable `gnark:",secret"`
}

// Define implements the frontend.Circuit interface
func (circuit *IdentityCircuit) Define(api frontend.API) error {
	currentTime := time.Now()
	currentYear := currentTime.Year()
	minValidYear := api.Sub(currentYear, 18)
	api.AssertIsLessOrEqual(circuit.AgeYear, minValidYear)

	currentMonth := int(currentTime.Month())
	api.AssertIsLessOrEqual(circuit.AgeMonth, currentMonth)

	currentDay := currentTime.Day()
	api.AssertIsLessOrEqual(circuit.AgeDay, currentDay)

	api.AssertIsLessOrEqual(1, circuit.AgeDay)
	api.AssertIsLessOrEqual(circuit.AgeDay, 31)
	api.AssertIsLessOrEqual(1, circuit.AgeMonth)
	api.AssertIsLessOrEqual(circuit.AgeMonth, 12)
	return nil
}

type ZkpResult struct {
	Proof         groth16.Proof
	VerifyingKey  groth16.VerifyingKey
	PublicWitness witness.Witness
	TxHash        string
}

func CreateZKP(birthDay, birthMonth, birthYear int) (*ZkpResult, error) {
	// 1. Compile the circuit (constraint system)
	ccs, err := frontend.Compile(
		ecc.BN254.ScalarField(),
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
