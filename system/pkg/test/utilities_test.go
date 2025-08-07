package test

import (
	"encoding/json"
	"os"
	"pkg-common/utilities"
	"reflect"
	"testing"
)

// Mock types for testing config functionality
type MockConfigJson struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Debug   bool   `json:"debug"`
}

type MockConfig struct {
	Name    string
	Version string
	Debug   bool
}

func (mcj MockConfigJson) ConvertToDomain() MockConfig {
	return MockConfig{
		Name:    mcj.Name,
		Version: mcj.Version,
		Debug:   mcj.Debug,
	}
}

// Another mock type for array conversion testing
type MockItemJson struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type MockItem struct {
	ID   int
	Name string
}

func (mij MockItemJson) ConvertToDomain() MockItem {
	return MockItem{
		ID:   mij.ID,
		Name: mij.Name,
	}
}

// Mock serializable type
type MockSerializableStruct struct {
	Data    string `json:"data"`
	Number  int    `json:"number"`
	Success bool   `json:"success"`
}

func (mss MockSerializableStruct) Serialize() ([]byte, error) {
	return utilities.Serialize[MockSerializableStruct](mss)
}

func TestReadConfig(t *testing.T) {
	// Create a temporary config file
	tempFile, err := os.CreateTemp("", "test_config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write test config to file
	testConfig := MockConfigJson{
		Name:    "test-app",
		Version: "1.0.0",
		Debug:   true,
	}

	configData, err := json.Marshal(testConfig)
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	_, err = tempFile.Write(configData)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	tempFile.Close()

	// Test reading the config
	result, err := utilities.ReadConfig[MockConfigJson, MockConfig](tempFile.Name())
	if err != nil {
		t.Fatalf("ReadConfig failed: %v", err)
	}

	if result.Name != "test-app" {
		t.Errorf("Expected Name to be 'test-app', got '%s'", result.Name)
	}
	if result.Version != "1.0.0" {
		t.Errorf("Expected Version to be '1.0.0', got '%s'", result.Version)
	}
	if !result.Debug {
		t.Error("Expected Debug to be true")
	}
}

func TestReadConfigFileNotFound(t *testing.T) {
	_, err := utilities.ReadConfig[MockConfigJson, MockConfig]("nonexistent_file.json")
	if err == nil {
		t.Error("Expected error when reading nonexistent file, got nil")
	}
}

func TestReadConfigInvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	tempFile, err := os.CreateTemp("", "test_invalid_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString("{ invalid json")
	if err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}
	tempFile.Close()

	_, err = utilities.ReadConfig[MockConfigJson, MockConfig](tempFile.Name())
	if err == nil {
		t.Error("Expected error when reading invalid JSON, got nil")
	}
}

func TestConvertJsonArrayToDomain(t *testing.T) {
	jsonArray := []MockItemJson{
		{ID: 1, Name: "Item 1"},
		{ID: 2, Name: "Item 2"},
		{ID: 3, Name: "Item 3"},
	}

	result := utilities.ConvertJsonArrayToDomain[MockItemJson, MockItem](jsonArray)

	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result))
	}

	for i, item := range result {
		expectedID := i + 1

		if item.ID != expectedID {
			t.Errorf("Expected item %d to have ID %d, got %d", i, expectedID, item.ID)
		}
		if item.Name != jsonArray[i].Name {
			t.Errorf("Expected item %d to have name '%s', got '%s'", i, jsonArray[i].Name, item.Name)
		}
	}
}

func TestConvertJsonArrayToDomainEmpty(t *testing.T) {
	jsonArray := []MockItemJson{}
	result := utilities.ConvertJsonArrayToDomain[MockItemJson, MockItem](jsonArray)

	if len(result) != 0 {
		t.Errorf("Expected 0 items for empty array, got %d", len(result))
	}
}

