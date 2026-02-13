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
		logFormatStr := initCmd.String("log-format", "", "log format of the application")
		initCmd.Parse(os.Args[2:])
		logOpts, level, format := logger.BuildLogOptions(*logLevelStr, *logFormatStr, stater)
		cfg.LogLevel = level
		cfg.LogFormat = format

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
	case "pull":
		logOpts, err := logger.GetLogOptions(cfg.ConfigFilePath)
		if err != nil {
			stater.Error("unable to get the logOptions from configfile")
			panic(err)
		}
		log, err := logger.New(logOpts)
		log.Debug("logopts", "logformat :", *logOpts.LogFormat, "loglevel :", *logOpts.LogLevel)
		if err != nil {
			stater.Error("unable to initalize the logger")
			panic(err)
		}
		stater.Success("Initialized the logger")
		err = runtime.Pull(log, stater, os.Args[2])
	case "help":
		fmt.Println("help output for crun")
	default:
		fmt.Printf("Unknown command : %s", os.Args[1])
	}
}
