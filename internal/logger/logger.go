package logger

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"time"
)

type LogEntry struct {
	Time    string                 `json:"time"`
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type Config struct {
	EnableFileLog bool   `toml:"enable_file_log"`
	EnableStdout  bool   `toml:"enable_stdout"`
	LogFile       string `toml:"log_file"`
}

type Logger struct {
	logger *log.Logger
	config *Config
}

func NewLogger(config *Config) *Logger {
	var writer io.Writer
	if config.EnableFileLog {
		logFile, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		writer = logFile
	}

	if config.EnableStdout {
		if writer != nil {
			writer = io.MultiWriter(writer, os.Stdout)
		} else {
			writer = os.Stdout
		}
	}

	if writer == nil {
		writer = io.Discard
	}

	return &Logger{
		logger: log.New(writer, "", 0),
		config: config,
	}
}

func (l *Logger) WriteJSONLog(level string, message string, data map[string]interface{}, err error) {
	entry := LogEntry{
		Time:    time.Now().UTC().Format(time.RFC3339),
		Level:   level,
		Message: message,
		Data:    data,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	// Add emoji prefix based on level
	switch level {
	case "info":
		entry.Message = "ℹ️ " + entry.Message
	case "success":
		entry.Message = "✅ " + entry.Message
	case "warn":
		entry.Message = "⚠️ " + entry.Message
	case "error":
		entry.Message = "❌ " + entry.Message
	case "fatal":
		entry.Message = "💀 " + entry.Message
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		l.logger.Printf("Error marshaling log entry: %v", err)
		return
	}

	l.logger.Println(string(jsonBytes))
}
