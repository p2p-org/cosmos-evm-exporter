package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	ethClient *ethclient.Client
}

func NewClient(endpoint string) (*Client, error) {
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, err
	}
	return &Client{
		ethClient: client,
	}, nil
}

func (c *Client) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return c.ethClient.BlockByNumber(ctx, number)
}

// Close releases any resources used by the client
func (c *Client) Close() {
	if c.ethClient != nil {
		c.ethClient.Close()
	}
}
