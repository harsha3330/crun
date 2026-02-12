package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/harsha3330/crun/internal/config"
	logger "github.com/harsha3330/crun/internal/log"
	"github.com/harsha3330/crun/internal/pkg"
	"github.com/harsha3330/crun/internal/runtime"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: crun <command>")
	}

	cfg, err := config.Load("")
	if err != nil {
		panic(err)
	}
	stater := logger.Console{}

	switch os.Args[1] {
	case "init":
		err := pkg.CheckPath(cfg.ConfigFilePath, false)
		if err == nil {
			stater.Error("crun init already happended , exiting")
			os.Exit(1)
		} else {
			stater.Status("initializing crun")
		}

		initCmd := flag.NewFlagSet("init", flag.ExitOnError)
		logLevelStr := initCmd.String("log-level", "", "log level of the application")
		logFormatStr := initCmd.String("log-format", "", "log format of the application")
		initCmd.Parse(os.Args[2:])

		level := logger.LevelInfo
		format := logger.JSONLogFormat

		if *logLevelStr == "" {
			stater.Step("log level not provided, using default", "level", level)
		} else {
			l := logger.LogLevel(*logLevelStr)
			switch l {
			case logger.LevelDebug, logger.LevelInfo, logger.LevelWarn, logger.LevelError:
				level = l
				stater.Step("using log level", "level", level)
			default:
				stater.Error("invalid log level", "value", *logLevelStr)
				stater.Warn("falling back to default log level", "level", level)
			}
		}

		if *logFormatStr == "" {
			stater.Step("log format not provided, using default", "format", format)
		} else {
			f := logger.LogFormat(*logFormatStr)
			switch f {
			case logger.JSONLogFormat, logger.TextLogFormat:
				format = f
				stater.Step("using log format", "format", format)
			default:
				stater.Error("invalid log format", "value", *logFormatStr)
				stater.Warn("falling back to default log format", "format", format)
			}
		}

		levelPtr := &level
		formatPtr := &format

		cfg.LogLevel = level
		cfg.LogFomat = format

		logOpts := logger.LogOptions{
			LogLevel:  levelPtr,
			LogFormat: formatPtr,
			AppLogDir: filepath.Join(os.TempDir(), "crun"),
		}

		log, err := logger.New(&logOpts)
		if err != nil {
			stater.Error("unable to init logger", "error", err.Error())
			panic(err)
		}

		if err := runtime.Init(&cfg, &logOpts, log, stater); err != nil {
			log.Error("crun init failed", "error", err.Error())
			stater.Error("init failed")
			os.Exit(1)
		} else {
			stater.Success("crun init completed")
		}
	case "help":
		fmt.Println("help output for crun")
	case "pull":
		fmt.Println("Pull the image")
	default:
		fmt.Printf("Unknown command : %s", os.Args[1])
	}
}
