package main

import (
	"os"
	"sync"
	"errors"
	"encoding/json"
)

type LocalJsonGeos struct {
	name	string
	lock	sync.RWMutex
}

func (s *LocalJsonGeos)filename() string {
	return s.name + ".data.json"
}

func (s *LocalJsonGeos)SavePoint(sv *SaveGeoReq) error {
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

func (s *LocalJsonGeos)LoadGeos() (*LoadGeosResp, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	f, err := s.loadFromFile()
	if err != nil {
		return nil, err
	}

	var ret LoadGeosResp
	for _, area := range f.Areas {
		ret.Areas = append(ret.Areas, &area.Area)
	}

	for _, pt := range f.Points {
		ret.Points = append(ret.Points, &pt.Point)
	}

	return &ret, nil
}

func (s *LocalJsonGeos)RemoveGeo(id int, typ string) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	f, err := s.loadFromFile()
	if err != nil {
		return true, err
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

func (s *LocalJsonGeos)SaveVisit(pid int, sv *SaveVisitReq) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	f, err := s.loadFromFile()
	if err != nil {
		return err
	}

	pt, ok := f.Points[pid]
	if !ok {
		return errors.New("no such point")
	}

	pt.Visits = append(pt.Visits, &sv.Visit)
	return s.saveToFile(f)
}

func (s *LocalJsonGeos)LoadVisits(pid int) (*LoadVisitsResp, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	f, err := s.loadFromFile()
	if err != nil {
		return nil, err
	}

	if pid != -1 {
		pt, ok := f.Points[pid]
		if !ok {
			return nil, errors.New("no such point")
		}

		a := pt.Visits
		if a == nil {
			a = []*Visit{}
		}

		return &LoadVisitsResp{A: a}, nil
	}

	ret := &LoadVisitsResp{}

	for _, pt := range f.Points {
		id := pt.Id
		for _, v := range pt.Visits {
			vis := *v
			vis.PId = &id
			ret.A = append(ret.A, &vis)
		}
	}

	return ret, nil
}

func (s *LocalJsonGeos)RemoveVisit(pid, vn int) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	f, err := s.loadFromFile()
	if err != nil {
		return true, err
	}

	pt, ok := f.Points[pid]
	if !ok {
		return false, errors.New("no such point")
	}

	va := pt.Visits
	if vn >= len(va) {
		return false, nil
	}

	pt.Visits = append(va[:vn], va[vn+1:]...)
	return true, s.saveToFile(f)
}

func (s *LocalJsonGeos)loadFromFile() (*GeosFile, error) {
	f, err := os.Open(s.filename())
	if err != nil {
		if os.IsNotExist(err) {
			return &GeosFile{ }, nil
		}

		return nil, err
	}

	defer f.Close()

	var ret GeosFile

	err = json.NewDecoder(f).Decode(&ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *LocalJsonGeos)saveToFile(fc *GeosFile) error {
	f, err := os.OpenFile("." + s.filename(), os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0600)
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

	return os.Rename(f.Name(), s.filename())
}
