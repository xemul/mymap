package main

type Storage interface {
	Save(*SaveReq) error
	Load() (*LoadResp, error)
	Remove(int, string) (bool, error)
}

var storage = &LocalJsonStorage{}
