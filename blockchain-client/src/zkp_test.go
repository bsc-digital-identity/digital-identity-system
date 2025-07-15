package main

import (
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

func TestZkpShouldProveRightOver18(t *testing.T) {
	tt := ZkpTestingParams{"Over 18", 15, 7, 1990, true}

	zkp, err := CreateZKP(tt.day, tt.month, tt.year)
	if err != nil && tt.shouldVerify {
		t.Fatalf("Failed to create ZKP: %v", err)
	}

	if !tt.shouldVerify {
		t.Log("Function that should not verify passed")
		return
	}

	err = groth16.Verify(zkp.Proof, zkp.VerifyingKey, zkp.PublicWitness)
	if tt.shouldVerify && err != nil {
		t.Errorf("Expected verification to pass but got error: %v", err)
	}
	if !tt.shouldVerify && err == nil {
		t.Error("Expected verification to fail but it passed")
	}
}

func TestZkpShouldProveRightExactly18(t *testing.T) {
	tt := ZkpTestingParams{"Exactly 18", time.Now().Day(), int(time.Now().Month()), time.Now().Year() - 18, true}

	zkp, err := CreateZKP(tt.day, tt.month, tt.year)
	if err != nil && tt.shouldVerify {
		t.Fatalf("Failed to create ZKP: %v", err)
	}

	if !tt.shouldVerify {
		t.Log("Function that should not verify passed")
		return
	}

	err = groth16.Verify(zkp.Proof, zkp.VerifyingKey, zkp.PublicWitness)
	if tt.shouldVerify && err != nil {
		t.Errorf("Expected verification to pass but got error: %v", err)
	}
	if !tt.shouldVerify && err == nil {
		t.Error("Expected verification to fail but it passed")
	}

}

func TestZkpShouldProveWrongUnder18(t *testing.T) {
	tt := ZkpTestingParams{"Under 18", 15, 7, time.Now().Year() - 10, false}

	zkp, err := CreateZKP(tt.day, tt.month, tt.year)
	if err != nil && tt.shouldVerify {
		t.Fatalf("Failed to create ZKP: %v", err)
	}

	if !tt.shouldVerify {
		t.Log("Function that should not verify passed")
		return
	}

	err = groth16.Verify(zkp.Proof, zkp.VerifyingKey, zkp.PublicWitness)
	if tt.shouldVerify && err != nil {
		t.Errorf("Expected verification to pass but got error: %v", err)
	}
	if !tt.shouldVerify && err == nil {
		t.Error("Expected verification to fail but it passed")
	}
}
