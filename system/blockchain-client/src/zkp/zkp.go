package zkp

import (
	"blockchain-client/src/types/domain"
	"encoding/json"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

const (
	ElipticalCurveID = ecc.BN254
)

func CreateZKP(base domain.ZkpCircuitBase) (*ZkpResult, error) {
	schemaJSON := base.SchemaJSON
	if schemaJSON == "" {
		schemaJSON = DefaultAgeSchema
	}

	schema, err := ParseSchema([]byte(schemaJSON))
	if err != nil {
		return nil, err
	}

	circuit, err := NewDynamicCircuit(schema)
	if err != nil {
		return nil, err
	}

	// 1. Compile the circuit (constraint system)
	ccs, err := frontend.Compile(
		ElipticalCurveID.ScalarField(),
		r1cs.NewBuilder,
		circuit,
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
	assignment := circuit.Clone()

	valueMap := make(map[string]interface{}, len(base.VerifiedValues))
	for _, field := range base.VerifiedValues {
		valueMap[field.Key] = field.Value
	}

	if len(valueMap) == 0 {
		var defaults map[string]interface{}
		if err := json.Unmarshal([]byte(`{"birth_year":1990,"birth_month":10,"birth_day":18}`), &defaults); err != nil {
			return nil, fmt.Errorf("load default assignment: %w", err)
		}
		valueMap = defaults
	}

	if err := assignment.AssignValues(valueMap); err != nil {
		return nil, err
	}

	fullWitness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
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
