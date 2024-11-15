package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cosmos-evm-exporter/internal/blockchain"
	"cosmos-evm-exporter/internal/config"
	"cosmos-evm-exporter/internal/logger"
	"cosmos-evm-exporter/internal/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// Initialize metrics and register with default prometheus handler
	metrics := metrics.NewBlockMetrics()
	http.Handle("/metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}))

	// Start metrics server
	go func() {
		if err := http.ListenAndServe(cfg.MetricsPort, nil); err != nil {
			log.WriteJSONLog("error", "Failed to start metrics server", nil, err)
			os.Exit(1)
		}
	}()

	// Initialize block processor
	processor, err := blockchain.NewBlockProcessor(cfg, metrics, log)
	if err != nil {
		log.WriteJSONLog("error", "Failed to create block processor", nil, err)
		os.Exit(1)
	}

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

	// Start metrics updater
	processor.StartMetricsUpdater(ctx, 5*time.Second)

	// Log startup
	log.WriteJSONLog("info", "Starting exporter", map[string]interface{}{
		"metrics_port": cfg.MetricsPort,
	}, nil)

	// Start processing blocks
	processor.Start(ctx)
}
