package server

import (
	"io"
	"net"
	"net/textproto"
	"strings"

	log "github.com/inconshreveable/log15"
)

type Server struct {
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

func buildServer() *Server {
	return new(Server)
}

func logError(err error, msg string) {
	if err != nil {
		log.Crit(msg, err)
	}
}
func main() {
	// github.com/olebedev/config CONFIG
	// https://github.com/ivpusic/grpool
	log.Info("starting server...")
	var server = buildServer()

	addr, err := net.ResolveTCPAddr("tcp", ":8080")
	logError(err, "can not build address")

	listener, err := net.ListenTCP("tcp", addr)
	logError(err, "can not listen address")

	defer listener.Close()

	for {
		// TODO: limit connection amount
		connection, err := listener.AcceptTCP()
		logError(err, "error during connection accepting")
		go server.handle(connection)
	}
}
