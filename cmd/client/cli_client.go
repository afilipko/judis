package main

import (
	"bufio"
	"fmt"
	"judis/config"
	"os"

	log "github.com/inconshreveable/log15"
)

const exitCommand string = "exit"

var envName, confPath, command string

func main() {
	log.Info("starting CLI Judis client...")
	if envName = os.Getenv("JUDIS_ENV"); envName == "" {
		envName = config.DefaultEnv
	}
	log.Info("Running env " + envName)
	if confPath = os.Getenv("CONF_PATH"); confPath == "" {
		confPath = config.DefaultConfigPath
	}
	log.Info("Config path " + confPath)
	cfg := config.ParseConfig(envName, confPath)
	log.Info("accepting you commands ")
	log.Info("to stop client enter exit or press Ctrl-C  ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {

		if command = scanner.Text(); command == exitCommand {
			os.Exit(1)
		} else {

		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
