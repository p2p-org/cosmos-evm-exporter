package blockchain

import (
	"cosmos-evm-exporter/internal/logger"
)

func newTestLogger() *logger.Logger {
	return logger.NewLogger(&logger.Config{
		EnableStdout: false,
		LogFile:      "",
	})
}
