package zkp

import (
	"fmt"
	"time"

	"github.com/consensys/gnark/frontend"
)

type DynamicCircuit struct {
	SecretValues []frontend.Variable `gnark:",secret"`
	PublicValues []frontend.Variable `gnark:",public"`

	Schema        *SchemaDefinition `gnark:"-"`
	secretOrder   []string          `gnark:"-"`
	publicOrder   []string          `gnark:"-"`
	secretIndex   map[string]int    `gnark:"-"`
	publicIndex   map[string]int    `gnark:"-"`
	fieldMetadata map[string]FieldDefinition
}

func NewDynamicCircuit(schema *SchemaDefinition) (*DynamicCircuit, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema cannot be nil")
	}

	secretOrder := schema.SecretFieldOrder()
	publicOrder := schema.PublicFieldOrder()

	circuit := &DynamicCircuit{
		SecretValues:  make([]frontend.Variable, len(secretOrder)),
		PublicValues:  make([]frontend.Variable, len(publicOrder)),
		Schema:        schema,
		secretOrder:   append([]string(nil), secretOrder...),
		publicOrder:   append([]string(nil), publicOrder...),
		secretIndex:   make(map[string]int, len(secretOrder)),
		publicIndex:   make(map[string]int, len(publicOrder)),
		fieldMetadata: make(map[string]FieldDefinition, len(schema.Fields)),
	}

	for i, name := range secretOrder {
		circuit.secretIndex[name] = i
	}
	for i, name := range publicOrder {
		circuit.publicIndex[name] = i
	}
	for _, field := range schema.Fields {
		circuit.fieldMetadata[field.Name] = field
	}

	return circuit, nil
}

func (dc *DynamicCircuit) Clone() *DynamicCircuit {
	clone := &DynamicCircuit{
		SecretValues:  make([]frontend.Variable, len(dc.SecretValues)),
		PublicValues:  make([]frontend.Variable, len(dc.PublicValues)),
		Schema:        dc.Schema,
		secretOrder:   append([]string(nil), dc.secretOrder...),
		publicOrder:   append([]string(nil), dc.publicOrder...),
		secretIndex:   dc.secretIndex,
		publicIndex:   dc.publicIndex,
		fieldMetadata: dc.fieldMetadata,
	}
	return clone
}

func (dc *DynamicCircuit) AssignValues(values map[string]interface{}) error {
	assigned := make(map[string]struct{}, len(values))

	for name, rawValue := range values {
		field, ok := dc.fieldMetadata[name]
		if !ok {
			return fmt.Errorf("assignment references unknown field '%s'", name)
		}

		variable, err := convertToVariable(field, rawValue)
		if err != nil {
			return fmt.Errorf("invalid value for field '%s': %w", name, err)
		}

		if idx, ok := dc.secretIndex[name]; ok {
			dc.SecretValues[idx] = variable
		} else if idx, ok := dc.publicIndex[name]; ok {
			dc.PublicValues[idx] = variable
		} else {
			return fmt.Errorf("field '%s' does not map to a circuit input", name)
		}

		assigned[name] = struct{}{}
	}

	for _, field := range dc.Schema.Fields {
		if field.Required {
			if _, ok := assigned[field.Name]; !ok {
				return fmt.Errorf("required field '%s' missing from assignments", field.Name)
			}
		}
	}

	return nil
}

func (dc *DynamicCircuit) Define(api frontend.API) error {
	for _, constraint := range dc.Schema.Constraints {
		if err := dc.applyConstraint(api, constraint); err != nil {
			return err
		}
	}
	return nil
}

func (dc *DynamicCircuit) applyConstraint(api frontend.API, constraint ConstraintDefinition) error {
	switch constraint.Type {
	case ConstraintRange:
		return dc.applyRangeConstraint(api, constraint)
	case ConstraintComparison:
		return dc.applyComparisonConstraint(api, constraint)
	case ConstraintAge:
		return dc.applyAgeConstraint(api, constraint)
	default:
		return fmt.Errorf("unsupported constraint type '%s'", constraint.Type)
	}
}

