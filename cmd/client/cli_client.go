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
	"strings"

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
	utils.LogError("err", err)
	conn, err := net.Dial("tcp", "localhost:"+strconv.Itoa(port))
	textProto := textproto.NewConn(conn)

	utils.LogError("cannot connect server", err)
	for scanner.Scan() {

		if command = scanner.Text(); command == exitCommand {
			os.Exit(1)
		} else {
			commandName := strings.Split(command, " ")[0]
			commands := strings.Split(command, " ")[1:]
			id, err := textProto.Cmd(commandName, commands)

			if err != nil {
				log.Error("m", err)
			}
			textProto.StartResponse(id)
			defer textProto.EndResponse(id)
			code, msg, err := textProto.ReadResponse(333)
			if err != nil {
				log.Error("m", err)
			}
			log.Info("CODE " + strconv.Itoa(code) + " MSG " + msg)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
