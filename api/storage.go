package main

import (
	"io"
	"errors"
	"strings"
)

type DB interface {
	openUDB(*Claims) (UDB, error)
}

type UDB interface {
	Close()

	openMDB(int, bool) (MDB, error)
	List() ([]*Map, error)
	Create(*Map) (error)
	Remove(int) (error)
	PatchMap(MDB, *Map) error
}

type MDB interface {
	Id() int
	Close()

	Raw() ([]byte, error)
	Put(io.Reader) error

	SavePoint(*SaveGeoReq) error
	LoadGeos() (*LoadGeosResp, error)
	RemoveGeo(int, string) (bool, error)
	PatchPoint(int, *Point) error

	SaveVisit(int, *SaveVisitReq) error
	LoadVisits(int) (*LoadVisitsResp, error)
	RemoveVisit(int, int) (bool, error)
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
