package logger

import (
	"log/slog"
	"os"
)

func LoggerInitiator() {

	homeDir, err := os.UserHomeDir()
	if err != nil {

		slog.Error("Failed to open the log file", "error", err)
		os.Exit(1)

	}

	// Create the directory if not exist
	logFolderPath := homeDir + "/.local/go-Auth"
	// i want to make sure that this should have the mkdir -p kind of the behaviour thing
	err = os.MkdirAll(logFolderPath, 0755)
	if err != os.ErrExist || err != nil {
		slog.Error("Failed in the creation of the folder")
		os.Exit(1)
	}

	logFilepath := homeDir + "/.local/go-Auth/logs/unilog.log"
	file, err := os.OpenFile(logFilepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Error("Failed to open the log file", "error", err)
		os.Exit(1)

	}

	handler := slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
