package zkp

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type FieldType string

const (
	FieldTypeInteger FieldType = "integer"
	FieldTypeNumber  FieldType = "number"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeString  FieldType = "string"
	FieldTypeDate    FieldType = "date"
)

type ConstraintType string

const (
	ConstraintRange      ConstraintType = "range_check"
	ConstraintComparison ConstraintType = "comparison"
	ConstraintAge        ConstraintType = "age_verification"
)

type SchemaDefinition struct {
	SchemaID    string                 `json:"schema_id"`
	Version     string                 `json:"version"`
	Fields      []FieldDefinition      `json:"fields"`
	Constraints []ConstraintDefinition `json:"constraints"`

	fieldIndex map[string]FieldDefinition
}

type FieldDefinition struct {
	Name        string    `json:"name"`
	Type        FieldType `json:"type"`
	Required    bool      `json:"required"`
	Secret      bool      `json:"secret"`
	Public      bool      `json:"public"`
	Description string    `json:"description"`
}

type ConstraintDefinition struct {
	Type         ConstraintType  `json:"type"`
	Fields       []string        `json:"fields"`
	Operator     string          `json:"operator"`
	Value        json.RawMessage `json:"value"`
	ErrorMessage string          `json:"error_message"`
}

func ParseSchema(data []byte) (*SchemaDefinition, error) {
	var schema SchemaDefinition
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}

	if err := schema.prepare(); err != nil {
		return nil, err
	}

	return &schema, nil
}

func (s *SchemaDefinition) prepare() error {
	if len(s.Fields) == 0 {
		return errors.New("schema must declare at least one field")
	}

	s.fieldIndex = make(map[string]FieldDefinition, len(s.Fields))
	for idx, field := range s.Fields {
		if field.Name == "" {
			return errors.New("schema field name cannot be empty")
		}
		if _, exists := s.fieldIndex[field.Name]; exists {
			return fmt.Errorf("duplicate field '%s' in schema", field.Name)
		}

		// Default visibility: secret if neither explicitly public nor secret.
		if field.Public {
			field.Secret = false
		} else if !field.Secret {
			field.Secret = true
		}

		if field.Type == "" {
			field.Type = FieldTypeString
		}

		s.fieldIndex[field.Name] = field
		s.Fields[idx] = field
	}

	for _, constraint := range s.Constraints {
		if constraint.Type == "" {
			return fmt.Errorf("constraint must declare type")
		}
		if len(constraint.Fields) == 0 {
			return fmt.Errorf("constraint '%s' must reference at least one field", constraint.Type)
		}
		for _, fieldName := range constraint.Fields {
			if _, ok := s.fieldIndex[fieldName]; !ok {
				return fmt.Errorf("constraint references unknown field '%s'", fieldName)
			}
		}
	}

	return nil
}

func (s *SchemaDefinition) field(name string) (FieldDefinition, bool) {
	fd, ok := s.fieldIndex[name]
	return fd, ok
}

func (s *SchemaDefinition) SecretFieldOrder() []string {
	order := make([]string, 0, len(s.Fields))
	for _, field := range s.Fields {
		if field.Public {
			continue
		}
		order = append(order, field.Name)
	}
	return order
}

func (s *SchemaDefinition) PublicFieldOrder() []string {
	order := make([]string, 0, len(s.Fields))
	for _, field := range s.Fields {
		if !field.Public {
			continue
		}
		order = append(order, field.Name)
	}
	return order
}

func (c ConstraintDefinition) ValueAsInt() (int64, error) {
	if len(c.Value) == 0 {
		return 0, errors.New("constraint missing value")
	}

	var number json.Number
	if err := json.Unmarshal(c.Value, &number); err == nil {
		if v, err := number.Int64(); err == nil {
			return v, nil
		}
		if f, err := number.Float64(); err == nil {
			return int64(f), nil
		}
	}

	var v float64
	if err := json.Unmarshal(c.Value, &v); err == nil {
		return int64(v), nil
	}

	var s string
	if err := json.Unmarshal(c.Value, &s); err == nil {
		return parseStringInt(s)
	}

	return 0, fmt.Errorf("constraint value is not numeric: %s", string(c.Value))
}

func (c ConstraintDefinition) ValueAsNumberSlice() ([]float64, error) {
	if len(c.Value) == 0 {
		return nil, errors.New("constraint missing value array")
	}
	var values []float64
	if err := json.Unmarshal(c.Value, &values); err != nil {
		return nil, fmt.Errorf("constraint expects numeric array value: %w", err)
	}
	return values, nil
}

func parseStringInt(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, errors.New("empty numeric string")
	}
	var number json.Number = json.Number(value)
	if v, err := number.Int64(); err == nil {
		return v, nil
	}
	if f, err := number.Float64(); err == nil {
		return int64(f), nil
	}
	return 0, fmt.Errorf("invalid numeric value '%s'", value)
}

func (s *SchemaDefinition) FieldDefinition(name string) (FieldDefinition, error) {
	field, ok := s.field(name)
	if !ok {
		return FieldDefinition{}, fmt.Errorf("unknown field '%s'", name)
	}
	return field, nil
}
