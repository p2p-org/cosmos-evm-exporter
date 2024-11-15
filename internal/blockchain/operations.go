package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	httpClient "cosmos-evm-exporter/internal/http"
)

func GetBlock(client *httpClient.Client, endpoint string, height int64) (*BlockResponse, error) {
	maxRetries := 6
	retryDelay := 2 * time.Second

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		url := fmt.Sprintf("%s/block?height=%d", endpoint, height)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.DoRequest(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to fetch block: %w", err)
			time.Sleep(retryDelay)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			time.Sleep(retryDelay)
			continue
		}

		if len(body) < 100 {
			lastErr = fmt.Errorf("received suspicious response: %s", string(body))
			time.Sleep(retryDelay)
			continue
		}

		var blockResp BlockResponse
		if err := json.Unmarshal(body, &blockResp); err != nil {
			lastErr = fmt.Errorf("failed to parse block response: %w", err)
			time.Sleep(retryDelay)
			continue
		}

		proposerAddr := blockResp.Result.Block.Header.ProposerAddress
		if proposerAddr == "" {
			lastErr = fmt.Errorf("received empty proposer address")
			if attempt < maxRetries-1 {
				time.Sleep(retryDelay)
				continue
			}
		} else {
			return &blockResp, nil
		}
	}

	return nil, lastErr
}

func (p *BlockProcessor) GetCurrentHeight() (int64, error) {
	req, err := http.NewRequest("GET", p.config.RPCEndpoint+"/status", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	client := httpClient.NewClient()
	resp, err := client.DoRequest(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch status: %w", err)
	}
	defer resp.Body.Close()

	var status StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return 0, fmt.Errorf("failed to parse status response: %w", err)
	}

	var height int64
	_, err = fmt.Sscanf(status.Result.SyncInfo.LatestBlockHeight, "%d", &height)
	if err != nil {
		return 0, fmt.Errorf("failed to parse block height: %w", err)
	}

	return height, nil
}

func (p *BlockProcessor) GetCurrentELHeight() (int64, error) {
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
		return 0, fmt.Errorf("failed to marshal EL request: %w", err)
	}

	client := httpClient.NewClient()
	req, err := http.NewRequest("POST", p.config.ETHEndpoint, bytes.NewBuffer(elBody))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.DoRequest(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch EL height: %w", err)
	}
	defer resp.Body.Close()

	var elResp struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&elResp); err != nil {
		return 0, fmt.Errorf("failed to decode EL response: %w", err)
	}

	elHeight, err := strconv.ParseInt(strings.TrimPrefix(elResp.Result, "0x"), 16, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse EL height: %w", err)
	}

	return elHeight, nil
}
