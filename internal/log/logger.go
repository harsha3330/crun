package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/harsha3330/crun/internal/config"
	"github.com/harsha3330/crun/internal/pkg"
)

type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

type LogFormat string

const (
	JSONLogFormat LogFormat = "json"
	TextLogFormat LogFormat = "text"
)

type LogOptions struct {
	LogLevel  *LogLevel
	LogFormat *LogFormat
}

func New(cfg config.Config, logOpts *LogOptions) (*slog.Logger, error) {

	logFile := filepath.Join(cfg.AppLogDir, "crun.log")

	err := pkg.EnsureFile(logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create logfile: %s", err.Error())
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %s", err.Error())
	}

	level := parseLevel(*logOpts.LogLevel)
	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	switch *logOpts.LogFormat {
	case JSONLogFormat:
		handler = slog.NewJSONHandler(file, opts)
	case TextLogFormat:
		handler = slog.NewTextHandler(file, opts)
	default:
		handler = slog.NewJSONHandler(file, opts)
	}

	logger := slog.New(handler)

	return logger, nil
}

func parseLevel(l LogLevel) slog.Level {
	switch l {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
