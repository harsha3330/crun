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

	cfg := config.Default()
	log, err := logger.New(cfg)
	if err != nil {
		panic(err)
	}

	switch os.Args[1] {
	case "init":
		if pkg.CheckPath(cfg.ConfigFilePath, false) == nil {
			log.Error("crun init has already happended")
			os.Exit(1)
		}

		if err := runtime.Init(cfg, log); err != nil {
			log.Error(fmt.Sprintf("crun init failed : %v", err.Error()))
			os.Exit(1)
		}
	case "help":
		fmt.Println("help output for crun")
	case "pull":
		fmt.Println("Pull the image")
	default:
		fmt.Printf("Unknown command : %s", os.Args[1])
	}
}
