package main

import (
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
		if pkg.CheckPath(cfg.ConfigFilePath, false) == nil {
			log.Error("crun init has already happended")
			stater.Error("crun init has already happended")
			os.Exit(1)
		}

		if err := runtime.Init(cfg, log, stater); err != nil {
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
