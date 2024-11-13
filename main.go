package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	EVMAddress      string `toml:"evm_address"`
	TargetValidator string `toml:"target_validator"`
	RPCEndpoint     string `toml:"rpc_endpoint"`
	LogFile         string `toml:"log_file"`
	MetricsPort     string `toml:"metrics_port"`
	EnableFileLog   bool   `toml:"enable_file_log"`
	EnableStdout    bool   `toml:"enable_stdout"`
	ETHEndpoint     string `toml:"eth_endpoint"`
}

var config *Config

func loadConfig(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}

	// Validate required fields
	if cfg.EVMAddress == "" {
		return nil, fmt.Errorf("evm_address is required")
	}
	if cfg.TargetValidator == "" {
		return nil, fmt.Errorf("target_validator is required")
	}
	if cfg.RPCEndpoint == "" {
		return nil, fmt.Errorf("rpc_endpoint is required")
	}
	if cfg.ETHEndpoint == "" {
		return nil, fmt.Errorf("eth_endpoint is required")
	}
	if cfg.MetricsPort == "" {
		cfg.MetricsPort = ":2113" // Set default if not specified
	}

	return &cfg, nil
}

var (
	blockMetrics = struct {
		totalProposed        prometheus.Counter
		executionConfirmed   prometheus.Counter
		executionMissed      prometheus.Counter
		emptyConsensusBlocks prometheus.Counter
		emptyExecutionBlocks prometheus.Counter
		errors               prometheus.Counter
		currentHeight        prometheus.Gauge
		elToClGap            prometheus.Gauge // New metric to track the gap
	}{
		totalProposed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "validator_total_blocks_proposed",
			Help: "Total number of blocks proposed by our validator",
		}),
		executionConfirmed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "validator_execution_blocks_confirmed",
			Help: "Number of proposed blocks that made it to the execution layer",
		}),
		executionMissed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "validator_execution_blocks_missed",
			Help: "Number of proposed blocks that failed to make it to the execution layer",
		}),
		emptyConsensusBlocks: promauto.NewCounter(prometheus.CounterOpts{
			Name: "validator_empty_consensus_blocks",
			Help: "Number of blocks proposed with no transactions on consensus layer",
		}),
		emptyExecutionBlocks: promauto.NewCounter(prometheus.CounterOpts{
			Name: "validator_empty_execution_blocks",
			Help: "Number of blocks confirmed on execution layer with no transactions",
		}),
		errors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "validator_block_processing_errors",
			Help: "Number of errors encountered while processing blocks",
		}),
		currentHeight: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "validator_current_block_height",
			Help: "Current block height being processed",
		}),
		elToClGap: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "validator_el_to_cl_gap",
			Help: "Gap between execution and consensus layer block heights",
		}),
	}
)

// Status response structures
type StatusResponse struct {
	Result struct {
		SyncInfo struct {
			LatestBlockHeight string `json:"latest_block_height"`
		} `json:"sync_info"`
	} `json:"result"`
}

// Block response structures
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
			Evidence struct {
				Evidence []interface{} `json:"evidence"`
			} `json:"evidence"`
			LastCommit struct {
				Height  string `json:"height"`
				Round   int    `json:"round"`
				BlockID struct {
					Hash  string `json:"hash"`
					Parts struct {
						Total int    `json:"total"`
						Hash  string `json:"hash"`
					} `json:"parts"`
				} `json:"block_id"`
				Signatures []struct {
					BlockIDFlag      int       `json:"block_id_flag"`
					ValidatorAddress string    `json:"validator_address"`
					Timestamp        time.Time `json:"timestamp"`
					Signature        string    `json:"signature"`
				} `json:"signatures"`
			} `json:"last_commit"`
		} `json:"block"`
	} `json:"result"`
}

// Add structure for parsing the transaction
type BerachainTx struct {
	MsgType    uint32
	DataLength uint32
	Payload    []byte
}

var (
	ethClient    *ethclient.Client
	stdoutLogger *log.Logger
	fileLogger   *log.Logger
)

