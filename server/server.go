package server

import (
	"errors"
	"io"
	"judis/config"
	"judis/utils"
	"net"
	"net/textproto"
	"reflect"
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

type StorableHash struct {
	hash map[string]string
}

func (storable *StorableString) Get(args ...string) string {
	return storable.str
}

func (storable *StorableString) Set(args ...string) error {
	if len(args) == 0 {
		return errors.New("cannot use empty value")
	}
	storable.str = args[0]
	return nil
}

func (storable *StorableHash) Get(args ...string) string {
	log.Info("GET TEG")
	field := args[0]
	log.Info(" 1 Getting fiel", field)
	if storable.hash == nil {
		storable.hash = make(map[string]string)
	}
	log.Info("Getting fiel", field)
	log.Info(storable.hash[field])
	value := storable.hash[field]
	return value
}

func (storable *StorableHash) Set(args ...string) error {
	if storable.hash == nil {
		storable.hash = make(map[string]string)
	}
	field := args[0]
	value := args[1]
	storable.hash[field] = value
	log.Info("StrHsh " + storable.hash[field])
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
	store.items[key] = &storable
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

func (s *Server) hset(args []string) (string, error) {
	if len(args) != 3 {
		return "", errors.New("wrong number of arguments for 'hset' command")
	}
	store := s.storage
	store.Lock()
	defer store.Unlock()

	key := args[0]
	if store.items[key] == nil {
		store.items[key] = new(StorableHash)
	}
	error := store.items[key].Set(args[1:]...)
	if error != nil {
		return "FAIL", error
	}
	return "OK", nil
}

func (s *Server) hget(args []string) (string, error) {
	if len(args) != 2 {
		return "", errors.New("wrong number of arguments for 'hget' command")
	}
	store := s.storage
	store.RLock()
	defer store.RUnlock()

	key := args[0]
	field := args[1]
	log.Info(key)
	log.Info(field)
	if store.items[key] == nil {
		return "", nil
	}
	log.Info("Try to get field")
	hash, ok := store.items[key].(*StorableHash)
	if !ok {
		log.Info("Wrong type")
		return "", errors.New("Operation against a key holding the wrong kind of value")
	}
	value := hash.Get(field)
	return value, nil
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
	commands["HSET"] = server.hset
	commands["HGET"] = server.hget
	commands["KEYS"] = server.keys
	server.commands = commands

	return &server
}

func (server *Server) keys(args []string) (string, error) {
	if server.storage == nil {
		return "", nil
	}
	keys := reflect.ValueOf(server.storage.items).MapKeys()
	strkeys := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		strkeys[i] = keys[i].String()
	}
	return strings.Join(strkeys, " "), nil
}
func (server *Server) Keys() string {
	k := make([]string, 1)
	r, e := server.keys(k)
	if e != nil {
		return ""
	}
	return r
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
		log.Info(server.Keys())
		args := strings.Fields(l)
		command := strings.ToUpper(args[0])
		if server.commands[command] == nil {
			log.Error("not existing command")
		}
		log.Info("keys before command")
		log.Info(server.Keys())

		resp, err := server.commands[command](args[1:])

		log.Info("keys after command")
		log.Info(server.Keys())
		if err != nil {
			protoConn.PrintfLine("500-error")
			log.Error("500 Error ", err)

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
