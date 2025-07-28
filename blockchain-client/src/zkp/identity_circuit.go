package zkp

import (
	"time"

	"github.com/consensys/gnark/frontend"
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