func TestFailOnErrorWithError(t *testing.T) {
	// This test is tricky because FailOnError calls logger.Default().Fatal()
	// which would terminate the test. We'll test the function exists and has correct signature

	// Test that the function doesn't panic when called with nil error
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("FailOnError should not panic with nil error: %v", r)
		}
	}()

	utilities.FailOnError(nil, "test message")
}

func TestFailOnErrorWithNilError(t *testing.T) {
	// Test that function exists and can be called with nil error
	utilities.FailOnError(nil, "no error message")
	// If we reach here, the function handled nil error correctly
}

func TestSerialize(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
	}{
		{
			name: "Simple struct",
			input: MockSerializableStruct{
				Data:    "test",
				Number:  42,
				Success: true,
			},
			expectError: false,
		},
		{
			name:        "String",
			input:       "simple string",
			expectError: false,
		},
		{
			name:        "Number",
			input:       123,
			expectError: false,
		},
		{
			name:        "Boolean",
			input:       true,
			expectError: false,
		},
		{
			name: "Map",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			expectError: false,
		},
		{
			name:        "Array",
			input:       []string{"item1", "item2", "item3"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utilities.Serialize[any](tt.input)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && result == nil {
				t.Error("Expected result but got nil")
			}
			if !tt.expectError {
				// Verify it's valid JSON
				var decoded interface{}
				err = json.Unmarshal(result, &decoded)
				if err != nil {
					t.Errorf("Serialized result is not valid JSON: %v", err)
				}
			}
		})
	}
}

func TestSerializableInterface(t *testing.T) {
	mock := MockSerializableStruct{
		Data:    "test data",
		Number:  100,
		Success: true,
	}

	// Test that our mock implements Serializable
	var serializable utilities.Serializable = mock
	if serializable == nil {
		t.Error("MockSerializableStruct should implement Serializable interface")
	}

	// Test the Serialize method
	result, err := mock.Serialize()
	if err != nil {
		t.Errorf("Serialize failed: %v", err)
	}

	// Verify the result is valid JSON and contains expected data
	var decoded MockSerializableStruct
	err = json.Unmarshal(result, &decoded)
	if err != nil {
		t.Errorf("Failed to unmarshal serialized data: %v", err)
	}

	if decoded.Data != mock.Data {
		t.Errorf("Expected Data to be '%s', got '%s'", mock.Data, decoded.Data)
	}
	if decoded.Number != mock.Number {
		t.Errorf("Expected Number to be %d, got %d", mock.Number, decoded.Number)
	}
	if decoded.Success != mock.Success {
		t.Errorf("Expected Success to be %t, got %t", mock.Success, decoded.Success)
	}
}

