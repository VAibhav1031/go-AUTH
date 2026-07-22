package logger

import (
	"log/slog"
	"os"
	"path/filepath"
)

func LoggerInitiator() {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error("failed to get home directory", "error", err)
		os.Exit(1)
	}

	logDir := filepath.Join(homeDir, ".local", "go-Auth", "logs")

	if err := os.MkdirAll(logDir, 0755); err != nil {
		slog.Error("failed to create log directory", "error", err)
		os.Exit(1)
	}

	logFile := filepath.Join(logDir, "unilog.log")

	file, err := os.OpenFile(logFile,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		slog.Error("failed to open log file", "error", err)
		os.Exit(1)
	}

	handler := slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
