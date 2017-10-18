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
	// serv := server.BuildServer(cfg)
	// log.Info("PORT " + strconv.Itoa(serv.Port))
	p, err := cfg.Cfg.Int("development.port")
	if err != nil {
		log.Error("4", err)
	}
	s := server.Server{Port: p}
	s.Handle()
}