func (dc *DynamicCircuit) applyRangeConstraint(api frontend.API, constraint ConstraintDefinition) error {
	if len(constraint.Fields) != 1 {
		return fmt.Errorf("range constraint requires exactly one field")
	}
	bounds, err := constraint.ValueAsNumberSlice()
	if err != nil {
		return err
	}
	if len(bounds) != 2 {
		return fmt.Errorf("range constraint requires two bounds")
	}

	value, err := dc.fieldVariable(constraint.Fields[0])
	if err != nil {
		return err
	}

	api.AssertIsLessOrEqual(bounds[0], value)
	api.AssertIsLessOrEqual(value, bounds[1])

	return nil
}

func (dc *DynamicCircuit) applyComparisonConstraint(api frontend.API, constraint ConstraintDefinition) error {
	if len(constraint.Fields) == 0 {
		return fmt.Errorf("comparison constraint must declare at least one field")
	}

	left, err := dc.fieldVariable(constraint.Fields[0])
	if err != nil {
		return err
	}

	var right frontend.Variable
	if len(constraint.Fields) > 1 {
		right, err = dc.fieldVariable(constraint.Fields[1])
		if err != nil {
			return err
		}
	} else {
		number, err := constraint.ValueAsInt()
		if err != nil {
			return err
		}
		right = number
	}

	switch constraint.Operator {
	case "greater_equal", "ge":
		api.AssertIsLessOrEqual(right, left)
	case "greater_than", "gt":
		api.AssertIsLessOrEqual(api.Add(right, 1), left)
	case "less_equal", "le":
		api.AssertIsLessOrEqual(left, right)
	case "less_than", "lt":
		api.AssertIsLessOrEqual(api.Add(left, 1), right)
	case "equal", "eq":
		api.AssertIsEqual(left, right)
	case "not_equal", "ne":
		api.AssertIsDifferent(left, right)
	default:
		return fmt.Errorf("unsupported comparison operator '%s'", constraint.Operator)
	}

	return nil
}

func (dc *DynamicCircuit) applyAgeConstraint(api frontend.API, constraint ConstraintDefinition) error {
	if len(constraint.Fields) != 3 {
		return fmt.Errorf("age constraint expects three fields (year, month, day)")
	}

	minAge, err := constraint.ValueAsInt()
	if err != nil {
		return err
	}

	yearVar, err := dc.fieldVariable(constraint.Fields[0])
	if err != nil {
		return err
	}
	monthVar, err := dc.fieldVariable(constraint.Fields[1])
	if err != nil {
		return err
	}
	dayVar, err := dc.fieldVariable(constraint.Fields[2])
	if err != nil {
		return err
	}

	currentTime := time.Now()
	currentYear := currentTime.Year()
	currentMonth := int(currentTime.Month())
	currentDay := currentTime.Day()

	minValidYear := currentYear - int(minAge)
	minValidYearVar := frontend.Variable(minValidYear)
	currentMonthVar := frontend.Variable(currentMonth)
	currentDayVar := frontend.Variable(currentDay)

	api.AssertIsLessOrEqual(yearVar, minValidYearVar)

	yearIsMinValid := api.IsZero(api.Sub(yearVar, minValidYearVar))

	api.AssertIsLessOrEqual(1, monthVar)
	api.AssertIsLessOrEqual(monthVar, 12)

	api.AssertIsLessOrEqual(monthVar, api.Select(yearIsMinValid, currentMonthVar, 12))

	monthIsCurrent := api.IsZero(api.Sub(monthVar, currentMonthVar))

	api.AssertIsLessOrEqual(1, dayVar)
	api.AssertIsLessOrEqual(dayVar, 31)

	api.AssertIsLessOrEqual(
		dayVar,
		api.Select(
			api.And(yearIsMinValid, monthIsCurrent),
			currentDayVar,
			31,
		),
	)

	return nil
}

func (dc *DynamicCircuit) fieldVariable(name string) (frontend.Variable, error) {
	if idx, ok := dc.secretIndex[name]; ok {
		return dc.SecretValues[idx], nil
	}
	if idx, ok := dc.publicIndex[name]; ok {
		return dc.PublicValues[idx], nil
	}
	return nil, fmt.Errorf("unknown circuit field '%s'", name)
}
