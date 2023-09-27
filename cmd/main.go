package main

import (
	"flag"
	"net-capture/pkg/emitter"
	"net-capture/pkg/logger"
	"net-capture/pkg/plugin"
	"net-capture/pkg/util"
	"os"
	"os/signal"
	"syscall"
)

var ConfigFile string

func main() {
	flag.Parse()
	checkConfigFile()

	config, err := util.GetConfig(ConfigFile)
	if err != nil {
		logger.Fatal(err, "Process config file error")
	}

	if config.DebugMode {
		logger.SetGlobalLogLevel(logger.DEBUG)
	}

	plugins := plugin.InitPlugins(config.Input)

	e := emitter.NewEmitter()
	go e.Start(plugins)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	e.Close()
	logger.Info("Shutdown Server")
	os.Exit(1)
}

func checkConfigFile() {
	if ConfigFile == "" {
		logger.Fatal(nil, "config file not specify")
	}
}

func init() {
	flag.StringVar(&ConfigFile, "config-file", "", "Specify config from yml file")
}
