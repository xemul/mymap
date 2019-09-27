package main

type Storage interface {
	Save(*SaveReq) error
	Load() (*LoadResp, error)
	Remove(int) (bool, error)
}

var storage = &LocalJsonStorage{}
