package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"runtime"
	"trace-monitor-collector/config"
)

var (
	configPathFlag           = flag.String("config", "", "path to config file")
	version                  = flag.Bool("version", false, "package version")
	IsVerbose                = flag.Bool("v", false, "Verbose mode")
	IsVeryVerbose            = flag.Bool("vv", false, "Very Verbose mode")
	IsVeryVeryVerbose        = flag.Bool("vvv", false, "Very Very Verbose mode")
	AppVersion        string = "dev"
)

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("trace-monitor-collector: %s\n", AppVersion)
		os.Exit(0)
	}
	fmt.Printf("Version app: %s\n", AppVersion)

	configPath := *configPathFlag
	if configPath == "" {
		if *IsVerbose || *IsVeryVerbose || *IsVeryVeryVerbose {
			log.Println("--config flag is not set, trying to find config.yaml in current folder")
		}
		configPath = "config.yaml"
	}
	_, err := os.Stat(configPath)
	if errors.Is(err, os.ErrNotExist) {
		log.Fatalf("Could not find a config file at \"%s\"", configPath)
	}
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		log.Fatal(err)
	}
	cfg.SetVerbosity(*IsVerbose, *IsVeryVerbose, *IsVeryVeryVerbose)

	runtime.SetBlockProfileRate(1)

	ctx, _ := context.WithCancel(context.Background())

	go handleUdp(ctx, cfg)

	go handleFpmStatus(cfg)

	go handleHttp(cfg)

	go handlePrometheus(cfg)

	<-make(chan struct{})
}
