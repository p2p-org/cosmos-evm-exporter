package rpc

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  "0x1",
		})
	}))
	defer server.Close()

	tests := []struct {
		name      string
		endpoint  string
		wantError bool
	}{
		{
			name:      "valid endpoint",
			endpoint:  server.URL,
			wantError: false,
		},
		{
			name:      "invalid endpoint",
			endpoint:  "invalid-url",
			wantError: true,
		},
		{
			name:      "empty endpoint",
			endpoint:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.endpoint)
			if (err != nil) != tt.wantError {
				t.Errorf("NewClient() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && client == nil {
				t.Error("NewClient() returned nil client without error")
			}
			if client != nil {
				client.Close()
			}
		})
	}
}

func TestBlockByNumber(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result": map[string]interface{}{
				"number":           "0x1",
				"hash":             "0x1234567890123456789012345678901234567890123456789012345678901234",
				"parentHash":       "0x1234567890123456789012345678901234567890123456789012345678901234",
				"sha3Uncles":       "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
				"logsBloom":        "0x" + strings.Repeat("0", 512),
				"transactionsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
				"stateRoot":        "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
				"receiptsRoot":     "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
				"miner":            "0x0000000000000000000000000000000000000000",
				"difficulty":       "0x0",
				"totalDifficulty":  "0x0",
				"extraData":        "0x",
				"size":             "0x0",
				"gasLimit":         "0x0",
				"gasUsed":          "0x0",
				"timestamp":        "0x0",
				"transactions":     []string{},
				"uncles":           []string{},
				"mixHash":          "0x0000000000000000000000000000000000000000000000000000000000000000",
				"nonce":            "0x0000000000000000",
				"baseFeePerGas":    "0x0",
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	tests := []struct {
		name      string
		number    *big.Int
		wantError bool
	}{
		{
			name:      "valid block number",
			number:    big.NewInt(1),
			wantError: false,
		},
		{
			name:      "nil block number (latest)",
			number:    nil,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, err := client.BlockByNumber(context.Background(), tt.number)
			if (err != nil) != tt.wantError {
				t.Errorf("BlockByNumber() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && block == nil {
				t.Error("BlockByNumber() returned nil block without error")
			}
		})
	}
}
