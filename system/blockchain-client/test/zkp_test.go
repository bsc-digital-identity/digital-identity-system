package test

import (
	"blockchain-client/src/types/domain"
	"blockchain-client/src/zkp"
	"testing"
	"time"

	"github.com/consensys/gnark/backend/groth16"
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

func (z ZkpTestingParams) toCircuitBase() domain.ZkpCircuitBase {
	return domain.ZkpCircuitBase{
		VerifiedValues: []domain.ZkpField[any]{
			{Key: "birth_day", Value: z.day},
			{Key: "birth_month", Value: z.month},
			{Key: "birth_year", Value: z.year},
		},
	}
}

func TestZkpShouldProveRightOver18(t *testing.T) {
	tt := ZkpTestingParams{"Over 18", 15, 7, 1990, true}

	zkpRes, err := zkp.CreateZKP(tt.toCircuitBase())
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

	zkpRes, err := zkp.CreateZKP(tt.toCircuitBase())
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

	zkpRes, err := zkp.CreateZKP(tt.toCircuitBase())
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
	zkpRes, err := zkp.CreateZKP(tt.toCircuitBase())
	if err != nil && tt.shouldVerify {
		t.Fatalf("Failed to create ZKP: %v", err)
	}

	serialized, err := zkpRes.SerializeBorsh()
	if err != nil {
		t.Fatalf("Proof serialization failed: %v", err)
	}

	reconstructedZkp, err := zkp.ReconstructZkpResult(serialized)
	if err != nil {
		t.Fatalf("ZKP recoonstruction failed: %v", err)
	}

	err = groth16.Verify(reconstructedZkp.Proof, reconstructedZkp.VerifyingKey, reconstructedZkp.PublicWitness)
	if tt.shouldVerify && err != nil {
		t.Errorf("Expected verification to pass but got error: %v", err)
	}
	if !tt.shouldVerify && err == nil {
		t.Error("Expected verification to fail but it passed")
	}
}
