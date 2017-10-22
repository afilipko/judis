package server

import (
	"io"
	"judis/config"
	"judis/utils"
	"net"
	"net/textproto"
	"strings"
	"sync"

	log "github.com/inconshreveable/log15"
)

// Server contains server info and methods
type Server struct {
	conf  *config.Config
	cache *Cache
	commands map[string]Command
}

type TtlCleaner struct {
	defaultTtl int
}

type Cache struct {
	sync.Mutex
	items      map[string]interface{}
	ttlCleaner *TtlCleaner
}

func InitServer(config *config.Config) *Server {
	ttlCleaner := new(TtlCleaner)
	cache := new(Cache)

	ttlCleaner.defaultTtl = config.DefaultTTL()
	cache.ttlCleaner = ttlCleaner
	cache.items = make(map[string]interface{})

	server := Server{conf: config, cache: cache}
	return &server
}

func (server *Server) AcceptRequests() error {
	addr, err := net.ResolveTCPAddr("tcp", ":3002")
	utils.LogError("can not build address", err)

	listener, err := net.ListenTCP("tcp", addr)
	utils.LogError("can not listen address", err)

	defer listener.Close()

	for {
		// TODO: limit connection amount
		connection, err := listener.AcceptTCP()
		utils.LogError("error during connection accepting", err)
		defer connection.Close()
		protoConn := textproto.NewConn(connection)

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
	// go server.handle(connection)

}

// // BuildServer returns pointer to new Server instance
// func BuildServer(config *Config) *Server {
// 	s := new(Server)
// 	cfg := config.Cfg
// 	var err error
// 	s.Port, err = cfg.Int(config.Env + ".port")
// 	utils.LogError("port must be present in config", err)
// 	return s
// }

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
