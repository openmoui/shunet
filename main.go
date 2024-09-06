package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/context"
	"os"
	"os/signal"
	"shunet/config"
	"shunet/shuclient"
	"shunet/utils"
	"syscall"
)

func usage() {
	fmt.Println(`SHU Net is a tool designed to maintain the network connections of Shanghai University. 
Usage: 
	shunet
Options:`)
	flag.PrintDefaults()
}

var (
	log        = utils.Log
	stop       = flag.Bool("stop", false, "stop connect school network, and kill running process")
	configPath = `config.yaml`
	ctx        = context.Background()
)

func main() {
	flag.Usage = usage
	flag.Parse() // 默认有个help参数

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if *stop {
		// kill running process
		utils.Kill(cfg.Pid)
		return
	}

	runCtx, cancel := context.WithCancel(ctx)
	go ListenSignal(cancel)

	client := shuclient.NewClient(cfg)
	client.Run(runCtx)
}

// ListenSignal to stop process
func ListenSignal(cf context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Info("ListenSignal Receive stop signal, shunet exit")
	cf()
}
