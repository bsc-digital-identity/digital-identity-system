// prove_dynamic.go
package zkp

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

func hashStringToFieldElement(s string) *big.Int {
	h := sha256.Sum256([]byte(s))
	return new(big.Int).SetBytes(h[:])
}

// PK przychodzi z DI (server-held PK)
func ProveDynamicFromSchema(schemaJSON []byte, assignments map[string]interface{}, pkBytes []byte) (*ZkpResult, error) {
	if len(schemaJSON) == 0 {
		return nil, fmt.Errorf("empty schema JSON")
	}
	if len(pkBytes) == 0 {
		return nil, fmt.Errorf("empty proving key bytes")
	}

	// 1) Parse schema
	schema, err := ParseSchema(schemaJSON)
	if err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}

	// 2) Build circuit prototype
	circuit, err := NewDynamicCircuit(schema)
	if err != nil {
		return nil, fmt.Errorf("new dynamic circuit: %w", err)
	}

	// 3) Hash string inputs
	hashedAssignments := make(map[string]interface{}, len(assignments))
	for name, val := range assignments {
		fieldDef, err := schema.FieldDefinition(name)
		if err == nil && fieldDef.Type == FieldTypeString {
			s := fmt.Sprint(val)
			hashedAssignments[name] = hashStringToFieldElement(s)
		} else {
			hashedAssignments[name] = val
		}
	}

	// 4) Compile circuit
	ccs, err := frontend.Compile(
		ElipticalCurveID.ScalarField(),
		r1cs.NewBuilder,
		circuit,
	)
	if err != nil {
		return nil, fmt.Errorf("compile circuit: %w", err)
	}

	// 5) Fill witness
	witnessCircuit := circuit.Clone()
	if err := witnessCircuit.AssignValues(hashedAssignments); err != nil {
		return nil, fmt.Errorf("assign values: %w", err)
	}

	fullWitness, err := frontend.NewWitness(
		witnessCircuit,
		ElipticalCurveID.ScalarField(),
	)
	if err != nil {
		return nil, fmt.Errorf("new witness: %w", err)
	}

	publicWitness, err := fullWitness.Public()
	if err != nil {
		return nil, fmt.Errorf("public witness: %w", err)
	}

	// 6) Load PK from DI
	pk := groth16.NewProvingKey(ElipticalCurveID)
	if _, err := pk.ReadFrom(bytes.NewReader(pkBytes)); err != nil {
		return nil, fmt.Errorf("read pk: %w", err)
	}

	// 7) Generate proof using DI's PK
	proof, err := groth16.Prove(ccs, pk, fullWitness)
	if err != nil {
		return nil, fmt.Errorf("groth16 prove: %w", err)
	}

	return &ZkpResult{
		Proof:         proof,
		PublicWitness: publicWitness,
	}, nil
}
