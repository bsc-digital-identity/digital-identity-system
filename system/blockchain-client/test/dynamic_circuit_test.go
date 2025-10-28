package test

import (
	"blockchain-client/src/zkp"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

func TestDynamicCircuitCompileDefaultSchema(t *testing.T) {
	schema, err := zkp.ParseSchema([]byte(zkp.DefaultAgeSchema))
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	circuit, err := zkp.NewDynamicCircuit(schema)
	if err != nil {
		t.Fatalf("failed to create circuit: %v", err)
	}

	_, err = frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, circuit)
	if err != nil {
		t.Fatalf("failed to compile circuit: %v", err)
	}

	assignment := circuit.Clone()
	values := map[string]interface{}{
		"birth_year":  1990,
		"birth_month": 5,
		"birth_day":   9,
	}

	if err := assignment.AssignValues(values); err != nil {
		t.Fatalf("failed to assign values: %v", err)
	}

	witness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatalf("failed to build witness: %v", err)
	}

	if _, err := witness.Public(); err != nil {
		t.Fatalf("failed to extract public witness: %v", err)
	}
}

func TestDynamicCircuitAssignMissingRequired(t *testing.T) {
	schema, err := zkp.ParseSchema([]byte(zkp.DefaultAgeSchema))
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	circuit, err := zkp.NewDynamicCircuit(schema)
	if err != nil {
		t.Fatalf("failed to create circuit: %v", err)
	}

	assignment := circuit.Clone()
	values := map[string]interface{}{
		"birth_year": 1990,
	}

	if err := assignment.AssignValues(values); err == nil {
		t.Fatalf("expected error when assigning incomplete values")
	}
}
