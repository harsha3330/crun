package main

import (
	"flag"
	"fmt"
	"os"

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
		logFormat := initCmd.String("log-format", "", "log format of the application")
		initCmd.Parse(os.Args[2:])

		var levelPtr *logger.LogLevel
		var formatPtr *logger.LogFormat
		if *logLevelStr != "" {
			level := logger.LogLevel(*logLevelStr)

			switch level {
			case logger.LevelDebug, logger.LevelInfo, logger.LevelWarn, logger.LevelError:
				levelPtr = &level
			default:
				stater.Error("invalid log level flag", "value", levelPtr)
				os.Exit(1)
			}
		}

		if *logFormat != "" {
			format := logger.LogFormat(*logFormat)

			switch format {
			case logger.JSONLogFormat, logger.TextLogFormat:
				formatPtr = &format
			default:
				stater.Error("invalid log level format", "value", logFormat)
				os.Exit(1)
			}
		}

		logOpts := logger.LogOptions{
			LogLevel:  levelPtr,
			LogFormat: formatPtr,
		}

		log, err := logger.New(cfg, &logOpts)
		if err != nil {
			stater.Error("unable to init logger", "error", err.Error())
			panic(err)
		}

		if err := runtime.Init(&cfg, log, stater); err != nil {
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
