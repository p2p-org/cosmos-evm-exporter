package logger

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"strings"
	"testing"
)

func TestWriteJSONLog(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer

	// Create logger with buffer
	testLogger := &Logger{
		logger: log.New(&buf, "", 0),
		config: &Config{
			EnableStdout: true,
			LogFile:      "test.log",
		},
	}

	// Create test metadata
	testMetadata := map[string]interface{}{
		"cl_height":          int64(7368349),
		"expected_el_height": int64(7368343),
		"gap":                int64(6),
	}

	// Call WriteJSONLog
	testLogger.WriteJSONLog("info", "Processing Validator Block", testMetadata, nil)

	// Parse the log output
	var logEntry struct {
		Level     string                 `json:"level"`
		Message   string                 `json:"message"`
		Data      map[string]interface{} `json:"data"`
		Error     interface{}            `json:"error"`
		Timestamp string                 `json:"timestamp"`
	}

	if err := json.NewDecoder(&buf).Decode(&logEntry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	// Verify the log structure
	if logEntry.Level != "info" {
		t.Errorf("Expected level 'info', got %v", logEntry.Level)
	}

	if !strings.Contains(logEntry.Message, "Processing Validator Block") {
		t.Errorf("Expected message to contain 'Processing Validator Block', got %v", logEntry.Message)
	}

	// Verify data exists and has correct values
	if logEntry.Data == nil {
		t.Fatal("Expected non-nil data")
	}

	// Check all expected fields exist in data with correct values
	expectedFields := map[string]int64{
		"cl_height":          7368349,
		"expected_el_height": 7368343,
		"gap":                6,
	}

	for field, expectedValue := range expectedFields {
		if value, exists := logEntry.Data[field]; !exists {
			t.Errorf("Expected data to contain field %s", field)
		} else {
			// Convert the value to float64 since JSON numbers are decoded as float64
			if floatVal, ok := value.(float64); !ok {
				t.Errorf("Expected %s to be a number, got %T", field, value)
			} else if int64(floatVal) != expectedValue {
				t.Errorf("Expected %s to be %d, got %f", field, expectedValue, floatVal)
			}
		}
	}
}

func TestSetupLogger(t *testing.T) {
	// Create temp log file
	tmpfile, err := os.CreateTemp("", "test.*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Setup test config
	testConfig := &Config{
		EnableFileLog: true,
		EnableStdout:  true,
		LogFile:       tmpfile.Name(),
	}

	logger := NewLogger(testConfig)
	if logger == nil {
		t.Error("Expected non-nil logger")
	}
}
