package server

import (
	"net/textproto"
)

type dispatcher struct {
}

func CommandDispatcher() *dispatcher {
	return new(dispatcher)
}

type Command func(args []string, conn *textproto.Conn) error
