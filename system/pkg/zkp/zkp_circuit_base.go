package zkp

type ZkpCircuitBase struct {
	SchemaJSON     string
	VerifiedValues []ZkpField[any]
}
