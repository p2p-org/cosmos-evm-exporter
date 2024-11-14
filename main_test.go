package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
)

func TestLoadConfig(t *testing.T) {
	// Create temporary config file
	tmpfile, err := os.CreateTemp("", "config.*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write test config
	configContent := `
evm_address = "0x1234567890123456789012345678901234567890"
target_validator = "0xabcdef1234567890"
rpc_endpoint = "http://localhost:26657"
eth_endpoint = "http://localhost:8545"
log_file = "test.log"
metrics_port = ":2113"
enable_file_log = true
enable_stdout = true
`
	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Test valid config
	cfg, err := loadConfig(tmpfile.Name())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if cfg.EVMAddress != "0x1234567890123456789012345678901234567890" {
		t.Errorf("Wrong EVM address")
	}

	// Test missing required fields
	tmpfile2, _ := os.CreateTemp("", "config.*.toml")
	defer os.Remove(tmpfile2.Name())
	tmpfile2.Write([]byte(`metrics_port = ":2113"`))
	tmpfile2.Close()

	_, err = loadConfig(tmpfile2.Name())
	if err == nil {
		t.Error("Expected error for missing required fields")
	}
}

func TestHTTPClientRetries(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := newHTTPClient()
	client.retries = 2

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.doRequest(req)

	if err != nil {
		t.Errorf("Expected success after retries, got error: %v", err)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 status, got %d", resp.StatusCode)
	}
}

func TestGetCurrentHeight(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/status" {
			t.Errorf("Expected /status path, got %s", r.URL.Path)
		}
		resp := StatusResponse{}
		resp.Result.SyncInfo.LatestBlockHeight = "123456"
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	height, err := getCurrentHeight(server.URL)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if height != 123456 {
		t.Errorf("Expected height 123456, got %d", height)
	}
}

func TestProcessBlock(t *testing.T) {
	// Setup mock servers
	elServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var req struct {
				Jsonrpc string      `json:"jsonrpc"`
				Method  string      `json:"method"`
				Params  interface{} `json:"params"`
				Id      interface{} `json:"id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("Failed to decode request: %v", err)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			resp := struct {
				Jsonrpc string      `json:"jsonrpc"`
				Id      interface{} `json:"id"`
				Result  string      `json:"result"`
			}{
				Jsonrpc: "2.0",
				Id:      req.Id,
				Result:  "0x67E1B5",
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer elServer.Close()

	clServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := StatusResponse{}
		resp.Result.SyncInfo.LatestBlockHeight = "7368349"
		json.NewEncoder(w).Encode(resp)
	}))
	defer clServer.Close()

	// Setup test config
	config = &Config{
		EVMAddress:      "0x1234",
		TargetValidator: "ABCD",
		ETHEndpoint:     elServer.URL,
		RPCEndpoint:     clServer.URL,
	}

	// Create test block
	block := &BlockResponse{}
	block.Result.Block.Header.ProposerAddress = "ABCD"
	block.Result.Block.Header.Height = "7368349"

	// Setup mock eth client
	var err error
	ethClient, err = ethclient.Dial(elServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	// Setup logger
	stdoutLogger = log.New(io.Discard, "", 0) // Suppress logs during test

	err = processBlock(block)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestGetCurrentGap(t *testing.T) {
	// Setup mock servers
	clServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/status" {
			t.Errorf("Expected /status path, got %s", r.URL.Path)
		}
		resp := StatusResponse{
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
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer clServer.Close()

	elServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var req struct {
				Jsonrpc string      `json:"jsonrpc"`
				Method  string      `json:"method"`
				Params  interface{} `json:"params"`
				Id      interface{} `json:"id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("Failed to decode request: %v", err)
				return
			}

			if req.Method != "eth_blockNumber" {
				t.Errorf("Expected eth_blockNumber method, got %s", req.Method)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			resp := struct {
				Jsonrpc string      `json:"jsonrpc"`
				Id      interface{} `json:"id"`
				Result  string      `json:"result"`
			}{
				Jsonrpc: "2.0",
				Id:      req.Id,
				Result:  "0x67e44b",
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer elServer.Close()

	config = &Config{
		RPCEndpoint: clServer.URL,
		ETHEndpoint: elServer.URL,
	}

	gap, err := getCurrentGap()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	clHeight := int64(7368349)
	elHeight := int64(6808651) // 0x67E1B5 in decimal
	expectedGap := clHeight - elHeight

	if gap != expectedGap {
		t.Errorf("Expected gap %d, got %d (CL: %d, EL: %d, EL hex: 0x%x)",
			expectedGap, gap, clHeight, elHeight, elHeight)
	}
}
