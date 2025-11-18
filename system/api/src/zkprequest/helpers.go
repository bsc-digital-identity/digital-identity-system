package zkprequest

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"pkg-common/zkp"
)

func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h[:])
}

// ensureVKForSchema
func (s *Service) ensureVKForSchema(canon string) (string, error) {
	hash := "sha256:" + sha256Hex([]byte(canon))

	// szybki check cache
	s.cacheMu.RLock()
	_, vkOK := s.vkCache[hash]
	_, pkOK := s.pkCache[hash]
	s.cacheMu.RUnlock()

	if vkOK && pkOK {
		return hash, nil
	}

	// 1) parsowanie schemy
	schema, err := zkp.ParseSchema([]byte(canon))
	if err != nil {
		return "", err
	}

	// 2) budowa circuitu
	circ, err := zkp.NewDynamicCircuit(schema)
	if err != nil {
		return "", err
	}

	// 3) kompilacja R1CS na *tej samej* krzywej co wallet
	ccs, err := frontend.Compile(
		zkp.ElipticalCurveID.ScalarField(),
		r1cs.NewBuilder,
		circ,
	)
	if err != nil {
		return "", err
	}

	// 4) Setup Groth16
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		return "", err
	}

	var vkBuf, pkBuf bytes.Buffer
	if _, err := vk.WriteTo(&vkBuf); err != nil {
		return "", err
	}
	if _, err := pk.WriteTo(&pkBuf); err != nil {
		return "", err
	}

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	if s.vkCache == nil {
		s.vkCache = make(map[string][]byte)
	}
	if s.pkCache == nil {
		s.pkCache = make(map[string][]byte)
	}
	if s.schemaCache == nil {
		s.schemaCache = make(map[string][]byte)
	}

	s.vkCache[hash] = vkBuf.Bytes()
	s.pkCache[hash] = pkBuf.Bytes()
	s.schemaCache[hash] = []byte(canon)

	return hash, nil
}
