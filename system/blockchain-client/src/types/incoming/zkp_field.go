package incoming

type supportedType string

const (
	stringType supportedType = "string"
	intType    supportedType = "int"
	floatType  supportedType = "float"
	dateType   supportedType = "date"
	boolType   supportedType = "bool"
)

type ZkpFieldDto struct {
	Key                  string `json:"key"`
	Value                string `json:"value"`
	Type                 string `json:"type"`
	VerificationPositive bool   `json:"verification_positive"`
}
