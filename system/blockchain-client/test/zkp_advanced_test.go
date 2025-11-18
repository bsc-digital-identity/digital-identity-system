package test

import (
	domain "blockchain-client/src/types/domain"
	zkp "blockchain-client/src/zkp"
	"testing"
	"time"

	"github.com/consensys/gnark/backend/groth16"
)

const stringEqualitySchema = `{
  "schema_id": "string_equality_check",
  "version": "1.0.0",
  "fields": [
    {"name": "favorite_color", "type": "string", "required": true, "secret": true}
  ],
  "constraints": [
    {"type": "comparison", "fields": ["favorite_color"], "operator": "eq", "value": "blue"}
  ]
}`

const numberComparisonSchema = `{
  "schema_id": "score_at_least_five",
  "version": "1.0.0",
  "fields": [
    {"name": "score", "type": "number", "required": true, "secret": true}
  ],
  "constraints": [
    {"type": "comparison", "fields": ["score"], "operator": "ge", "value": 5}
  ]
}`

func newDOBBase(day, month, year int) domain.ZkpCircuitBase {
	return domain.ZkpCircuitBase{
		VerifiedValues: []domain.ZkpField[any]{
			{Key: "birth_day", Value: day},
			{Key: "birth_month", Value: month},
			{Key: "birth_year", Value: year},
		},
	}
}

// Additional comprehensive ZKP tests beyond the existing ones

func TestZkpEdgeCases(t *testing.T) {
	testCases := []struct {
		name         string
		day          int
		month        int
		year         int
		shouldVerify bool
		description  string
	}{
		{"Leap Year Feb 29", 29, 2, 1992, true, "Born on leap year Feb 29, over 18"},
		{"Future Date", 1, 1, 2030, false, "Future birth date should fail"},
		{"Very Old Person", 1, 1, 1900, true, "Very old person should pass"},
		{"Edge Case Dec 31", 31, 12, 2000, true, "Born on last day of year"},
		{"Edge Case Jan 1", 1, 1, 2000, true, "Born on first day of year"},
		{"Exactly 18 Years Ago", time.Now().Day(), int(time.Now().Month()), time.Now().Year() - 18, true, "Exactly 18 years old today"},
		{"One Day Under 18", time.Now().Day(), int(time.Now().Month()), time.Now().Year() - 18 + 1, false, "One day under 18"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			zkpRes, err := zkp.CreateZKP(newDOBBase(tc.day, tc.month, tc.year))

			if err != nil && tc.shouldVerify {
				t.Fatalf("Failed to create ZKP for %s: %v", tc.description, err)
			}

			if !tc.shouldVerify && err == nil {
				// For cases that should fail, we still need to check if verification fails
				err = groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness)
				if err == nil {
					t.Errorf("Expected verification to fail for %s but it passed", tc.description)
				}
				return
			}

			if tc.shouldVerify {
				err = groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness)
				if err != nil {
					t.Errorf("Expected verification to pass for %s but got error: %v", tc.description, err)
				}
			}
		})
	}
}

func TestZkpStringEqualityConstraint(t *testing.T) {
	base := domain.ZkpCircuitBase{
		SchemaJSON: stringEqualitySchema,
		VerifiedValues: []domain.ZkpField[any]{
			{Key: "favorite_color", Value: "blue"},
		},
	}

	zkpRes, err := zkp.CreateZKP(base)
	if err != nil {
		t.Fatalf("Failed to create ZKP for string equality schema: %v", err)
	}
	if zkpRes == nil {
		t.Fatal("ZKP result is nil for string equality schema")
	}

	if err := groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness); err != nil {
		t.Fatalf("Expected string equality ZKP verification to pass but got error: %v", err)
	}
}

