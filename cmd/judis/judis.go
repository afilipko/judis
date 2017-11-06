package main

import (
	"judis/config"
	"judis/server"
	"os"

	log "github.com/inconshreveable/log15"
)

var envName, confPath string

func main() {
	log.Info("starting Judis...")
	log.Info(os.Getenv("JUDIS_ENV"))
	if envName = os.Getenv("JUDIS_ENV"); envName == "" {
		envName = config.DefaultEnv
	}
	log.Info("Running env " + envName)
	if confPath = os.Getenv("CONF_PATH"); confPath == "" {
		confPath = config.DefaultConfigPath
	}
	log.Info("Config path " + confPath)
	cfg := config.ParseConfig(envName, confPath)
	server := server.InitServer(cfg)

	log.Info("Started server")
	err := server.AcceptRequests()
	if err != nil {
		log.Error("Failed to start server", err)
	}
}
