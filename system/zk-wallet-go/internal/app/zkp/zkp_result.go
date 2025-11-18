// zkp_result.go
package zkp

import (
	"bytes"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"
	"github.com/near/borsh-go"
)

type ZkpResult struct {
	Proof         groth16.Proof
	PublicWitness witness.Witness
	TxHash        string `borsh_skip:"true"`
}

type intermediateSerializationStep struct {
	Proof         []byte `borsh:"proof"`
	PublicWitness []byte `borsh:"public_witness"`
}

func (zr *ZkpResult) SerializeBorsh() ([]byte, error) {
	var proofBuf bytes.Buffer
	_, err := zr.Proof.WriteTo(&proofBuf)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	var witnessBuf bytes.Buffer
	_, err = zr.PublicWitness.WriteTo(&witnessBuf)
	if err != nil {
		return nil, err
	}

	zkpSerializable := intermediateSerializationStep{
		Proof:         proofBuf.Bytes(),
		PublicWitness: witnessBuf.Bytes(),
	}

	return borsh.Serialize(zkpSerializable)
}

// proof reconstruction
func ReconstructZkpResult(serializedZkp []byte) (*ZkpResult, error) {
	var deserialized intermediateSerializationStep
	err := borsh.Deserialize(&deserialized, serializedZkp)
	if err != nil {
		return nil, err
	}

	proof := groth16.NewProof(ElipticalCurveID)
	_, err = proof.ReadFrom(bytes.NewReader(deserialized.Proof))
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	witness, err := witness.New(ElipticalCurveID.ScalarField())
	_, err = witness.ReadFrom(bytes.NewReader(deserialized.PublicWitness))
	if err != nil {
		return nil, err
	}

	return &ZkpResult{
		Proof:         proof,
		PublicWitness: witness,
	}, nil
}
