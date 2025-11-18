package zkprequest

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"pkg-common/zkp"
)

func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h[:])
}

func (s *Service) ensureVKForSchema(canon string) (schemaHash string, err error) {
	hash := "sha256:" + sha256Hex([]byte(canon))
	if _, ok := s.vkCache[hash]; ok {
		return hash, nil
	}
	// compile circuit
	schema, err := zkp.ParseSchema([]byte(canon))
	if err != nil {
		return "", err
	}
	circ, err := zkp.NewDynamicCircuit(schema)
	if err != nil {
		return "", err
	}
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, circ)
	if err != nil {
		return "", err
	}

	// setup and cache VK
	_, vk, err := groth16.Setup(ccs)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if _, err := vk.WriteTo(&buf); err != nil {
		return "", err
	}
	s.vkCache[hash] = buf.Bytes()
	return hash, nil
}
