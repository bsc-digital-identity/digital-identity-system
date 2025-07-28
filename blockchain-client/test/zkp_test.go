package test

import (
	zkp "blockchain-client/src/zkp"
	"bytes"
	"testing"
	"time"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"
	"github.com/near/borsh-go"
)

type ZkpTestingParams struct {
	name         string
	day          int
	month        int
	year         int
	shouldVerify bool
}

type ZkpResultBinary struct {
	Proof         []byte
	VerifyingKey  []byte
	PublicWitness []byte
	TxHash        string `borsh_skip:"true"`
}

func TestZkpShouldProveRightOver18(t *testing.T) {
	tt := ZkpTestingParams{"Over 18", 15, 7, 1990, true}

	zkpRes, err := zkp.CreateZKP(tt.day, tt.month, tt.year)
	if err != nil && tt.shouldVerify {
		t.Fatalf("Failed to create ZKP: %v", err)
	}

	if !tt.shouldVerify {
		t.Log("Function that should not verify passed")
		return
	}

	err = groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness)
	if tt.shouldVerify && err != nil {
		t.Errorf("Expected verification to pass but got error: %v", err)
	}
	if !tt.shouldVerify && err == nil {
		t.Error("Expected verification to fail but it passed")
	}
}

func TestZkpShouldProveRightExactly18(t *testing.T) {
	tt := ZkpTestingParams{"Exactly 18", time.Now().Day(), int(time.Now().Month()), time.Now().Year() - 18, true}

	zkpRes, err := zkp.CreateZKP(tt.day, tt.month, tt.year)
	if err != nil && tt.shouldVerify {
		t.Fatalf("Failed to create ZKP: %v", err)
	}

	if !tt.shouldVerify {
		t.Log("Function that should not verify passed")
		return
	}

	err = groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness)
	if tt.shouldVerify && err != nil {
		t.Errorf("Expected verification to pass but got error: %v", err)
	}
	if !tt.shouldVerify && err == nil {
		t.Error("Expected verification to fail but it passed")
	}

}

func TestZkpShouldProveWrongUnder18(t *testing.T) {
	tt := ZkpTestingParams{"Under 18", 15, 7, time.Now().Year() - 10, false}

	zkpRes, err := zkp.CreateZKP(tt.day, tt.month, tt.year)
	if err != nil && tt.shouldVerify {
		t.Fatalf("Failed to create ZKP: %v", err)
	}

	if !tt.shouldVerify {
		t.Log("Function that should not verify passed")
		return
	}

	err = groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness)
	if tt.shouldVerify && err != nil {
		t.Errorf("Expected verification to pass but got error: %v", err)
	}
	if !tt.shouldVerify && err == nil {
		t.Error("Expected verification to fail but it passed")
	}
}

func TestSerializationForZKP(t *testing.T) {
	tt := ZkpTestingParams{"Over 18", 15, 7, 1990, true}

	// create and serialize ZKP to solana compatible format
	zkpRes, err := zkp.CreateZKP(tt.day, tt.month, tt.year)
	if err != nil && tt.shouldVerify {
		t.Fatalf("Failed to create ZKP: %v", err)
	}

	var proofBuf bytes.Buffer
	_, err = zkpRes.Proof.WriteTo(&proofBuf)
	if err != nil {
		t.Fatalf("Proof serialization failed: %v", err)
	}

	var vkBuf bytes.Buffer
	_, err = zkpRes.VerifyingKey.WriteTo(&vkBuf)
	if err != nil {
		t.Fatalf("VerifyingKey serialization failed: %v", err)
	}

	var witnessBuf bytes.Buffer
	_, err = zkpRes.PublicWitness.WriteTo(&witnessBuf)
	if err != nil {
		t.Fatalf("Witness serialization failed: %v", err)
	}

	zkpSerializable := struct {
		Proof         []byte
		VerifyingKey  []byte
		PublicWitness []byte
	}{
		Proof:         proofBuf.Bytes(),
		VerifyingKey:  vkBuf.Bytes(),
		PublicWitness: witnessBuf.Bytes(),
	}

	serializedZkp, err := borsh.Serialize(zkpSerializable)
	if err != nil {
		t.Fatalf("Borsh serialization failed: %v", err)
	}

	// deserialize the data
	var deserialized struct {
		Proof         []byte
		VerifyingKey  []byte
		PublicWitness []byte
	}
	err = borsh.Deserialize(&deserialized, serializedZkp)
	if err != nil {
		t.Fatalf("Borsh deserialization failed: %v", err)
	}

	// proof deconstruction and verification
	proof := groth16.NewProof(zkp.ElipticalCurveID)
	_, err = proof.ReadFrom(bytes.NewReader(deserialized.Proof))
	if err != nil {
		t.Fatalf("Proof deserialization failed: %v", err)
	}

	vk := groth16.NewVerifyingKey(zkp.ElipticalCurveID)
	_, err = vk.ReadFrom(bytes.NewReader(deserialized.VerifyingKey))
	if err != nil {
		t.Fatalf("VerifyingKey deserialization failed: %v", err)
	}

	witness, err := witness.New(zkp.ElipticalCurveID.ScalarField())
	_, err = witness.ReadFrom(bytes.NewReader(deserialized.PublicWitness))
	if err != nil {
		t.Fatalf("Witness deserialization failed: %v", err)
	}

	err = groth16.Verify(proof, vk, witness)
	if tt.shouldVerify && err != nil {
		t.Errorf("Expected verification to pass but got error: %v", err)
	}
	if !tt.shouldVerify && err == nil {
		t.Error("Expected verification to fail but it passed")
	}
}
