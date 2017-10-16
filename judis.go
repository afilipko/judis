package main

import (
	"judis/server"
	"judis/utils"
	"os"
	"strconv"

	log "github.com/inconshreveable/log15"
	"github.com/olebedev/config"
)

const defaultEnv string = "development"
const defaultConfigPath string = "conf.yml"

var envName, confPath string

func main() {
	log.Info("starting Judis...")
	log.Info(os.Getenv("JUDIS_ENV"))
	if envName = os.Getenv("JUDIS_ENV"); envName == "" {
		envName = defaultEnv
	}
	if confPath = os.Getenv("CONF_PATH"); confPath == "" {
		confPath = defaultConfigPath
	}
	log.Info("Config path " + confPath)
	log.Info("Running env " + envName)

	cfg, err := config.ParseYamlFile(confPath)
	utils.LogErrorAndExit("Can't read config file", err)

	serverConfig := server.Config{Env: envName, Cfg: cfg}
	serv := server.BuildServer(&serverConfig)
	log.Info("PORT " + strconv.Itoa(serv.Port))
}