func getCurrentHeight(rpcEndpoint string) (int64, error) {
	resp, err := http.Get(rpcEndpoint + "/status")
	if err != nil {
		return 0, fmt.Errorf("failed to fetch status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	var status StatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		return 0, fmt.Errorf("failed to parse status response: %w", err)
	}

	var height int64
	_, err = fmt.Sscanf(status.Result.SyncInfo.LatestBlockHeight, "%d", &height)
	if err != nil {
		return 0, fmt.Errorf("failed to parse block height: %w", err)
	}

	return height, nil
}

func getBlock(rpcEndpoint string, height int64) (*BlockResponse, error) {
	resp, err := http.Get(fmt.Sprintf("%s/block?height=%d", rpcEndpoint, height))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch block: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if len(body) < 100 {
		return nil, fmt.Errorf("received suspicious response: %s", string(body))
	}

	var blockResp BlockResponse
	if err := json.Unmarshal(body, &blockResp); err != nil {
		return nil, fmt.Errorf("failed to parse block response: %w", err)
	}

	proposerAddr := blockResp.Result.Block.Header.ProposerAddress
	if proposerAddr == "" {
		log.Printf("Warning: Empty proposer address for block %d", height)
		return nil, fmt.Errorf("received empty proposer address")
	}

	return &blockResp, nil
}

func setupLogger() (*log.Logger, *log.Logger) {
	var fileLogger, stdoutLogger *log.Logger

	if config.EnableFileLog {
		logFile, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		fileLogger = log.New(logFile, "", log.LstdFlags)
	} else {
		fileLogger = log.New(io.Discard, "", 0)
	}

	if config.EnableStdout {
		stdoutLogger = log.New(os.Stdout, "", log.LstdFlags)
	} else {
		stdoutLogger = log.New(io.Discard, "", 0)
	}

	return fileLogger, stdoutLogger
}

// Add this function to decode base64 transactions
func decodeTx(txBase64 string) (*BerachainTx, error) {
	// Decode base64
	txData, err := base64.StdEncoding.DecodeString(txBase64)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %v", err)
	}

	// Debug log the raw data
	stdoutLogger.Printf("Raw tx data length: %d bytes", len(txData))
	if len(txData) < 8 {
		return nil, fmt.Errorf("tx data too short: %d bytes", len(txData))
	}

	// Create BerachainTx
	tx := &BerachainTx{
		MsgType:    binary.BigEndian.Uint32(txData[0:4]),
		DataLength: binary.BigEndian.Uint32(txData[4:8]),
	}

	// Debug log the message type and length
	stdoutLogger.Printf("Message Type: 0x%x", tx.MsgType)
	stdoutLogger.Printf("Data Length: %d", tx.DataLength)

	// Verify payload length
	if len(txData) < 8+int(tx.DataLength) {
		return nil, fmt.Errorf("payload length mismatch: expected %d, got %d", tx.DataLength, len(txData)-8)
	}

	// Copy payload
	tx.Payload = make([]byte, tx.DataLength)
	copy(tx.Payload, txData[8:8+tx.DataLength])

	return tx, nil
}

// Add this function to help with debugging
func dumpPayload(payload []byte) {
	stdoutLogger.Printf("  Payload length: %d bytes", len(payload))
	stdoutLogger.Printf("  Full payload hex: %x", payload)

	// Try to show it in chunks for better readability
	for i := 0; i < len(payload); i += 32 {
		end := i + 32
		if end > len(payload) {
			end = len(payload)
		}
		stdoutLogger.Printf("  Bytes %d-%d: %x", i, end-1, payload[i:end])
	}
}

func getCurrentGap() (int64, error) {
	// Get EL block height
	var elResp struct {
		Result string `json:"result"`
	}
	elReq := struct {
		Jsonrpc string   `json:"jsonrpc"`
		Id      string   `json:"id"`
		Method  string   `json:"method"`
		Params  []string `json:"params"`
	}{
		Jsonrpc: "2.0",
		Id:      "1",
		Method:  "eth_blockNumber",
		Params:  []string{},
	}

	elBody, err := json.Marshal(elReq)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal EL request: %v", err)
	}

	resp, err := http.Post(config.ETHEndpoint, "application/json", bytes.NewBuffer(elBody))
	if err != nil {
		return 0, fmt.Errorf("failed to fetch EL height: %v", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&elResp); err != nil {
		return 0, fmt.Errorf("failed to decode EL response: %v", err)
	}

	elHeight, err := strconv.ParseInt(strings.TrimPrefix(elResp.Result, "0x"), 16, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse EL height: %v", err)
	}

	clHeight, err := getCurrentHeight(config.RPCEndpoint)
	if err != nil {
		return 0, fmt.Errorf("failed to get CL height: %v", err)
	}

	return clHeight - elHeight, nil
}