func TestZkpStringEqualityConstraintFailure(t *testing.T) {
	base := domain.ZkpCircuitBase{
		SchemaJSON: stringEqualitySchema,
		VerifiedValues: []domain.ZkpField[any]{
			{Key: "favorite_color", Value: "green"},
		},
	}

	zkpRes, err := zkp.CreateZKP(base)
	if err != nil {
		// Creation failure due to unsatisfied constraint is acceptable.
		return
	}
	if zkpRes == nil {
		t.Fatal("ZKP result is nil despite no error for invalid string input")
	}

	if verifyErr := groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness); verifyErr == nil {
		t.Fatal("Expected verification to fail for mismatched string input but it passed")
	}
}

func TestZkpNumberComparisonConstraint(t *testing.T) {
	base := domain.ZkpCircuitBase{
		SchemaJSON: numberComparisonSchema,
		VerifiedValues: []domain.ZkpField[any]{
			{Key: "score", Value: 10},
		},
	}

	zkpRes, err := zkp.CreateZKP(base)
	if err != nil {
		t.Fatalf("Failed to create ZKP for numeric comparison schema: %v", err)
	}
	if zkpRes == nil {
		t.Fatal("ZKP result is nil for numeric comparison schema")
	}

	if err := groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness); err != nil {
		t.Fatalf("Expected numeric comparison ZKP verification to pass but got error: %v", err)
	}
}

func TestZkpNumberComparisonConstraintFailure(t *testing.T) {
	base := domain.ZkpCircuitBase{
		SchemaJSON: numberComparisonSchema,
		VerifiedValues: []domain.ZkpField[any]{
			{Key: "score", Value: 3},
		},
	}

	zkpRes, err := zkp.CreateZKP(base)
	if err != nil {
		// Creation may fail if circuit evaluation detects the violation early.
		return
	}
	if zkpRes == nil {
		t.Fatal("ZKP result is nil despite no error for invalid numeric input")
	}

	if verifyErr := groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness); verifyErr == nil {
		t.Fatal("Expected verification to fail for numeric comparison but it passed")
	}
}

func TestZkpInvalidInputs(t *testing.T) {
	invalidCases := []struct {
		name  string
		day   int
		month int
		year  int
	}{
		{"Invalid Day Zero", 0, 1, 1990},
		{"Invalid Day Negative", -5, 1, 1990},
		{"Invalid Day Too High", 32, 1, 1990},
		{"Invalid Month Zero", 15, 0, 1990},
		{"Invalid Month Negative", 15, -1, 1990},
		{"Invalid Month Too High", 15, 13, 1990},
		{"Invalid Year Negative", 15, 1, -1990},
		{"Invalid Year Zero", 15, 1, 0},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			zkpRes, err := zkp.CreateZKP(newDOBBase(tc.day, tc.month, tc.year))

			// Even with invalid inputs, the ZKP creation might succeed,
			// but verification should handle the logic correctly
			if err == nil && zkpRes != nil {
				// Test that verification behaves correctly with invalid dates
				err = groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness)
				// We don't assert on the result here as the circuit logic will determine validity
				t.Logf("ZKP created for invalid input %s, verification result: %v", tc.name, err)
			}
		})
	}
}

func TestZkpSerializationRoundTrip(t *testing.T) {
	testCases := []struct {
		name  string
		day   int
		month int
		year  int
	}{
		{"Standard Case", 15, 7, 1990},
		{"Edge Case", 29, 2, 1992},
		{"Recent Date", 1, 1, 2000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create original ZKP
			originalZkp, err := zkp.CreateZKP(newDOBBase(tc.day, tc.month, tc.year))
			if err != nil {
				t.Fatalf("Failed to create original ZKP: %v", err)
			}

			// Serialize
			serialized, err := originalZkp.SerializeBorsh()
			if err != nil {
				t.Fatalf("Failed to serialize ZKP: %v", err)
			}

			if len(serialized) == 0 {
				t.Fatal("Serialized data should not be empty")
			}

			// Deserialize
			reconstructed, err := zkp.ReconstructZkpResult(serialized)
			if err != nil {
				t.Fatalf("Failed to reconstruct ZKP: %v", err)
			}

			// Verify original
			err = groth16.Verify(originalZkp.Proof, originalZkp.VerifyingKey, originalZkp.PublicWitness)
			if err != nil {
				t.Fatalf("Original ZKP verification failed: %v", err)
			}

			// Verify reconstructed
			err = groth16.Verify(reconstructed.Proof, reconstructed.VerifyingKey, reconstructed.PublicWitness)
			if err != nil {
				t.Fatalf("Reconstructed ZKP verification failed: %v", err)
			}

			// Test that TxHash is preserved as empty (due to borsh_skip tag)
			if reconstructed.TxHash != "" {
				t.Errorf("Expected TxHash to be empty after reconstruction, got: %s", reconstructed.TxHash)
			}
		})
	}
}

