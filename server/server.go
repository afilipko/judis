package server

import (
	"errors"
	"io"
	"judis/config"
	"judis/storage"
	"judis/utils"
	"net"
	"net/textproto"
	"reflect"
	"runtime"
	"strings"

	log "github.com/inconshreveable/log15"
	"github.com/jeffail/tunny"
)

// Server contains server info and methods
type Server struct {
	conf     *config.Config
	storage  *storage.Storage
	commands map[string]Handler
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

func (s *Server) rpush(args []string) (string, error) {
	if len(args) != 2 {
		return "", errors.New("wrong number of arguments for 'rpush' command")
	}
	store := s.storage
	store.Lock()
	defer store.Unlock()

	key := args[0]
	if store.items[key] == nil {
		store.items[key] = new(StorableList)
	}
	list, ok := store.items[key].(*StorableList)
	if !ok {
		log.Info("Wrong type")
		return "", errors.New("Operation against a key holding the wrong kind of value")
	}
	return list.Rpush(args[1]), nil
}

func (s *Server) rpop(args []string) (string, error) {
	if len(args) != 1 {
		return "", errors.New("wrong number of arguments for 'rpush' command")
	}
	store := s.storage
	store.RLock()
	defer store.RUnlock()

	key := args[0]
	if store.items[key] == nil {
		return "", nil
	}

	list, ok := store.items[key].(*StorableList)
	if !ok {
		log.Info("Wrong type")
		return "", errors.New("Operation against a key holding the wrong kind of value")
	}
	return list.Rpop(), nil
}

func (s *Server) lrange(args []string) (string, error) {
	if len(args) != 3 {
		return "", errors.New("wrong number of arguments for 'lrange' command")
	}
	store := s.storage
	store.RLock()
	defer store.RUnlock()

	key := args[0]
	if store.items[key] == nil {
		return "(empty list or set)", nil
	}

	list, ok := store.items[key].(*StorableList)
	if !ok {
		log.Info("Wrong type")
		return "", errors.New("Operation against a key holding the wrong kind of value")
	}
	return list.Get(args[1], args[2]), nil
}

func (s *Server) lset(args []string) (string, error) {
	if len(args) != 3 {
		return "", errors.New("wrong number of arguments for 'rpush' command")
	}
	store := s.storage
	store.Lock()
	defer store.Unlock()

	key := args[0]
	if store.items[key] == nil {
		store.items[key] = new(StorableList)
	}
	list, ok := store.items[key].(*StorableList)
	if !ok {
		log.Info("Wrong type")
		return "", errors.New("Operation against a key holding the wrong kind of value")
	}
	err := list.Set(args[1], args[2])
	if err != nil {
		return "FAIL", err
	}
	return "OK", nil
}

func (s *Server) del(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("wrong number of arguments for 'rpush' command")
	}
	store := s.storage
	store.Lock()
	defer store.Unlock()
	for _, key := range args {
		delete(store.items, key)
	}

	return "OK", nil
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
	commands["RPUSH"] = server.rpush
	commands["RPOP"] = server.rpop
	commands["LRANGE"] = server.lrange
	commands["LSET"] = server.lset
	commands["DEL"] = server.del

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

func (server *Server) handleConnection(conn *net.TCPConn) {
	defer conn.Close()
	protoConn := textproto.NewConn(conn)

	l, err := protoConn.ReadLine()
	if err != nil {
		if err == io.EOF {
			log.Info("the connection is dropped on client side")
			// return err
		}
		log.Warn("closing connection", "error reading", err)
		// return err
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
	protoConn.PrintfLine("200-" + resp)
	protoConn.PrintfLine("200 ")
	log.Info("Close conn")
}

func (server *Server) AcceptRequests() error {
	addr, err := net.ResolveTCPAddr("tcp", ":3002")
	utils.LogError("can not build address", err)

	listener, err := net.ListenTCP("tcp", addr)
	utils.LogError("can not listen address", err)

	pool := server.initPool()
	pool, err = pool.Open()
	utils.LogError("can not init workers pool", err)

	defer listener.Close()
	defer pool.Close()

	for {
		connection, err := listener.AcceptTCP()
		utils.LogError("error during connection accepting", err)
		pool.SendWork(connection)
	}

}

func (server *Server) initPool() *tunny.WorkPool {
	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs + 1) // numCPUs hot threads + one for async tasks.

	pool := tunny.CreatePool(numCPUs, func(object interface{}) interface{} {
		conn := object.(*net.TCPConn)
		server.handleConnection(conn)
		return nil
	})
	return pool
}
