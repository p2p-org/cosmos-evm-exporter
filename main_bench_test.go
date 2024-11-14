package main

import (
	"testing"
)

func BenchmarkProcessBlock(b *testing.B) {
	// Setup test config and block
	config = &Config{
		EVMAddress:      "0x1234",
		TargetValidator: "ABCD",
	}

	block := &BlockResponse{}
	block.Result.Block.Header.ProposerAddress = "ABCD"
	block.Result.Block.Header.Height = "7368349"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processBlock(block)
	}
}

func BenchmarkGetCurrentGap(b *testing.B) {
	// Setup test config
	config = &Config{
		RPCEndpoint: "http://localhost:26657",
		ETHEndpoint: "http://localhost:8545",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = getCurrentGap()
	}
}
