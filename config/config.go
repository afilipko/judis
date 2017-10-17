package config

import (
	"judis/utils"

	conf "github.com/olebedev/config"
)

// DefaultEnv default environment for server and cli client
const DefaultEnv string = "development"

// DefaultConfigPath default conf file path for server and cli client
const DefaultConfigPath string = "conf.yml"

// Config contains env name and *config.Config
type Config struct {
	Env string
	Cfg *conf.Config
}

// ParseConfig read conf from CONF_PATH
func ParseConfig(envName string, confPath string) *Config {
	cfg, err := conf.ParseYamlFile(confPath)
	utils.LogErrorAndExit("Can't read config file", err)
	serverCfg := Config{Env: envName, Cfg: cfg}
	return &serverCfg
}