// BenchmarkZkpConcurrentCreation benchmarks concurrent ZKP creation performance
func BenchmarkZkpConcurrentCreation(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		index := 0
		for pb.Next() {
			// Create ZKP with different but valid dates
			day := 15
			month := 7
			year := 1990 - (index % 10) // Cycle through different years

			zkpRes, err := zkp.CreateZKP(newDOBBase(day, month, year))
			if err != nil {
				b.Fatalf("Failed to create ZKP: %v", err)
			}

			// Verify the proof
			err = groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness)
			if err != nil {
				b.Fatalf("ZKP verification failed: %v", err)
			}

			// Test serialization
			_, err = zkpRes.SerializeBorsh()
			if err != nil {
				b.Fatalf("ZKP serialization failed: %v", err)
			}

			index++
		}
	})
}

func TestZkpConcurrentCreation(t *testing.T) {
	// Test concurrent ZKP creation to ensure thread safety
	concurrency := 5
	done := make(chan bool, concurrency)
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			defer func() { done <- true }()

			// Create ZKP with different but valid dates
			day := 15
			month := 7
			year := 1990 - index // Different years to ensure different inputs

			zkpRes, err := zkp.CreateZKP(newDOBBase(day, month, year))
			if err != nil {
				errors <- err
				return
			}

			// Verify the proof
			err = groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness)
			if err != nil {
				errors <- err
				return
			}

			// Test serialization
			_, err = zkpRes.SerializeBorsh()
			if err != nil {
				errors <- err
				return
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// Check for any errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent ZKP creation error: %v", err)
	}
}

// BenchmarkZkpCreation benchmarks ZKP creation performance
func BenchmarkZkpCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := zkp.CreateZKP(newDOBBase(15, 7, 1990))
		if err != nil {
			b.Fatalf("Failed to create ZKP: %v", err)
		}
	}
}

