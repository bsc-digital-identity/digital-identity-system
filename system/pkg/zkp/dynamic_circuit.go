package zkp

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/consensys/gnark/frontend"
)

// ------------------------------------------------------------------

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

func hashStringToFieldElement(s string) *big.Int {
	h := sha256.Sum256([]byte(s))
	return new(big.Int).SetBytes(h[:])
}

func (dc *DynamicCircuit) AssignValues(values map[string]interface{}) error {
	assigned := make(map[string]struct{}, len(values))

	for name, rawValue := range values {
		field, ok := dc.fieldMetadata[name]
		if !ok {
			return fmt.Errorf("assignment references unknown field '%s'", name)
		}

		var variable frontend.Variable

		// POPRAWKA: Bardziej elastyczne sprawdzanie czy pole jest stringiem.
		// Rzutujemy field.Type na string, aby strings.EqualFold zadziałało poprawnie.
		isString := strings.EqualFold(string(field.Type), "string") || field.Type == FieldTypeString

		if isString {
			// CRITICAL FIX: Check if the input is ACTUALLY a string before hashing.
			// If the upstream service passed a number (or BigInt) for a string field,
			// treating it as a string ("12345...") and hashing it results in Double Hashing.
			if strVal, ok := rawValue.(string); ok {
				fmt.Printf("DEBUG: AssignValues hashing string field '%s': '%s'\n", name, strVal)
				variable = hashStringToFieldElement(strVal)
			} else {
				// Fallback: The input is not a string (e.g. it's a big.Int or number).
				// Assume it is already a field element/hash provided by the caller.
				fmt.Printf("DEBUG: AssignValues received non-string for string field '%s', treating as number: %v\n", name, rawValue)
				var err error
				variable, err = convertToVariable(field, rawValue)
				if err != nil {
					return fmt.Errorf("invalid numeric value for string field '%s': %w", name, err)
				}
			}
		} else {
			var err error
			variable, err = convertToVariable(field, rawValue)
			if err != nil {
				return fmt.Errorf("invalid value for field '%s': %w", name, err)
			}
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

	minBound, err := toIntBound(bounds[0])
	if err != nil {
		return err
	}
	maxBound, err := toIntBound(bounds[1])
	if err != nil {
		return err
	}

	value, err := dc.fieldVariable(constraint.Fields[0])
	if err != nil {
		return err
	}

	api.AssertIsLessOrEqual(minBound, value)
	api.AssertIsLessOrEqual(value, maxBound)

	return nil
}

func (dc *DynamicCircuit) applyComparisonConstraint(api frontend.API, constraint ConstraintDefinition) error {
	// DEBUG LOG: Pozwala upewnić się, że Constraint widzi wartość "AGH"
	fmt.Printf("DEBUG: Sprawdzam constraint. Value raw: %s\n", constraint.Value)

	if len(constraint.Fields) == 0 {
		return fmt.Errorf("comparison constraint must declare at least one field")
	}

	left, err := dc.fieldVariable(constraint.Fields[0])
	if err != nil {
		return err
	}

	var right frontend.Variable
	// Check if we are comparing two fields or one field against a constant value
	if len(constraint.Fields) > 1 {
		// Case 1: Comparing two circuit variables (e.g., fieldA == fieldB)
		right, err = dc.fieldVariable(constraint.Fields[1])
		if err != nil {
			return err
		}
	} else {
		// Case 2: Comparing a circuit variable against a constant value from the schema
		// The value is stored as json.RawMessage.

		var strVal string
		// Attempt to unmarshal the raw JSON value as a string.
		// This ensures that "AGH" (string) is treated differently than 123 (number).
		if err := json.Unmarshal(constraint.Value, &strVal); err == nil {
			// SUKCES: To jest string. Haszujemy go.
			// To musi pasować do logiki w AssignValues!
			right = hashStringToFieldElement(strVal)
		} else {
			// FALLBACK: To nie jest string, więc zakładamy liczbę.
			number, err := constraint.ValueAsInt()
			if err != nil {
				return fmt.Errorf("constraint value is neither a valid string nor a number: %w", err)
			}
			right = number
		}
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
	if len(constraint.Fields) != 1 {
		return fmt.Errorf("age constraint expects one field (birthdate timestamp)")
	}

	minAgeYears, err := constraint.ValueAsInt()
	if err != nil {
		return err
	}

	const secondsPerYear = int64(365 * 24 * 60 * 60)

	birthTsVar, err := dc.fieldVariable(constraint.Fields[0])
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	minBirthTs := now - int64(minAgeYears)*secondsPerYear

	api.AssertIsLessOrEqual(birthTsVar, frontend.Variable(minBirthTs))

	return nil
}

func toIntBound(bound float64) (int64, error) {
	if math.Trunc(bound) != bound {
		return 0, fmt.Errorf("range constraint bound must be whole number, got %v", bound)
	}
	return int64(bound), nil
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
