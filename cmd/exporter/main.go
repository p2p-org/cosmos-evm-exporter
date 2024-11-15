package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cosmos-evm-exporter/internal/blockchain"
	"cosmos-evm-exporter/internal/config"
	"cosmos-evm-exporter/internal/logger"
	"cosmos-evm-exporter/internal/metrics"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.NewLogger(&logger.Config{
		EnableStdout: cfg.EnableStdout,
		LogFile:      cfg.LogFile,
	})

	// Initialize metrics
	blockMetrics := metrics.NewBlockMetrics()

	// Start metrics server
	metricsServer := metrics.NewServer(cfg.MetricsPort, blockMetrics.Registry)
	go func() {
		if err := metricsServer.Start(); err != nil {
			log.WriteJSONLog("error", "Metrics server failed", nil, err)
		}
	}()

	// Initialize block processor
	processor, err := blockchain.NewBlockProcessor(cfg, blockMetrics, log)
	if err != nil {
		log.WriteJSONLog("error", "Failed to create block processor", nil, err)
		os.Exit(1)
	}

	// Start metrics updater
	processor.StartMetricsUpdater(5 * time.Second)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.WriteJSONLog("info", "Shutting down...", nil, nil)
		cancel()
	}()

	// Log startup
	log.WriteJSONLog("info", "Starting exporter", map[string]interface{}{
		"metrics_port": cfg.MetricsPort,
	}, nil)

	// Start processing blocks
	processor.Start(ctx)
}