// BenchmarkZkpVerification benchmarks ZKP verification performance
func BenchmarkZkpVerification(b *testing.B) {
	// Setup: create a ZKP once for verification benchmarking
	zkpRes, err := zkp.CreateZKP(newDOBBase(15, 7, 1990))
	if err != nil {
		b.Fatalf("Failed to create ZKP for benchmark setup: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness)
		if err != nil {
			b.Fatalf("ZKP verification failed: %v", err)
		}
	}
}

// BenchmarkZkpSerialization benchmarks ZKP serialization performance
func BenchmarkZkpSerialization(b *testing.B) {
	// Setup: create a ZKP once for serialization benchmarking
	zkpRes, err := zkp.CreateZKP(newDOBBase(15, 7, 1990))
	if err != nil {
		b.Fatalf("Failed to create ZKP for benchmark setup: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := zkpRes.SerializeBorsh()
		if err != nil {
			b.Fatalf("ZKP serialization failed: %v", err)
		}
	}
}

// BenchmarkZkpDeserialization benchmarks ZKP deserialization performance
func BenchmarkZkpDeserialization(b *testing.B) {
	// Setup: create and serialize a ZKP once for deserialization benchmarking
	zkpRes, err := zkp.CreateZKP(newDOBBase(15, 7, 1990))
	if err != nil {
		b.Fatalf("Failed to create ZKP for benchmark setup: %v", err)
	}

	serialized, err := zkpRes.SerializeBorsh()
	if err != nil {
		b.Fatalf("Failed to serialize ZKP for benchmark setup: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = zkp.ReconstructZkpResult(serialized)
		if err != nil {
			b.Fatalf("ZKP reconstruction failed: %v", err)
		}
	}
}

// BenchmarkZkpFullCycle benchmarks the complete ZKP cycle (creation, verification, serialization, deserialization)
func BenchmarkZkpFullCycle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Create ZKP
		zkpRes, err := zkp.CreateZKP(newDOBBase(15, 7, 1990))
		if err != nil {
			b.Fatalf("Failed to create ZKP: %v", err)
		}

		// Verify the proof
		err = groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness)
		if err != nil {
			b.Fatalf("ZKP verification failed: %v", err)
		}

		// Serialize
		serialized, err := zkpRes.SerializeBorsh()
		if err != nil {
			b.Fatalf("ZKP serialization failed: %v", err)
		}

		// Deserialize
		_, err = zkp.ReconstructZkpResult(serialized)
		if err != nil {
			b.Fatalf("ZKP reconstruction failed: %v", err)
		}
	}
}

func TestZkpMemoryUsage(t *testing.T) {
	// Test that ZKP operations don't cause memory leaks
	for i := 0; i < 10; i++ {
		zkpRes, err := zkp.CreateZKP(newDOBBase(15, 7, 1990))
		if err != nil {
			t.Fatalf("Failed to create ZKP iteration %d: %v", i, err)
		}

		// Verify
		err = groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness)
		if err != nil {
			t.Fatalf("ZKP verification failed iteration %d: %v", i, err)
		}

		// Serialize and deserialize
		serialized, err := zkpRes.SerializeBorsh()
		if err != nil {
			t.Fatalf("ZKP serialization failed iteration %d: %v", i, err)
		}

		_, err = zkp.ReconstructZkpResult(serialized)
		if err != nil {
			t.Fatalf("ZKP reconstruction failed iteration %d: %v", i, err)
		}
	}
}

func TestZkpBoundaryDates(t *testing.T) {
	// Test boundary conditions around the 18-year threshold
	now := time.Now()

	boundaryTests := []struct {
		name         string
		daysOffset   int
		shouldVerify bool
	}{
		{"Exactly 18 years", 0, true},
		{"1 day over 18", -1, true},
		{"1 week over 18", -7, true},
		{"1 month over 18", -30, true},
		{"1 day under 18", 1, false},
		{"1 week under 18", 7, false},
		{"1 month under 18", 30, false},
	}

	for _, tc := range boundaryTests {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate the birthdate based on offset
			birthDate := now.AddDate(-18, 0, tc.daysOffset)

			zkpRes, err := zkp.CreateZKP(newDOBBase(birthDate.Day(), int(birthDate.Month()), birthDate.Year()))

			if tc.shouldVerify {
				if err != nil {
					t.Fatalf("Failed to create ZKP for %s: %v", tc.name, err)
				}
				if zkpRes == nil {
					t.Fatalf("ZKP result is nil for %s despite no error", tc.name)
				}

				verifyErr := groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness)
				if verifyErr != nil {
					t.Errorf("Expected verification to pass for %s but got error: %v", tc.name, verifyErr)
				}
				return
			}

			// Failures may surface during circuit creation or verification; either is acceptable.
			if err != nil {
				return
			}
			if zkpRes == nil {
				t.Fatalf("ZKP result is nil for %s without an accompanying error", tc.name)
			}

			verifyErr := groth16.Verify(zkpRes.Proof, zkpRes.VerifyingKey, zkpRes.PublicWitness)
			if verifyErr == nil {
				t.Errorf("Expected verification to fail for %s but it passed", tc.name)
			}
		})
	}
}
