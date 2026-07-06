package logger

import (
	"log/slog"
	"os"
)

func InitLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // logging level
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
