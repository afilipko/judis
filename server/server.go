package server

import (
	"io"
	"judis/utils"
	"net"
	"net/textproto"
	"strings"

	log "github.com/inconshreveable/log15"
)

// Server contains server info and methods
type Server struct {
	Port int
}

func (server *Server) handle(conn *net.TCPConn) error {
	defer conn.Close()
	protoConn := textproto.NewConn(conn)

	// if err := c.PrintfLine("200 Connected"); err != nil {
	// 	log15.Warn("closing connection", "error writing to client", err)
	// 	return
	// }

	for {
		l, err := protoConn.ReadLine()
		if err != nil {
			if err == io.EOF {
				log.Info("the connection is dropped on client side")
				return err
			}
			log.Warn("closing connection", "error reading", err)
			return err
		}

		cmd := strings.Fields(l)
		log.Debug("got command", "line", l, "cmd", cmd)
	}
}

// BuildServer returns pointer to new Server instance
func BuildServer(config *Config) *Server {
	s := new(Server)
	cfg := config.Cfg
	var err error
	s.Port, err = cfg.Int(config.Env + ".port")
	utils.LogError("port must be present in config", err)
	return s
}

// func main() {

// 	// https://github.com/ivpusic/grpool

// 	var server = buildServer()

// 	addr, err := net.ResolveTCPAddr("tcp", ":8080")
// 	utils.LogError(err, "can not build address")

// 	listener, err := net.ListenTCP("tcp", addr)
// 	utils.LogError(err, "can not listen address")

// 	defer listener.Close()

// 	for {
// 		// TODO: limit connection amount
// 		connection, err := listener.AcceptTCP()
// 		utils.LogError(err, "error during connection accepting")
// 		go server.handle(connection)
// 	}
// }
