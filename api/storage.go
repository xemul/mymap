package main

type Storage interface {
	SavePoint(*SavePointReq) error
	LoadGeos() (*LoadGeosResp, error)
	RemoveGeo(int, string) (bool, error)
}

var storage = &LocalJsonStorage{}
