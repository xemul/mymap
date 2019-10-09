package main

import (
	"os"
	"io"
	"log"
	"time"
	"errors"
	"strconv"
	"math/rand"
	"io/ioutil"
	"encoding/json"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	defaultMapName = "default"
)

type LocalJsonStorage struct {
	dir	string
}

func (ljs *LocalJsonStorage)openUDB(c *Claims) (UDB, error) {
	return ljs.localUDB(c.UserId), nil
}

type LocalJsonGeos struct {
	id	int
	fname	string
	s	*LocalJsonStorage
}

func (s *LocalJsonGeos)Id() int { return s.id }

func (s *LocalJsonGeos)Raw() ([]byte, error) {
	f, err := s.loadFromFile()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(f, "", "    ")
}

func (geos *LocalJsonGeos)Put(src io.Reader) error {
	to, err := ioutil.TempFile(geos.s.dir, "map.*.json")
	if err != nil {
		return err
	}

	defer func() {
		to.Close()
		if err != nil {
			os.Remove(to.Name())
		}
	}()

	rdr := io.TeeReader(&io.LimitedReader{R: src, N: 1028 * 1024}, to)

	var f GeosFile

	err = json.NewDecoder(rdr).Decode(&f)
	if err != nil {
		return err
	}

	err = os.Rename(to.Name(), geos.s.dir + "/" + geos.fname)

	return err
}

