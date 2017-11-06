package storage

import (
	"errors"
	"strconv"
	"strings"
	"sync"
)

type Storable interface {
	Get(args ...string) string
	Set(args ...string) error
}

type StorableString struct {
	Str string
}

type StorableHash struct {
	hash map[string]string
}

type StorableList struct {
	list []string
}

type TtlCleaner struct {
	DefaultTtl int
}

type Storage struct {
	sync.RWMutex
	Items      map[string]Storable
	TTLCleaner *TtlCleaner
}

func (storable *StorableString) Get(args ...string) string {
	return storable.Str
}

func (storable *StorableString) Set(args ...string) error {
	if len(args) == 0 {
		return errors.New("cannot use empty value")
	}
	storable.Str = args[0]
	return nil
}

func (storable *StorableHash) Get(args ...string) string {
	field := args[0]
	if storable.hash == nil {
		storable.hash = make(map[string]string)
	}
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
	return nil
}

func (storable *StorableList) Get(args ...string) string {
	if storable.list == nil {
		return ""
	}
	start, err := strconv.Atoi(args[0])
	if err != nil {
		start = 0
	}
	end, err := strconv.Atoi(args[1])
	if err != nil || end < len(storable.list) {
		end = len(storable.list)
	}
	if start >= end {
		return ""
	}

	return strings.Join(storable.list[start:end], " ")
}

func (storable *StorableList) Set(args ...string) error {
	if storable.list == nil {
		return errors.New("empty list")
	}
	index, err := strconv.Atoi(args[0])
	el := args[1]
	if err != nil {
		return err
	}
	if index < 0 || index >= len(storable.list) {
		return errors.New("out of range")
	}
	storable.list[index] = el
	return nil
}

func (storable *StorableList) Rpop() string {
	if storable.list == nil {
		return ""
	}
	var value string
	value, storable.list = storable.list[len(storable.list)-1], storable.list[:len(storable.list)-1]
	return value
}

func (storable *StorableList) Rpush(el string) string {
	if storable.list == nil {
		storable.list = []string{el}
		return "OK"
	}
	storable.list = append(storable.list, el)
	return "OK"
}
