package http

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.client == nil {
		t.Fatal("HTTP client is nil")
	}

	// Test that client has expected timeout
	if client.client.Timeout != 10*time.Second {
		t.Errorf("Expected 10s timeout, got %v", client.client.Timeout)
	}

	// Test number of retries
	if client.retries != 3 {
		t.Errorf("Expected 3 retries, got %d", client.retries)
	}
}
