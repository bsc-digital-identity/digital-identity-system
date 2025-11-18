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

// hashStringToFieldElement bierze string i zwraca *big.Int z SHA-256(s) jako big-endian.
func hashStringToFieldElement(s string) *big.Int {
	h := sha256.Sum256([]byte(s))
	return new(big.Int).SetBytes(h[:])
}

// ProveDynamicFromSchema takes a raw JSON schema and a set of field assignments,
// builds a DynamicCircuit, and returns a ZkpResult (proof + VK + public witness).
//
// assignments is a map: fieldName -> value (int/float/string/bool/etc.)
// The convertToVariable + AssignValues logic will do type conversions.
func ProveDynamicFromSchema(schemaJSON []byte, assignments map[string]interface{}, pkBytes []byte) (*ZkpResult, error) {
	if len(schemaJSON) == 0 {
		return nil, fmt.Errorf("empty schema JSON")
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
		fieldDef, _ := schema.FieldDefinition(name)
		if fieldDef.Type == FieldTypeString {
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

	// 6) Load PK from server
	pk := groth16.NewProvingKey(ElipticalCurveID)
	_, err = pk.ReadFrom(bytes.NewReader(pkBytes))
	if err != nil {
		return nil, fmt.Errorf("read pk: %w", err)
	}

	// 7) Generate proof using server PK
	proof, err := groth16.Prove(ccs, pk, fullWitness)
	if err != nil {
		return nil, fmt.Errorf("groth16 prove: %w", err)
	}

	return &ZkpResult{
		Proof:         proof,
		PublicWitness: publicWitness,
	}, nil
}