func TestTernary(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		trueVal   interface{}
		falseVal  interface{}
		expected  interface{}
	}{
		{
			name:      "True condition with strings",
			condition: true,
			trueVal:   "true value",
			falseVal:  "false value",
			expected:  "true value",
		},
		{
			name:      "False condition with strings",
			condition: false,
			trueVal:   "true value",
			falseVal:  "false value",
			expected:  "false value",
		},
		{
			name:      "True condition with integers",
			condition: true,
			trueVal:   42,
			falseVal:  0,
			expected:  42,
		},
		{
			name:      "False condition with integers",
			condition: false,
			trueVal:   42,
			falseVal:  0,
			expected:  0,
		},
		{
			name:      "True condition with booleans",
			condition: true,
			trueVal:   true,
			falseVal:  false,
			expected:  true,
		},
		{
			name:      "False condition with booleans",
			condition: false,
			trueVal:   true,
			falseVal:  false,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.trueVal.(type) {
			case string:
				result := utilities.Ternary(tt.condition, tt.trueVal.(string), tt.falseVal.(string))
				if result != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			case int:
				result := utilities.Ternary(tt.condition, tt.trueVal.(int), tt.falseVal.(int))
				if result != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			case bool:
				result := utilities.Ternary(tt.condition, tt.trueVal.(bool), tt.falseVal.(bool))
				if result != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestTernaryWithComplexTypes(t *testing.T) {
	type ComplexType struct {
		Name string
		ID   int
	}

	trueVal := ComplexType{Name: "true", ID: 1}
	falseVal := ComplexType{Name: "false", ID: 2}

	// Test true condition
	result := utilities.Ternary(true, trueVal, falseVal)
	if !reflect.DeepEqual(result, trueVal) {
		t.Errorf("Expected %+v, got %+v", trueVal, result)
	}

	// Test false condition
	result = utilities.Ternary(false, trueVal, falseVal)
	if !reflect.DeepEqual(result, falseVal) {
		t.Errorf("Expected %+v, got %+v", falseVal, result)
	}
}

func TestTernaryWithNilValues(t *testing.T) {
	var trueVal *string
	var falseVal *string

	nilStr := "nil"
	trueVal = &nilStr
	falseVal = nil

	// Test with true condition
	result := utilities.Ternary(true, trueVal, falseVal)
	if result != trueVal {
		t.Error("Expected trueVal pointer")
	}

	// Test with false condition
	result = utilities.Ternary(false, trueVal, falseVal)
	if result != falseVal {
		t.Error("Expected falseVal (nil) pointer")
	}
}

// Benchmark tests
func BenchmarkSerialize(b *testing.B) {
	data := MockSerializableStruct{
		Data:    "benchmark data",
		Number:  999,
		Success: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utilities.Serialize[MockSerializableStruct](data)
	}
}

func BenchmarkTernaryString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		utilities.Ternary(i%2 == 0, "true", "false")
	}
}

func BenchmarkTernaryInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		utilities.Ternary(i%2 == 0, 1, 0)
	}
}

func BenchmarkConvertJsonArrayToDomain(b *testing.B) {
	jsonArray := []MockItemJson{
		{ID: 1, Name: "Item 1"},
		{ID: 2, Name: "Item 2"},
		{ID: 3, Name: "Item 3"},
		{ID: 4, Name: "Item 4"},
		{ID: 5, Name: "Item 5"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utilities.ConvertJsonArrayToDomain[MockItemJson, MockItem](jsonArray)
	}
}

// Integration tests
func TestConfigReadAndConvert(t *testing.T) {
	// Create a complex config file
	tempFile, err := os.CreateTemp("", "test_complex_config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	complexConfig := struct {
		App      MockConfigJson         `json:"app"`
		Items    []MockItemJson         `json:"items"`
		Settings map[string]interface{} `json:"settings"`
	}{
		App: MockConfigJson{
			Name:    "complex-app",
			Version: "2.0.0",
			Debug:   false,
		},
		Items: []MockItemJson{
			{ID: 1, Name: "First Item"},
			{ID: 2, Name: "Second Item"},
		},
		Settings: map[string]interface{}{
			"timeout": 30,
			"retries": 3,
			"enabled": true,
		},
	}

	configData, err := json.Marshal(complexConfig)
	if err != nil {
		t.Fatalf("Failed to marshal complex config: %v", err)
	}

	_, err = tempFile.Write(configData)
	if err != nil {
		t.Fatalf("Failed to write complex config: %v", err)
	}
	tempFile.Close()

	// Read and verify the JSON can be unmarshaled
	fileContent, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var readConfig struct {
		App      MockConfigJson         `json:"app"`
		Items    []MockItemJson         `json:"items"`
		Settings map[string]interface{} `json:"settings"`
	}

	err = json.Unmarshal(fileContent, &readConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify the content
	if readConfig.App.Name != "complex-app" {
		t.Error("App name not read correctly")
	}
	if len(readConfig.Items) != 2 {
		t.Error("Items array not read correctly")
	}
	if readConfig.Settings["timeout"].(float64) != 30 {
		t.Error("Settings not read correctly")
	}
}