func (s *LocalJsonGeos)SavePoint(sv *SaveGeoReq) error {
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

func (s *LocalJsonGeos)PatchPoint(id int, npt *Point) error {
	f, err := s.loadFromFile()
	if err != nil {
		return err
	}

	pt, ok := f.Points[id]
	if !ok {
		return errors.New("no such point")
	}

	dirty := false

	if npt.Name != "" {
		pt.Name = npt.Name
		dirty = true
	}

	if !dirty {
		return nil
	}

	return s.saveToFile(f)
}

func (s *LocalJsonGeos)RemoveGeo(id int, typ string) (bool, error) {
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

func (geos *LocalJsonGeos)loadFromFile() (*GeosFile, error) {
	f, err := os.Open(geos.s.dir + "/" + geos.fname)
	if err != nil {
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

func (geos *LocalJsonGeos)saveToFile(fc *GeosFile) error {
	f, err := ioutil.TempFile(geos.s.dir, "map.*.json")
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

	err = os.Rename(f.Name(), geos.s.dir + "/" + geos.fname)

	return err
}

func (u *LocalUInfo)localGeos(mapid int) *LocalJsonGeos {
	return &LocalJsonGeos{
		fname: "map." + strconv.Itoa(mapid) + ".json",
		id: mapid,
		s: u.s,
	}
}

func (lui *LocalUInfo)copyGeos(id int) (*LocalJsonGeos, error) {
	geos := lui.localGeos(id)
	from, err := os.Open(geos.s.dir + "/" + geos.fname)
	if err != nil {
		return nil, err
	}

	defer from.Close()

	return lui.makeGeos(func(to io.Writer) error {
		_, err := io.Copy(to, from)
		return err
	})
}

func (lui *LocalUInfo)emptyGeos() (*LocalJsonGeos, error) {
	return lui.makeGeos(func(f io.Writer) error {
		return json.NewEncoder(f).Encode(&GeosFile{})
	})
}

func (lui *LocalUInfo)makeGeos(fill func(io.Writer) error) (*LocalJsonGeos, error) {
	for i := 0; i < 256; i++ {
		mapid := rand.Intn(1000000)
		geos := lui.localGeos(mapid)

		log.Printf("`- try geos %d\n", mapid)
		f, err := os.OpenFile(geos.s.dir + "/" + geos.fname, os.O_WRONLY | os.O_CREATE | os.O_EXCL, 0600)
		if err != nil {
			if os.IsExist(err) {
				continue
			}

			return nil, err
		}

		defer func() {
			f.Close()
			if err != nil {
				os.Remove(f.Name())
			}
		}()

		err = fill(f)
		return geos, err
	}

	return nil, errors.New("storage busy")
}

func (geos *LocalJsonGeos)Remove() error {
	return os.Remove(geos.s.dir + "/" + geos.fname)
}

func (s *LocalJsonGeos)Close() {
}

type LocalUInfo struct {
	uid	string
	s	*LocalJsonStorage
}

func (s *LocalJsonStorage)localUDB(uid string) *LocalUInfo {
	return &LocalUInfo{uid: uid, s: s}
}

func (lui *LocalUInfo)openMDB(mapid int, strict bool) (MDB, error) {
	if strict {
		if lui.uid == "" {
			return nil, errors.New("not allowed")
		}

		uf, err := lui.loadFile()
		if err != nil {
			return nil, err
		}

		_, ok := uf.Maps[mapid]
		if !ok {
			return nil, errors.New("no such map")
		}
	}

	return lui.localGeos(mapid), nil
}

func (lui *LocalUInfo)PatchMap(mp MDB, nm *Map) error {
	uf, err := lui.loadFile()
	if err != nil {
		return err
	}

	m, ok := uf.Maps[mp.Id()]
	if !ok {
		return errors.New("No such map")
	}

	dirty := false

	if nm.Name != "" {
		m.Name = nm.Name
		dirty = true
	}

	if !dirty {
		return nil
	}

	return lui.writeFile(uf)
}

func (lui *LocalUInfo)List() ([]*Map, error) {
	uf, err := lui.loadFile()
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		uf, err = lui.makeFile()
		if err != nil {
			return nil, err
		}
	}

	ret := []*Map{ &Map{} }

	for _, m := range uf.Maps {
		if (m.Name == defaultMapName) {
			*(ret[0]) = *m
		} else {
			ret = append(ret, m)
		}
	}

	return ret, nil
}

func (lui *LocalUInfo)Create(m *Map) error {
	if m.Name == defaultMapName {
		return errors.New("Name taken")
	}

	log.Printf("New map @%s\n", lui.uid)
	uf, err := lui.loadFile()
	if err != nil {
		return err
	}

	var geos *LocalJsonGeos

	if m.Copy != nil {
		log.Printf("`- copy geos @%s\n", lui.uid)
		geos, err = lui.copyGeos(*m.Copy)
		m.Copy = nil
	} else {
		log.Printf("`- make geos @%s\n", lui.uid)
		geos, err = lui.emptyGeos()
	}

	if err != nil {
		return err
	}

	m.Id = geos.id
	uf.Maps[geos.id] = m

	log.Printf("`- write to file @%s [%d]\n", lui.uid, geos.id)
	err = lui.writeFile(uf)
	if err != nil {
		geos.Remove()
	}

	return err
}

func (lui *LocalUInfo)Remove(mapid int) error {
	log.Printf("Remove map @%s.%d\n", lui.uid, mapid)
	uf, err := lui.loadFile()
	if err != nil {
		return err
	}

	_, ok := uf.Maps[mapid]
	if !ok {
		return errors.New("no such map")
	}

	delete(uf.Maps, mapid)
	err = lui.writeFile(uf)
	if err != nil {
		return err
	}

	lui.localGeos(mapid).Remove()

	return nil
}

func (lui *LocalUInfo)loadFile() (*UserFile, error) {
	if lui.uid == "" {
		return nil, errors.New("empty uid")
	}

	f, err := os.Open(lui.s.dir + "/" + lui.uid + ".json")
	if err != nil {
		return nil, err
	}

	defer f.Close()
	var uf UserFile

	err = json.NewDecoder(f).Decode(&uf)
	if err != nil {
		return nil, err
	}

	return &uf, nil
}

func (lui *LocalUInfo)makeFile() (*UserFile, error) {
	geos, err := lui.emptyGeos()
	if err != nil {
		return nil, err
	}

	uf := &UserFile{ Maps: map[int]*Map {
			geos.id: &Map{
				Id: geos.id,
				Name: defaultMapName,
			},
		},
	}

	err = lui.writeFile(uf)
	if err != nil {
		geos.Remove()
	}

	return uf, err
}


func (lui *LocalUInfo)writeFile(uf *UserFile) error {
	if lui.uid == "" {
		return errors.New("empty uid")
	}

	f, err := ioutil.TempFile(lui.s.dir, "lui.*.json")
	if err != nil {
		return err
	}

	defer func() {
		f.Close()
		if err != nil {
			os.Remove(f.Name())
		}
	}()

	err = json.NewEncoder(f).Encode(uf)
	if err != nil {
		return err
	}

	err = os.Rename(f.Name(), lui.s.dir + "/" + lui.uid + ".json")

	return err
}

func (lui *LocalUInfo)Close() {
}
