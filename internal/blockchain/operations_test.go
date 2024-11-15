package blockchain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cosmos-evm-exporter/internal/config"
	httpClient "cosmos-evm-exporter/internal/http"
	"cosmos-evm-exporter/internal/metrics"
)

func TestGetCurrentHeight(t *testing.T) {
	metrics := metrics.NewBlockMetrics()
	testLogger := newTestLogger()

	clServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/status" {
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
		}
	}))
	defer clServer.Close()

	config := &config.Config{
		RPCEndpoint: clServer.URL,
		ETHEndpoint: "http://mock-eth-endpoint",
	}

	processor, err := NewBlockProcessor(config, metrics, testLogger)
	if err != nil {
		t.Fatalf("Failed to create BlockProcessor: %v", err)
	}

	height, err := processor.GetCurrentHeight()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedHeight := int64(7368349)
	if height != expectedHeight {
		t.Errorf("Expected height %d, got %d", expectedHeight, height)
	}
}

// Add error case tests
func TestGetCurrentELHeight_Error(t *testing.T) {
	// Setup server that returns error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := &config.Config{
		ETHEndpoint: server.URL,
	}

	processor, err := NewBlockProcessor(config, metrics.NewBlockMetrics(), newTestLogger())
	if err != nil {
		t.Fatalf("Failed to create BlockProcessor: %v", err)
	}

	_, err = processor.GetCurrentELHeight()
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestGetBlock(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/block" {
			t.Errorf("Expected path '/block', got %s", r.URL.Path)
		}

		height := r.URL.Query().Get("height")
		if height == "1000" {
			// Valid response
			response := BlockResponse{
				JsonRPC: "2.0",
				ID:      1,
				Result: struct {
					BlockID struct {
						Hash  string `json:"hash"`
						Parts struct {
							Total int    `json:"total"`
							Hash  string `json:"hash"`
						} `json:"parts"`
					} `json:"block_id"`
					Block struct {
						Header struct {
							Version struct {
								Block string `json:"block"`
								App   string `json:"app"`
							} `json:"version"`
							ChainID     string    `json:"chain_id"`
							Height      string    `json:"height"`
							Time        time.Time `json:"time"`
							LastBlockID struct {
								Hash  string `json:"hash"`
								Parts struct {
									Total int    `json:"total"`
									Hash  string `json:"hash"`
								} `json:"parts"`
							} `json:"last_block_id"`
							LastCommitHash     string `json:"last_commit_hash"`
							DataHash           string `json:"data_hash"`
							ValidatorsHash     string `json:"validators_hash"`
							NextValidatorsHash string `json:"next_validators_hash"`
							ConsensusHash      string `json:"consensus_hash"`
							AppHash            string `json:"app_hash"`
							LastResultsHash    string `json:"last_results_hash"`
							EvidenceHash       string `json:"evidence_hash"`
							ProposerAddress    string `json:"proposer_address"`
						} `json:"header"`
						Data struct {
							Txs []string `json:"txs"`
						} `json:"data"`
					} `json:"block"`
				}{
					Block: struct {
						Header struct {
							Version struct {
								Block string `json:"block"`
								App   string `json:"app"`
							} `json:"version"`
							ChainID     string    `json:"chain_id"`
							Height      string    `json:"height"`
							Time        time.Time `json:"time"`
							LastBlockID struct {
								Hash  string `json:"hash"`
								Parts struct {
									Total int    `json:"total"`
									Hash  string `json:"hash"`
								} `json:"parts"`
							} `json:"last_block_id"`
							LastCommitHash     string `json:"last_commit_hash"`
							DataHash           string `json:"data_hash"`
							ValidatorsHash     string `json:"validators_hash"`
							NextValidatorsHash string `json:"next_validators_hash"`
							ConsensusHash      string `json:"consensus_hash"`
							AppHash            string `json:"app_hash"`
							LastResultsHash    string `json:"last_results_hash"`
							EvidenceHash       string `json:"evidence_hash"`
							ProposerAddress    string `json:"proposer_address"`
						} `json:"header"`
						Data struct {
							Txs []string `json:"txs"`
						} `json:"data"`
					}{
						Header: struct {
							Version struct {
								Block string `json:"block"`
								App   string `json:"app"`
							} `json:"version"`
							ChainID     string    `json:"chain_id"`
							Height      string    `json:"height"`
							Time        time.Time `json:"time"`
							LastBlockID struct {
								Hash  string `json:"hash"`
								Parts struct {
									Total int    `json:"total"`
									Hash  string `json:"hash"`
								} `json:"parts"`
							} `json:"last_block_id"`
							LastCommitHash     string `json:"last_commit_hash"`
							DataHash           string `json:"data_hash"`
							ValidatorsHash     string `json:"validators_hash"`
							NextValidatorsHash string `json:"next_validators_hash"`
							ConsensusHash      string `json:"consensus_hash"`
							AppHash            string `json:"app_hash"`
							LastResultsHash    string `json:"last_results_hash"`
							EvidenceHash       string `json:"evidence_hash"`
							ProposerAddress    string `json:"proposer_address"`
						}{
							Height:          "1000",
							ProposerAddress: "validProposer",
							Time:            time.Now(),
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := httpClient.NewClient()

	tests := []struct {
		name        string
		height      int64
		wantError   bool
		wantAddress string
	}{
		{
			name:        "valid block",
			height:      1000,
			wantError:   false,
			wantAddress: "validProposer",
		},
		{
			name:      "server error",
			height:    500,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, err := GetBlock(client, server.URL, tt.height)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if block.Result.Block.Header.ProposerAddress != tt.wantAddress {
				t.Errorf("Expected proposer %s, got %s",
					tt.wantAddress,
					block.Result.Block.Header.ProposerAddress)
			}
		})
	}
}

func TestGetCurrentELHeight(t *testing.T) {
	metrics := metrics.NewBlockMetrics()
	testLogger := newTestLogger()

	tests := []struct {
		name       string
		response   string
		wantError  bool
		wantHeight int64
	}{
		{
			name:       "valid height",
			response:   `{"jsonrpc":"2.0","id":"1","result":"0x1234"}`,
			wantError:  false,
			wantHeight: 0x1234,
		},
		{
			name:      "invalid hex",
			response:  `{"jsonrpc":"2.0","id":"1","result":"invalid"}`,
			wantError: true,
		},
		{
			name:      "empty response",
			response:  `{"jsonrpc":"2.0","id":"1","result":""}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintln(w, tt.response)
			}))
			defer server.Close()

			config := &config.Config{
				ETHEndpoint: server.URL,
			}

			processor, err := NewBlockProcessor(config, metrics, testLogger)
			if err != nil {
				t.Fatalf("Failed to create BlockProcessor: %v", err)
			}

			height, err := processor.GetCurrentELHeight()

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if height != tt.wantHeight {
				t.Errorf("Expected height %d, got %d", tt.wantHeight, height)
			}
		})
	}
}
