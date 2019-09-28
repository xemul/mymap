package main

import (
	"os"
	"sync"
	"errors"
	"encoding/json"
)

const (
	localFileName	= "points.json"
)

type LocalJsonStorage struct {
	lock	sync.RWMutex
}

func (s *LocalJsonStorage)Save(sv *SaveReq) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	f, err := s.loadFromFile()
	if err != nil {
		return err
	}

	dirty := false

	if sv.Point != nil {
		if f.Points == nil {
			f.Points = make(map[int]*PointFile)
		}

		f.Points[sv.Point.Id] = &PointFile{Point: *sv.Point}
		dirty = true
	}

	for _, area := range sv.Areas {
		if f.Areas == nil {
			f.Areas = make(map[int]*AreaFile)
		}

		_, ok := f.Areas[area.Id]
		if !ok {
			f.Areas[area.Id] = &AreaFile{ Area: *area }
			dirty = true
		}
	}

	if !dirty {
		return nil
	}

	return s.saveToFile(f)
}

func (s *LocalJsonStorage)Load() (*LoadResp, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	f, err := s.loadFromFile()
	if err != nil {
		return nil, err
	}

	var ret LoadResp
	for _, area := range f.Areas {
		ret.Areas = append(ret.Areas, &area.Area)
	}

	for _, pt := range f.Points {
		ret.Points = append(ret.Points, &pt.Point)
	}

	return &ret, nil
}

func (s *LocalJsonStorage)Remove(id int, typ string) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	f, err := s.loadFromFile()
	if err != nil {
		return false, err
	}

	switch {
	case typ == "area":
		_, ok := f.Areas[id]
		if !ok {
			return false, nil
		}

		delete(f.Areas, id)
	case typ == "point":
		_, ok := f.Points[id]
		if !ok {
			return false, nil
		}

		delete(f.Points, id)
	default:
		return false, errors.New("")
	}

	return true, s.saveToFile(f)
}

func (s *LocalJsonStorage)loadFromFile() (*LocalJsonFile, error) {
	f, err := os.Open(localFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return &LocalJsonFile{ }, nil
		}

		return nil, err
	}

	defer f.Close()

	var ret LocalJsonFile

	err = json.NewDecoder(f).Decode(&ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *LocalJsonStorage)saveToFile(fc *LocalJsonFile) error {
	f, err := os.OpenFile("." + localFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer func() {
		f.Close()
		if err != nil {
			os.Remove(f.Name())
		}
	}()

	err = json.NewEncoder(f).Encode(fc)
	if err != nil {
		return err
	}

	return os.Rename(f.Name(), localFileName)
}
