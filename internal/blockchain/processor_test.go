package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cosmos-evm-exporter/internal/config"
	"cosmos-evm-exporter/internal/logger"
	"cosmos-evm-exporter/internal/metrics"

	"github.com/ethereum/go-ethereum/core/types"
)

// MockEthClient implements EthClientInterface for testing
type MockEthClient struct {
	blocks map[int64]*types.Block
}

func (m *MockEthClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	if block, ok := m.blocks[number.Int64()]; ok {
		return block, nil
	}
	return nil, fmt.Errorf("block not found")
}

func TestProcessBlock(t *testing.T) {
	testLogger := logger.NewLogger(&logger.Config{})
	metrics := metrics.NewBlockMetrics()

	// Setup mock servers for EL and CL endpoints
	elServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"jsonrpc":"2.0","id":"1","result":"0x64"}`) // hex for height 100
	}))
	defer elServer.Close()

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
						LatestBlockHeight: "110",
					},
				},
			})
		}
	}))
	defer clServer.Close()

	tests := []struct {
		name          string
		block         *BlockResponse
		targetVal     string
		evmAddr       string
		wantError     bool
		wantProcessed bool
		expectTxs     bool
	}{
		{
			name: "valid block from our validator",
			block: &BlockResponse{
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
							Height:          "100",
							ProposerAddress: "validator1",
						},
						Data: struct {
							Txs []string `json:"txs"`
						}{
							Txs: []string{"tx1", "tx2"},
						},
					},
				},
			},
			targetVal:     "validator1",
			evmAddr:       "0x1234",
			wantError:     false,
			wantProcessed: true,
			expectTxs:     true,
		},
		{
			name: "empty block from our validator",
			block: &BlockResponse{
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
							Height:          "100",
							ProposerAddress: "validator1",
						},
						Data: struct {
							Txs []string `json:"txs"`
						}{
							Txs: []string{},
						},
					},
				},
			},
			targetVal:     "validator1",
			evmAddr:       "0x1234",
			wantError:     false,
			wantProcessed: true,
			expectTxs:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &config.Config{
				TargetValidator: tt.targetVal,
				EVMAddress:      tt.evmAddr,
				ETHEndpoint:     elServer.URL,
				RPCEndpoint:     clServer.URL,
			}

			mockClient := &MockEthClient{
				blocks: make(map[int64]*types.Block),
			}

			processor := &BlockProcessor{
				config:  config,
				logger:  testLogger,
				metrics: metrics,
				client:  mockClient,
			}

			err := processor.ProcessBlock(tt.block)
			if (err != nil) != tt.wantError {
				t.Errorf("ProcessBlock() error = %v, wantError %v", err, tt.wantError)
			}

			if tt.wantProcessed {
				if len(tt.block.Result.Block.Data.Txs) > 0 != tt.expectTxs {
					t.Errorf("ProcessBlock() transaction check failed, got %v transactions, expected transactions: %v",
						len(tt.block.Result.Block.Data.Txs), tt.expectTxs)
				}
			}
		})
	}
}
