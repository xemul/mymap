package main

import (
	"os"
	"errors"
	"io/ioutil"
	"encoding/json"
)

var (
	IdNotFoundErr = errors.New("Id not found")
	IdExistsErr = errors.New("Id already exists")
)

type LocalJsonStorage struct {
	dir	string
}

func (ljs *LocalJsonStorage)Col(name string) Collection {
	return &LocalJsonCollection{
		dir:	ljs.dir,
		name:	name,
	}
}

func (ljs *LocalJsonStorage)Drop(name string) error {
	return ljs.Col(name).(*LocalJsonCollection).drop()
}

type LocalJsonCollection struct {
	dir	string
	name	string
}

type localJsonFile struct {
	Set	map[Id]json.RawMessage		`json:"set"`
}

func (lj *LocalJsonCollection)Create() error {
	f, err := os.OpenFile(lj.fname(), os.O_CREATE | os.O_EXCL | os.O_WRONLY, 0640)
	if err != nil {
		if os.IsExist(err) {
			return IdExistsErr
		}
		return err
	}

	defer func() {
		f.Close()
		if err != nil {
			os.Remove(f.Name())
		}
	}()

	err = json.NewEncoder(f).Encode(lj.empty())
	return err
}

func (lj *LocalJsonCollection)Upd(id Id, obj Obj, ucb func(Obj) error) error {
	jf, err := lj.loadFile()
	if err != nil {
		return err
	}

	data, ok := jf.Set[id]
	if !ok {
		return IdNotFoundErr
	}

	err = json.Unmarshal(data, obj)
	if err == nil {
		err = ucb(obj)
	}
	if err == nil {
		jf.Set[id], err = json.Marshal(obj)
	}

	if err != nil {
		return err
	}

	return lj.saveFile(jf)
}

func (lj *LocalJsonCollection)Add(id Id, obj Obj) error {
	jf, err := lj.loadFile()
	if err != nil {
		return err
	}

	if id == -1 {
		return errors.New("bad ID")
	}

	if _, ok := jf.Set[id]; ok {
		return IdExistsErr
	}

	jf.Set[id], err = json.Marshal(obj)
	if err != nil {
		return err
	}

	return lj.saveFile(jf)
}

func (lj *LocalJsonCollection)AddMany(og func() (Id, Obj)) error {
	jf, err := lj.loadFile()
	if err != nil {
		return err
	}

	for {
		id, obj := og()
		if id == -1 {
			break
		}

		if _, ok := jf.Set[id]; ok {
			return IdExistsErr
		}

		jf.Set[id], err = json.Marshal(obj)
		if err != nil {
			return err
		}
	}

	return lj.saveFile(jf)
}

func (lj *LocalJsonCollection)Get(id Id, obj Obj) error {
	jf, err := lj.loadFile()
	if err != nil {
		return err
	}

	data, ok := jf.Set[id]
	if !ok {
		return IdNotFoundErr
	}

	if obj == nil {
		return nil
	}

	return json.Unmarshal(data, obj)
}

func (lj *LocalJsonCollection)Del(id Id) error {
	jf, err := lj.loadFile()
	if err != nil {
		return err
	}

	if _, ok := jf.Set[id]; !ok {
		return IdNotFoundErr
	}

	delete(jf.Set, id)

	return lj.saveFile(jf)
}

func (lj *LocalJsonCollection)Iter(o Obj, fn func(id Id, o Obj) error) error {
	jf, err := lj.loadFile()
	if err != nil {
		return err
	}

	for id, data := range jf.Set {
		err = json.Unmarshal(data, o)
		if err == nil {
			err = fn(id, o)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (lj *LocalJsonCollection)Close() {
}

func (lj *LocalJsonCollection)fname() string { return lj.dir + "/" + lj.name + ".json" }

func (ljs *LocalJsonCollection)drop() error {
	return os.Remove(ljs.fname())
}

func (lf *LocalJsonCollection)empty() *localJsonFile {
	return &localJsonFile{Set: make(map[Id]json.RawMessage)}
}

func (lj *LocalJsonCollection)loadFile() (*localJsonFile, error) {
	f, err := os.Open(lj.fname())
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		return lj.empty(), nil
	}

	defer f.Close()

	var jf localJsonFile
	return &jf, json.NewDecoder(f).Decode(&jf)
}

func (lj *LocalJsonCollection)saveFile(jf *localJsonFile) error {
	f, err := ioutil.TempFile(lj.dir, lj.name + ".*.json")
	if err != nil {
		return err
	}

	defer func() {
		f.Close()
		if err != nil {
			os.Remove(f.Name())
		}
	}()

	err = json.NewEncoder(f).Encode(jf)
	if err == nil {
		err = os.Rename(f.Name(), lj.fname())
	}

	return err
}

func (lj *LocalJsonCollection)Raw() ([]byte, error) {
	jf, err := lj.loadFile()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(jf.Set, "", "    ")
}

func (lj *LocalJsonCollection)Write(data []byte) error {
	var jf localJsonFile

	err := json.Unmarshal(data, &jf.Set)
	if err != nil {
		return err
	}

	return lj.saveFile(&jf)
}
