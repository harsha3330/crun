package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/harsha3330/crun/internal/pkg"
	"github.com/pelletier/go-toml/v2"
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
	AppLogDir string
}

func New(logOpts *LogOptions) (*slog.Logger, error) {

	logFile := filepath.Join(logOpts.AppLogDir, "crun.log")

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

func BuildLogOptions(
	logLevelStr string,
	logFormatStr string,
	stater Console,
) (LogOptions, LogLevel, LogFormat) {

	level := LevelInfo
	format := JSONLogFormat

	if logLevelStr == "" {
		stater.Step("log level not provided, using default", "level", level)
	} else {
		l := LogLevel(logLevelStr)
		switch l {
		case LevelDebug, LevelInfo, LevelWarn, LevelError:
			level = l
			stater.Step("using log level", "level", level)
		default:
			stater.Error("invalid log level", "value", logLevelStr)
			stater.Warn("falling back to default log level", "level", level)
		}
	}
	if logFormatStr == "" {
		stater.Step("log format not provided, using default", "format", format)
	} else {
		f := LogFormat(logFormatStr)
		switch f {
		case JSONLogFormat, TextLogFormat:
			format = f
			stater.Step("using log format", "format", format)
		default:
			stater.Error("invalid log format", "value", logFormatStr)
			stater.Warn("falling back to default log format", "format", format)
		}
	}

	opts := LogOptions{
		LogLevel:  &level,
		LogFormat: &format,
		AppLogDir: filepath.Join(os.TempDir(), "crun"),
	}

	return opts, level, format
}

func GetLogOptions(tomlFilePath string) (*LogOptions, error) {
	data, err := os.ReadFile(tomlFilePath)
	if err != nil {
		return nil, err
	}

	var cfg struct {
		LogFormat LogFormat `toml:"logFormat"`
		LogLevel  LogLevel  `toml:"logLevel"`
		AppLogDir string    `toml:"appLogDir"`
	}

	err = toml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &LogOptions{
		LogFormat: &cfg.LogFormat,
		LogLevel:  &cfg.LogLevel,
		AppLogDir: cfg.AppLogDir,
	}, nil
}
