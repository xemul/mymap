package main

import (
	"errors"
	"strings"
)

type Id int64
type Obj interface{
	SetId(Id)
}

type Collection interface {
	Iter(Obj, func(Id, Obj) error) error
	Add(Id, Obj) (Id, error)
	AddMany(func() (Id, Obj)) error
	Get(Id, Obj) error
	Upd(Id, Obj, func(Obj) error) error
	Del(Id) error
	Raw() ([]byte, error)
	Close()
}

type DB interface {
	Col(string) Collection
	Drop(string) error
}

var storage DB

func setupStorage(db string) error {
	x := strings.SplitN(db, ":", 2)
	if len(x) != 2 {
		return errors.New("Bad storage option")
	}

	switch x[0] {
	case "ljs":
		storage = &LocalJsonStorage{dir : x[1]}
	default:
		return errors.New("Unsupported storage")
	}

	return nil
}
