package server

import (
	"errors"
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
	conf     *config.Config
	storage  *Storage
	commands map[string]Handler
}

type TtlCleaner struct {
	defaultTtl int
}

type Storable interface {
	Get(args ...string) string
	Set(args ...string) error
	// Update(args ...string) error
	// Remove(args ...string) string
}

type StorableString struct {
	str string
}

func (storable StorableString) Get(args ...string) string {
	return storable.str
}

func (storable StorableString) Set(args ...string) error {
	if len(args) == 0 {
		return errors.New("cannot use empty value")
	}
	storable.str = args[0]
	return nil
}

type Storage struct {
	sync.RWMutex
	items      map[string]Storable
	ttlCleaner *TtlCleaner
}

type Handler func(args []string) (string, error)

func (s *Server) set(args []string) (string, error) {
	if len(args) > 2 {
		return "", errors.New("wrong number of arguments for 'get' command")
	}
	store := s.storage
	store.Lock()
	defer store.Unlock()
	key := args[0]
	value := args[1]
	if store.items[key] != nil {
		return "OK", store.items[key].Set(value)
	}
	storable := StorableString{str: value}
	store.items[key] = storable
	return "OK", nil
}

func (s *Server) get(args []string) (string, error) {
	if len(args) > 1 {
		return "", errors.New("wrong number of arguments for 'get' command")
	}
	s.storage.RLock()
	defer s.storage.RUnlock()
	key := args[0]
	if s.storage.items[key] == nil {
		return "", nil
	}
	return s.storage.items[key].Get(), nil
}

func InitServer(config *config.Config) *Server {
	ttlCleaner := new(TtlCleaner)
	storage := new(Storage)

	ttlCleaner.defaultTtl = config.DefaultTTL()
	storage.ttlCleaner = ttlCleaner
	storage.items = make(map[string]Storable)

	server := Server{conf: config, storage: storage}

	commands := make(map[string]Handler)
	commands["GET"] = server.get
	commands["SET"] = server.set
	server.commands = commands

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
		log.Info("ACCEPt CONN")
		protoConn := textproto.NewConn(connection)

		// if err := c.PrintfLine("200 Connected"); err != nil {
		// 	log15.Warn("closing connection", "error writing to client", err)
		// 	return
		// }
		//
		// for {
		l, err := protoConn.ReadLine()
		if err != nil {
			if err == io.EOF {
				log.Info("the connection is dropped on client side")
				return err
			}
			log.Warn("closing connection", "error reading", err)
			return err
		}
		log.Info(l)
		args := strings.Fields(l)

		if server.commands[args[0]] == nil {
			log.Error("not existing command")
		}
		resp, err := server.commands[args[0]](args[1:])
		if err != nil {
			protoConn.PrintfLine("500-error")
			log.Error("Error", err)

		}
		log.Info("Started response...")
		log.Info(resp)
		protoConn.PrintfLine("200-" + resp)
		protoConn.PrintfLine("200 ")
		log.Info("Close conn")
		connection.Close()
		// }
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
