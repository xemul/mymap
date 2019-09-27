package main

import (
	"os"
	"sync"
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

	f.Points = append(f.Points, &PointFile{Point: sv.Point})

	for _, area := range sv.Areas {
		af, ok := f.Areas[area.Id]
		if !ok {
			af = &AreaFile{ Area: *area }
			f.Areas[area.Id] = af
		}
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

func (s *LocalJsonStorage)Remove(id int) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	f, err := s.loadFromFile()
	if err != nil {
		return false, err
	}

	_, ok := f.Areas[id]
	if !ok {
		return false, nil
	}

	delete(f.Areas, id)
	return true, s.saveToFile(f)
}

func (s *LocalJsonStorage)loadFromFile() (*LocalJsonFile, error) {
	f, err := os.Open(localFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return &LocalJsonFile{
				Areas:	make(map[int]*AreaFile),
			}, nil
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
