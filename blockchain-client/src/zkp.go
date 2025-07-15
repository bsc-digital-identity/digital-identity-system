package main

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

type VerifiedIdentity struct {
	Components []IdentityComponent
}

type IdentityComponent struct {
	Key   string
	Type  string
	Value any
}

type IdentityCircuit struct {
	Inputs map[string]frontend.Variable
	Schema []FieldConstraint
}

type FieldConstraint struct {
	Key        string
	Type       string
	Constraint string
	Value      interface{}
}

func (circuit *IdentityCircuit) Define(api frontend.API) error {
	for _, schema := range circuit.Schema {
		value := circuit.Inputs[schema.Key]

		var err error
		switch schema.Type {
		case "int":
			err = ValidateCircuitValue(value, schema, api)
		case "float":
			err = ValidateCircuitValue(value, schema, api)
		case "string":
			err = ValidateCircuitValue(value, schema, api)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func ValidateCircuitValue[T comparable](value T, schema FieldConstraint, api frontend.API) error {
	switch schema.Type {
	case "le":
		api.AssertIsLessOrEqual(value, api.Compiler().ConstantValue(schema.Value.(T)))
	case "eq":
		api.AssertIsEqual(value, schema.Value.(T))
	case "not":
		api.AssertIsDifferent(value, schema.Value.(T))
	}
	return nil
}

func CreateZKP(circuit *IdentityCircuit) {
	ccs, _ := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, circuit)

	pk, vk, _ := groth16.Setup(ccs)

}
