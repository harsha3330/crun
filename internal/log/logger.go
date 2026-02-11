package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/harsha3330/crun/internal/config"
)

func New(cfg config.Config) (*slog.Logger, error) {
	logFile := filepath.Join(cfg.AppLogDir, "crun.log")

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	writer := io.MultiWriter(file, os.Stdout)

	level := parseLevel(cfg.LogLevel)

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	switch cfg.LogFormat {
	case config.JSONLogFormat:
		handler = slog.NewJSONHandler(writer, opts)
	case config.TextLogFormat:
		handler = slog.NewTextHandler(writer, opts)
	default:
		handler = slog.NewJSONHandler(writer, opts)
	}

	logger := slog.New(handler)

	return logger, nil
}

func parseLevel(l config.LogLevel) slog.Level {
	switch l {
	case config.LevelDebug:
		return slog.LevelDebug
	case config.LevelInfo:
		return slog.LevelInfo
	case config.LevelWarn:
		return slog.LevelWarn
	case config.LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
