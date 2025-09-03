package model

type VerifiedSchema struct {
	Id              int    `gorm:"primaryKey;autoIncrement"`
	SchemaId        string `gorm:"uniqueIndex"`
	SuperIdentityId int    // actual foreign key
	//SuperIdentity   identity.Identity `gorm:"foreignKey:SuperIdentityId;references:Id"`
	Schema string // json format as string
}

type Schema struct {
	Constraints []Constraint `json:"constraints"`
}

type Constraint struct {
	Key        string         `json:"key"`
	Comparison ComparisonType `json:"comparison_type"`
	Value      any            `json:"value"`
}

type ComparisonType string

const (
	LessEquals  ComparisonType = "le"
	GreaterThan ComparisonType = "gt"
	Equls       ComparisonType = "eq"
	NotEquals   ComparisonType = "not"
)
