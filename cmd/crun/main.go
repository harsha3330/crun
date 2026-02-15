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
		printUsage()
		os.Exit(1)
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
			os.Exit(1)
		}

		if err := runtime.Init(&cfg, &logOpts, log, stater); err != nil {
			log.Error("crun init failed", "error", err.Error())
			stater.Error("init failed", "error", err)
			os.Exit(1)
		} else {
			stater.Success("crun init completed")
		}
	case "pull":
		logOpts, err := logger.GetLogOptions(cfg.ConfigFilePath)
		if err != nil {
			stater.Error("unable to get the logOptions from configfile")
			os.Exit(1)
		}
		log, err := logger.New(logOpts)
		log.Debug("logopts", "logformat :", *logOpts.LogFormat, "loglevel :", *logOpts.LogLevel)
		if err != nil {
			stater.Error("unable to initalize the logger")
			os.Exit(1)
		}
		stater.Success("Initialized the logger")
		err = runtime.Pull(cfg, log, stater, os.Args[2])
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
	case "run":
		runCmd := flag.NewFlagSet("run", flag.ExitOnError)
		networkHost := runCmd.Bool("network-host", false, "use host network (access UI at http://localhost)")
		if err := runCmd.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}
		if runCmd.NArg() < 1 {
			stater.Error("usage: crun run [--network-host] <image>")
			os.Exit(1)
		}
		image := runCmd.Arg(0)
		logOpts, err := logger.GetLogOptions(cfg.ConfigFilePath)
		if err != nil {
			stater.Error("unable to get the logOptions from configfile", "error", err)
			os.Exit(1)
		}
		log, err := logger.New(logOpts)
		log.Debug("logopts", "logformat :", *logOpts.LogFormat, "loglevel :", *logOpts.LogLevel)
		if err != nil {
			stater.Error("unable to initalize the logger")
			os.Exit(1)
		}
		stater.Success("Initialized the logger")
		runOpts := &runtime.RunOptions{HostNetwork: *networkHost}
		err = runtime.Run(cfg, log, stater, image, runOpts)
		if err != nil {
			log.Error(err.Error())
			stater.Error("container run failed", "error", err)
			os.Exit(1)
		}
	case "stop":
		if len(os.Args) < 3 {
			stater.Error("usage: crun stop <container-id>")
			os.Exit(1)
		}
		if err := runtime.Stop(cfg, stater, os.Args[2]); err != nil {
			stater.Error("stop failed", "error", err)
			os.Exit(1)
		}
	case "rmi":
		if len(os.Args) < 3 {
			stater.Error("usage: crun rmi <image>")
			os.Exit(1)
		}
		if err := runtime.RemoveImage(cfg, stater, os.Args[2]); err != nil {
			stater.Error("rmi failed", "error", err)
			os.Exit(1)
		}
	case "images":
		list, err := runtime.ImageList(cfg, stater)
		if err != nil {
			os.Exit(1)
		}
		if len(list) == 0 {
			fmt.Println("(no images)")
			break
		}
		for _, img := range list {
			fmt.Println(img)
		}
	case "ps":
		list, err := runtime.ContainerList(cfg, stater)
		if err != nil {
			os.Exit(1)
		}
		if len(list) == 0 {
			fmt.Println("(no running containers)")
			break
		}
		fmt.Printf("%-14s %-28s %-8s %s\n", "CONTAINER_ID", "IMAGE", "PID", "STATUS")
		for _, c := range list {
			fmt.Printf("%-14s %-28s %-8d %s\n", c.ID, c.Image, c.PID, c.Status)
		}
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: crun <command> [options] [args]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  init              Initialize crun (run once)")
	fmt.Println("  pull <image>      Pull image from registry (e.g. nginx:1-alpine-perl)")
	fmt.Println("  run [options] <image>   Run container (detached)")
	fmt.Println("    --network-host  Use host network (access at http://localhost)")
	fmt.Println("  stop <container-id>   Stop container and remove its filesystem")
	fmt.Println("  rmi <image>       Remove a pulled image")
	fmt.Println("  images            List pulled images")
	fmt.Println("  ps               List running containers")
	fmt.Println("")
	fmt.Println("Docs: see readme.md and docs/usage.md")
}
