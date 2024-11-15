package benchmarks

import (
	"testing"

	"cosmos-evm-exporter/internal/blockchain"
	"cosmos-evm-exporter/internal/config"
	"cosmos-evm-exporter/internal/logger"
	"cosmos-evm-exporter/internal/metrics"
)

func BenchmarkProcessBlock(b *testing.B) {
	metrics := metrics.NewBlockMetrics()
	testLogger := logger.NewLogger(&logger.Config{
		EnableStdout: true,
		LogFile:      "test.log",
	})

	config := &config.Config{
		EVMAddress:      "0x1234",
		TargetValidator: "ABCD",
	}

	processor, err := blockchain.NewBlockProcessor(config, metrics, testLogger)
	if err != nil {
		b.Fatalf("Failed to create BlockProcessor: %v", err)
	}

	block := &blockchain.BlockResponse{}
	block.Result.Block.Header.ProposerAddress = "ABCD"
	block.Result.Block.Header.Height = "7368349"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processor.ProcessBlock(block)
	}
}

func BenchmarkGetCurrentGap(b *testing.B) {
	metrics := metrics.NewBlockMetrics()
	testLogger := logger.NewLogger(&logger.Config{
		EnableStdout: true,
		LogFile:      "test.log",
	})

	config := &config.Config{
		RPCEndpoint: "http://localhost:26657",
		ETHEndpoint: "http://localhost:8545",
	}

	processor, err := blockchain.NewBlockProcessor(config, metrics, testLogger)
	if err != nil {
		b.Fatalf("Failed to create BlockProcessor: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = processor.GetCurrentGap()
	}
}

func BenchmarkGetCurrentHeight(b *testing.B) {
	metrics := metrics.NewBlockMetrics()
	testLogger := logger.NewLogger(&logger.Config{
		EnableStdout: true,
		LogFile:      "test.log",
	})

	config := &config.Config{
		RPCEndpoint: "http://localhost:26657",
	}

	processor, err := blockchain.NewBlockProcessor(config, metrics, testLogger)
	if err != nil {
		b.Fatalf("Failed to create BlockProcessor: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = processor.GetCurrentHeight()
	}
}

func BenchmarkGetCurrentELHeight(b *testing.B) {
	metrics := metrics.NewBlockMetrics()
	testLogger := logger.NewLogger(&logger.Config{
		EnableStdout: true,
		LogFile:      "test.log",
	})

	config := &config.Config{
		ETHEndpoint: "http://localhost:8545",
	}

	processor, err := blockchain.NewBlockProcessor(config, metrics, testLogger)
	if err != nil {
		b.Fatalf("Failed to create BlockProcessor: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = processor.GetCurrentELHeight()
	}
}
