package blockchain

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cosmos-evm-exporter/internal/config"
	"cosmos-evm-exporter/internal/metrics"
)

func TestDumpPayload(t *testing.T) {
	metrics := metrics.NewBlockMetrics()
	testLogger := newTestLogger()

	// Setup mock servers
	elServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  "0x67E1B5",
		})
	}))
	defer elServer.Close()

	clServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(StatusResponse{
			Result: struct {
				SyncInfo struct {
					LatestBlockHeight string `json:"latest_block_height"`
				} `json:"sync_info"`
			}{
				SyncInfo: struct {
					LatestBlockHeight string `json:"latest_block_height"`
				}{
					LatestBlockHeight: "7368349",
				},
			},
		})
	}))
	defer clServer.Close()

	config := &config.Config{
		ETHEndpoint: elServer.URL,
		RPCEndpoint: clServer.URL,
	}

	processor, err := NewBlockProcessor(config, metrics, testLogger)
	if err != nil {
		t.Fatalf("Failed to create BlockProcessor: %v", err)
	}

	testCases := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "empty payload",
			payload: []byte{},
		},
		{
			name:    "small payload",
			payload: []byte("test"),
		},
		{
			name:    "payload larger than 32 bytes",
			payload: []byte("this is a test payload that is longer than 32 bytes to test chunking"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Just verify it doesn't panic
			processor.DumpPayload(tc.payload)
		})
	}
}
