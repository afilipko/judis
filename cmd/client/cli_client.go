package main

import (
	"bufio"
	"fmt"
	"judis/config"
	"judis/utils"
	"net"
	"net/textproto"
	"os"
	"strconv"

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
	log.Info("to stop client enter exit or press Ctrl-C")
	scanner := bufio.NewScanner(os.Stdin)
	port, err := cfg.Cfg.Int("development.port")

	utils.LogError("cannot connect server", err)
	for scanner.Scan() {
		conn, err := net.Dial("tcp", "localhost:"+strconv.Itoa(port))
		utils.LogError("err", err)
		textProto := textproto.NewConn(conn)
		if command = scanner.Text(); command == exitCommand {
			os.Exit(1)
		} else {
			// commandName := strings.Split(command, " ")[0]
			// commands := strings.Split(command, " ")[1:]
			id, err := textProto.Cmd(command)

			if err != nil {
				log.Error("ERROR11", err)
			}
			log.Info("CMD id ", fmt.Sprint(id))
			textProto.StartResponse(id)

			code, msg, err := textProto.ReadResponse(200)
			if err != nil {
				log.Error("ERROR 222", err, msg)
			}
			log.Info("CODE " + strconv.Itoa(code) + " MSG " + msg)
			textProto.EndResponse(id)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
