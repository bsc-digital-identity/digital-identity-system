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
	currentMonth := int(currentTime.Month())
	currentDay := currentTime.Day()

	minValidYear := currentYear - 18
	minValidYearVar := frontend.Variable(minValidYear)
	currentMonthVar := frontend.Variable(currentMonth)
	currentDayVar := frontend.Variable(currentDay)

	api.AssertIsLessOrEqual(circuit.AgeYear, minValidYearVar)

	yearIsMinValid := api.IsZero(api.Sub(circuit.AgeYear, minValidYearVar))

	api.AssertIsLessOrEqual(1, circuit.AgeMonth)
	api.AssertIsLessOrEqual(circuit.AgeMonth, 12)

	api.AssertIsLessOrEqual(circuit.AgeMonth, api.Select(yearIsMinValid, currentMonthVar, 12))

	monthIsCurrent := api.IsZero(api.Sub(circuit.AgeMonth, currentMonthVar))

	api.AssertIsLessOrEqual(1, circuit.AgeDay)
	api.AssertIsLessOrEqual(circuit.AgeDay, 31)

	api.AssertIsLessOrEqual(
		circuit.AgeDay,
		api.Select(
			api.And(yearIsMinValid, monthIsCurrent),
			currentDayVar,
			31,
		),
	)

	return nil
}
