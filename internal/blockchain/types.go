package blockchain

import (
	"context"
	"math/big"
	"time"

	"cosmos-evm-exporter/internal/config"
	"cosmos-evm-exporter/internal/logger"
	"cosmos-evm-exporter/internal/metrics"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/prometheus/client_golang/prometheus"
)

// BlockResponse defines the structure for consensus layer block API responses
type BlockResponse struct {
	JsonRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
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
	} `json:"result"`
}

type StatusResponse struct {
	Result struct {
		SyncInfo struct {
			LatestBlockHeight string `json:"latest_block_height"`
		} `json:"sync_info"`
	} `json:"result"`
}

type EthRequest struct {
	Jsonrpc string   `json:"jsonrpc"`
	ID      string   `json:"id"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
}

type EthResponse struct {
	Result string `json:"result"`
}

type Config struct {
	EVMAddress      string `toml:"evm_address"`
	TargetValidator string `toml:"target_validator"`
	ETHEndpoint     string `toml:"eth_endpoint"`
	RPCEndpoint     string `toml:"rpc_endpoint"`
	EnableFileLog   bool   `toml:"enable_file_log"`
	EnableStdout    bool   `toml:"enable_stdout"`
	LogFile         string `toml:"log_file"`
}

type BlockMetrics struct {
	TotalProposed        prometheus.Counter
	ExecutionConfirmed   prometheus.Counter
	ExecutionMissed      prometheus.Counter
	EmptyConsensusBlocks prometheus.Counter
	EmptyExecutionBlocks prometheus.Counter
	Errors               prometheus.Counter
	CurrentHeight        prometheus.Gauge
	ElToClGap            prometheus.Gauge
}

type EthClientInterface interface {
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
}

type BlockProcessor struct {
	config  *config.Config
	metrics *metrics.BlockMetrics
	client  EthClientInterface
	logger  *logger.Logger
}
type EVMChainTx struct {
	MsgType    uint32
	DataLength uint32
	Payload    []byte
}
