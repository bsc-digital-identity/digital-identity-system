package authschemas

import "api/src/identity"

type VerifiableSchema struct {
	Id              int                    `gorm:"primaryKey;autoIncrement"`
	SchemaId        string                 `gorm:"uniqueIndex"`
	SuperIdentityId int                    // actual foreign key
	SuperIdentity   identity.SuperIdentity `gorm:"foreignKey:SuperIdentityId;references:Id"`
	Schema          string                 // json format as string
}

type Schema[T comparable] struct {
	Constraints []Constraint[T] `json:"constraints"`
}

type Constraint[T comparable] struct {
	Key        string         `json:"key"`
	Comparison ComparisonType `json:"comparison_type"`
	Value      T              `json:"value"`
}

type ComparisonType string

const (
	LessEquals  ComparisonType = "le"
	GreaterThan ComparisonType = "gt"
	Equls       ComparisonType = "eq"
	NotEquals   ComparisonType = "not"
)
