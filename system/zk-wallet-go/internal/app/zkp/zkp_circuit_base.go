// zkp_circuit_base.go
package zkp

import "github.com/consensys/gnark-crypto/ecc"

const (
	ElipticalCurveID = ecc.BN254
)

type ZkpCircuitBase struct {
	SchemaJSON     string
	VerifiedValues []ZkpField[any]
}
