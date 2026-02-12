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
	log, err := logger.New(cfg)
	stater := logger.Console{}
	if err != nil {
		panic(err)
	}

	switch os.Args[1] {
	case "init":
		err := pkg.CheckPath(cfg.ConfigFilePath, false)
		if err == nil {
			log.Error("crun config file already present")
			stater.Error("crun init already happended , exiting")
			os.Exit(1)
		} else {
			stater.Status("initializing crun")
		}

		initCmd := flag.NewFlagSet("init", flag.ExitOnError)
		logLevelStr := initCmd.String("log-level", "", "log level of the application")
		initCmd.Parse(os.Args[2:])

		var levelPtr *config.LogLevel

		if *logLevelStr != "" {
			level := config.LogLevel(*logLevelStr)

			switch level {
			case config.LevelDebug, config.LevelInfo, config.LevelWarn, config.LevelError:
				levelPtr = &level
			default:
				log.Error("invalid log level", "logLevel", *logLevelStr)
				stater.Error("invalid log level flag")
				os.Exit(1)
			}
		}

		opts := runtime.InitOptions{
			LogLevel: levelPtr,
		}

		if err := runtime.Init(&cfg, log, stater, &opts); err != nil {
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