// Remove the duplicate Block type and update processBlock to use BlockResponse
func processBlock(block *BlockResponse) error {
	if block.Result.Block.Header.ProposerAddress != config.TargetValidator {
		return nil // Not our validator
	}

	clHeight, _ := strconv.ParseInt(block.Result.Block.Header.Height, 10, 64)
	blockMetrics.currentHeight.Set(float64(clHeight))
	blockMetrics.totalProposed.Inc()

	// Get current gap
	gap, err := getCurrentGap()
	if err != nil {
		return fmt.Errorf("failed to get current gap: %v", err)
	}

	// Calculate expected EL height
	expectedELHeight := clHeight - gap

	stdoutLogger.Printf("\n=== Processing Validator Block ===")
	stdoutLogger.Printf("CL Height: %d", clHeight)
	stdoutLogger.Printf("Expected EL Height: %d (gap: %d)", expectedELHeight, gap)

	// Check blocks around our expected height
	startHeight := expectedELHeight - 2
	endHeight := expectedELHeight + 2

	for height := startHeight; height <= endHeight; height++ {
		block, err := ethClient.BlockByNumber(context.Background(), big.NewInt(height))
		if err != nil {
			stdoutLogger.Printf("Failed to fetch block %d: %v", height, err)
			continue
		}

		if strings.EqualFold(block.Coinbase().Hex(), config.EVMAddress) {
			stdoutLogger.Printf("✅ Found our block!")
			stdoutLogger.Printf("EL Height: %d", height)
			stdoutLogger.Printf("Hash: %s", block.Hash().Hex())
			blockMetrics.executionConfirmed.Inc()
			return nil
		}
	}

	stdoutLogger.Printf("❌ Block not found in range %d to %d", startHeight, endHeight)
	blockMetrics.executionMissed.Inc()
	return nil
}

func getCurrentELHeight() (int64, error) {
	var elResp struct {
		Result string `json:"result"`
	}
	elReq := struct {
		Jsonrpc string   `json:"jsonrpc"`
		Id      string   `json:"id"`
		Method  string   `json:"method"`
		Params  []string `json:"params"`
	}{
		Jsonrpc: "2.0",
		Id:      "1",
		Method:  "eth_blockNumber",
		Params:  []string{},
	}

	elBody, err := json.Marshal(elReq)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal EL request: %v", err)
	}

	resp, err := http.Post(config.ETHEndpoint, "application/json", bytes.NewBuffer(elBody))
	if err != nil {
		return 0, fmt.Errorf("failed to fetch EL height: %v", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&elResp); err != nil {
		return 0, fmt.Errorf("failed to decode EL response: %v", err)
	}

	return strconv.ParseInt(strings.TrimPrefix(elResp.Result, "0x"), 16, 64)
}

func main() {
	// Check for config flag
	if len(os.Args) != 3 || os.Args[1] != "--config" {
		log.Fatal("Usage: go run main.go --config=./config.toml")
	}

	// Load config
	var err error
	config, err = loadConfig(os.Args[2])
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup loggers
	fileLogger, stdoutLogger = setupLogger()

	// Connect to ETH client
	ethClient, err = ethclient.Dial(config.ETHEndpoint)
	if err != nil {
		fileLogger.Fatalf("Failed to connect to ETH client: %v", err)
	}

	// Get current height
	currentHeight, err := getCurrentHeight(config.RPCEndpoint)
	if err != nil {
		fileLogger.Fatalf("Failed to get current height: %v", err)
	}

	fileLogger.Printf("Starting monitor from height %d for validator %s", currentHeight, config.TargetValidator)
	fileLogger.Printf("-------------------")

	// Remove unused statistics variables
	var errorBlocks int // Keep this one as it's still used in error handling

	// Start Prometheus metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(config.MetricsPort, nil); err != nil {
			log.Fatal(err)
		}
	}()

	// Start a goroutine to continuously update the gap metric
	go func() {
		for {
			gap, err := getCurrentGap()
			if err != nil {
				stdoutLogger.Printf("Failed to get current gap: %v", err)
			} else {
				blockMetrics.elToClGap.Set(float64(gap))
			}
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		// Get the latest height first
		latestHeight, err := getCurrentHeight(config.RPCEndpoint)
		if err != nil {
			stdoutLogger.Printf("Failed to get latest height: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Don't process if we've caught up to the latest block
		if currentHeight > latestHeight {
			time.Sleep(1 * time.Second)
			continue
		}

		// Process all blocks between currentHeight and latestHeight
		for currentHeight <= latestHeight {
			blockMetrics.currentHeight.Set(float64(currentHeight))

			block, err := getBlock(config.RPCEndpoint, currentHeight)
			if err != nil {
				blockMetrics.errors.Inc()
				stdoutLogger.Printf("Failed to get block %d: %v", currentHeight, err)
				errorBlocks++
				time.Sleep(2 * time.Second)
				continue
			}

			// Process the block
			if err := processBlock(block); err != nil {
				stdoutLogger.Printf("Error processing block %d: %v", currentHeight, err)
				errorBlocks++
			}

			currentHeight++
		}

		time.Sleep(500 * time.Millisecond)
	}
}
