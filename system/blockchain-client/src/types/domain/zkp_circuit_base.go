package domain

type ZkpCircuitBase struct {
	SchemaJSON     string
	VerifiedValues []ZkpField[any]
}
