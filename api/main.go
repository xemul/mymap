package main

import (
	"os"
	"log"
	"flag"
	"sync"
	"errors"
	"strconv"
	"net/http"
	"encoding/json"
	"encoding/base64"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"github.com/dgrijalva/jwt-go"
)

var noValue = errors.New("no value")

func qId(pn string) (Id, error) {
	v, err := qInt(pn)
	return Id(v), err
}

func qInt(pn string) (int, error) {
	if pn == "" {
		return -1, noValue
	}

	return strconv.Atoi(pn)
}

type Claims struct {
	*jwt.StandardClaims
	UserId	string		`json:"id"`
}

func (c *Claims)mapsCol() string { return "maps." + c.UserId }
func (c *Claims)pointsCol(mapid Id) string { return "points." + strconv.Itoa(int(mapid)) }
func (c *Claims)areasCol(mapid Id) string { return "areas." + strconv.Itoa(int(mapid)) }

func (c *Claims)checkMapCol(mapid Id) bool {
	mcol := storage.Col(c.mapsCol())
	defer mcol.Close()

	err := mcol.Get(mapid, nil)
	return err == nil
}

func handleListGeos(c *Claims, mapid Id, w http.ResponseWriter, r *http.Request) {
	var geos LoadGeosResp
	var lw sync.WaitGroup

	lw.Add(2)

	var perr error
	go func() {
		pcol := storage.Col(c.pointsCol(mapid))
		defer pcol.Close()

		perr = pcol.Iter(&Point{}, func(id Id, x Obj) error {
			pt := *(x.(*Point))
			geos.Points = append(geos.Points, &pt)
			return nil
		})

		lw.Done()
	}()

	var aerr error
	go func() {
		acol := storage.Col(c.areasCol(mapid))
		defer acol.Close()

		aerr = acol.Iter(&Area{}, func(id Id, x Obj) error {
			a := *(x.(*Area))
			geos.Areas = append(geos.Areas, &a)
			return nil
		})

		lw.Done()
	}()

	lw.Wait()
	err := aerr
	if err == nil {
		err = perr
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&geos)
}

func handleSaveGeos(c *Claims, mapid Id, w http.ResponseWriter, r *http.Request) {
	var sv SaveGeoReq
	var sw sync.WaitGroup

	if !c.checkMapCol(mapid) {
		http.Error(w, "no such map", http.StatusNotFound)
		return
	}

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&sv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var perr error
	if sv.Point != nil {
		sw.Add(1)
		pt := sv.Point
		go func() {
			pcol := storage.Col(c.pointsCol(mapid))
			defer pcol.Close()
			_, perr = pcol.Add(pt.Id, pt)
			sw.Done()
		}()
	}

	var aerr error
	if len(sv.Areas) > 0 {
		sw.Add(1)
		go func() {
			i := -1
			acol := storage.Col(c.areasCol(mapid))
			defer acol.Close()
			aerr = acol.AddMany(func() (Id, Obj) {
				i++
				if i >= len(sv.Areas) {
					return -1, nil
				}
				a := sv.Areas[i]
				return a.Id, a
			})
			sw.Done()
		}()
	}

	sw.Wait()
	err = perr
	if err == nil {
		err = aerr
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleRemovePoint(c *Claims, mapid, pid Id, w http.ResponseWriter) {
	if !c.checkMapCol(mapid) {
		http.Error(w, "no such map", http.StatusNotFound)
		return
	}

	pcol := storage.Col(c.pointsCol(mapid))
	defer pcol.Close()

	err := pcol.Del(pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleUpdatePoint(c *Claims, mapid, pid Id, w http.ResponseWriter, r *http.Request) {
	if !c.checkMapCol(mapid) {
		http.Error(w, "no such map", http.StatusNotFound)
		return
	}

	var npt Point

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&npt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pcol := storage.Col(c.pointsCol(mapid))
	defer pcol.Close()

	err = pcol.Upd(pid, &PointX{}, func(o Obj) error {
		pt := o.(*PointX)

		if npt.Name != "" {
			pt.Name = npt.Name
		}

		if npt.Area != 0 {
			pt.Area = npt.Area
			pt.Lat = npt.Lat
			pt.Lng = npt.Lng
			pt.Countries = npt.Countries
		}

		return nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleMapPoint(c *Claims, w http.ResponseWriter, r *http.Request) {
	mapid, err := qId(mux.Vars(r)["mapid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	pid, err := qId(mux.Vars(r)["pid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch r.Method {
	case "PATCH":
		handleUpdatePoint(c, mapid, pid, w, r)
	case "DELETE":
		handleRemovePoint(c, mapid, pid, w)
	}
}

func handleMapArea(c *Claims, w http.ResponseWriter, r *http.Request) {
	mapid, err := qId(mux.Vars(r)["mapid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	aid, err := qId(mux.Vars(r)["aid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch r.Method {
	case "DELETE":
		handleRemoveArea(c, mapid, aid, w)
	}
}

func handleRemoveArea(c *Claims, mapid, aid Id, w http.ResponseWriter) {
	if !c.checkMapCol(mapid) {
		http.Error(w, "no such map", http.StatusNotFound)
		return
	}

	acol := storage.Col(c.areasCol(mapid))
	defer acol.Close()

	err := acol.Del(aid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleMapGeos(c *Claims, w http.ResponseWriter, r *http.Request) {
	mapid, err := qId(mux.Vars(r)["mapid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		handleListGeos(c, mapid, w, r)
	case "POST":
		handleSaveGeos(c, mapid, w, r)
	}
}

func handleListVisits(c *Claims, mapid, pid Id, w http.ResponseWriter) {
	var pts []*PointX

	pcol := storage.Col(c.pointsCol(mapid))
	defer pcol.Close()

	var err error

	if pid == -1 {
		var pt PointX
		err = pcol.Get(pid, &pt)
		pts = append(pts, &pt)
	} else {
		err = pcol.Iter(&PointX{}, func(id Id, x Obj) error {
			pt := *(x.(*PointX))
			pts = append(pts, &pt)
			return nil
		})
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var viss LoadVisitsResp
	for _, pt := range pts {
		viss.A = append(viss.A, pt.Vis...)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&viss)
}

func handleSaveVisit(c *Claims, mapid, pid Id, w http.ResponseWriter, r *http.Request) {
	if !c.checkMapCol(mapid) {
		http.Error(w, "no such map", http.StatusNotFound)
		return
	}

	var sv Visit

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&sv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pcol := storage.Col(c.pointsCol(mapid))
	defer pcol.Close()

	err = pcol.Upd(pid, &PointX{}, func(x Obj) error {
		pt := x.(*PointX)
		pt.Vis = append(pt.Vis, &sv)
		return nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handlePointVisit(c *Claims, w http.ResponseWriter, r *http.Request) {
	mapid, err := qId(mux.Vars(r)["mapid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	pid, err := qId(mux.Vars(r)["pid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	vn, err := qInt(mux.Vars(r)["vn"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch r.Method {
	case "DELETE":
		handleDeleteVisit(c, mapid, pid, vn, w)
	}
}

func handleDeleteVisit(c *Claims, mapid, pid Id, vn int, w http.ResponseWriter) {
	if !c.checkMapCol(mapid) {
		http.Error(w, "no such map", http.StatusNotFound)
		return
	}

	pcol := storage.Col(c.pointsCol(mapid))
	defer pcol.Close()

	err := pcol.Upd(pid, &PointX{}, func(o Obj) error {
		pt := o.(*PointX)
		va := pt.Vis

		if vn >= len(va) {
			return errors.New("no such visit")
		}

		pt.Vis = append(va[:vn], va[vn+1:]...)
		return nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleMapVisits(c *Claims, w http.ResponseWriter, r *http.Request) {
	mapid, err := qId(mux.Vars(r)["mapid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		handleListVisits(c, mapid, -1, w)
	}
}

func handlePointVisits(c *Claims, w http.ResponseWriter, r *http.Request) {
	mapid, err := qId(mux.Vars(r)["mapid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	pid, err := qId(mux.Vars(r)["pid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		handleListVisits(c, mapid, pid, w)
	case "POST":
		handleSaveVisit(c, mapid, pid, w, r)
	}
}

func handleListMaps(c *Claims, w http.ResponseWriter, r *http.Request) {
	if c.UserId == "" {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	mcol := storage.Col(c.mapsCol())
	defer mcol.Close()

	var maps []*Map

	err := mcol.Iter(&Map{}, func(id Id, x Obj) error {
		nm := *(x.(*Map)) /* copy the map! */
		maps = append(maps, &nm)
		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(maps) == 0 {
		id, err := mcol.Add(-1, &Map{Name: "default"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		maps = append(maps, &Map{Id: id, Name: "default"})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&LoadMapsResp{M: maps})
}

func handleCreateMap(c *Claims, w http.ResponseWriter, r *http.Request) {
	var m Map

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mcol := storage.Col(c.mapsCol())
	defer mcol.Close()

	id, err := mcol.Add(-1, &m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m.Id = id
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&m)
}

func handleDeleteMap(c *Claims, mapid Id, w http.ResponseWriter, r *http.Request) {
	mcol := storage.Col(c.mapsCol())
	defer mcol.Close()

	err := mcol.Del(mapid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	storage.Drop(c.pointsCol(mapid))
	storage.Drop(c.areasCol(mapid))

	w.WriteHeader(http.StatusNoContent)
}

func handlePutMap(c *Claims, mapid Id, w http.ResponseWriter, r *http.Request) {
	/* FIXME -- implement */
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func handleGetMap(c *Claims, mapid Id, w http.ResponseWriter, r *http.Request) {
	/* FIXME -- implement */
	w.WriteHeader(http.StatusMethodNotAllowed)
}


func handleMaps(c *Claims, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleListMaps(c, w, r)
	case "POST":
		handleCreateMap(c, w, r)
	}
}

func handleUpdateMap(c *Claims, mapid Id, w http.ResponseWriter, r *http.Request) {
	var nm Map

	mcol := storage.Col(c.mapsCol())
	defer mcol.Close()

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&nm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = mcol.Upd(mapid, &Map{}, func(x Obj) error {
		om := x.(*Map)
		om.Name = nm.Name
		return nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleMap(c *Claims, w http.ResponseWriter, r *http.Request) {
	mapid, err := qId(mux.Vars(r)["mapid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch r.Method {
	case "PUT":
		handlePutMap(c, mapid, w, r)
	case "GET":
		handleGetMap(c, mapid, w, r)
	case "PATCH":
		handleUpdateMap(c, mapid, w, r)
	case "DELETE":
		handleDeleteMap(c, mapid, w, r)
	}
}

type auth func(*Claims, http.ResponseWriter, *http.Request)

func (fn auth)ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var c *Claims

	token := r.Header.Get("Authorization")
	if token == "" {
		if r.Method != "GET" {
			http.Error(w, "no token", http.StatusUnauthorized)
			return
		}

		c = &Claims{}
	} else {
		tok, err := jwt.ParseWithClaims(token, &Claims{},
			func(tok *jwt.Token) (interface{}, error) {
				if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("Unexpected sign method")
				}

				return tokenKey, nil
			})
		if err != nil {
			log.Printf("Bad token: %s", err.Error())
			http.Error(w, "bad token", http.StatusUnauthorized)
			return
		}

		c = tok.Claims.(*Claims)
	}

	log.Printf("%s/%s @%s\n", r.Method, r.URL.Path, c.UserId)
	if *debug {
		log.Printf("q: %v\n", r.URL.Query())
		log.Printf("h: %v\n", r.Header)
	}
	fn(c, w, r)
}

var tokenKey []byte
var debug *bool

func main() {
	var err error

	debug = flag.Bool("debug", false, "debug")
	strg := flag.String("storage", "ljs:db", "storage type and location")
	flag.Parse()

	err = setupStorage(*strg)
	if err != nil {
		log.Printf("Cannot setup storage: %s", err.Error())
		return
	}

	tokenKey, err = base64.StdEncoding.DecodeString(os.Getenv("JWT_SIGN_KEY"))
	if err != nil || len(tokenKey) == 0 {
		log.Printf("No JWT key provided")
		return
	}

	r := mux.NewRouter()
	r.Handle("/maps", auth(handleMaps)).Methods("GET", "POST", "OPTIONS")
	r.Handle("/maps/{mapid}", auth(handleMap)).Methods("PUT", "PATCH", "DELETE", "GET", "OPTIONS")
	r.Handle("/maps/{mapid}/geos", auth(handleMapGeos)).Methods("GET", "POST", "OPTIONS")
	r.Handle("/maps/{mapid}/visits", auth(handleMapVisits)).Methods("GET", "OPTIONS")
	r.Handle("/maps/{mapid}/geos/points/{pid}", auth(handleMapPoint)).Methods("DELETE", "PATCH", "OPTIONS")
	r.Handle("/maps/{mapid}/geos/points/{pid}/visits", auth(handlePointVisits)).Methods("GET", "POST", "OPTIONS")
	r.Handle("/maps/{mapid}/geos/points/{pid}/visits/{vn}", auth(handlePointVisit)).Methods("DELETE", "OPTIONS")
	r.Handle("/maps/{mapid}/geos/areas/{aid}", auth(handleMapArea)).Methods("DELETE", "OPTIONS")

	headersOk := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "PUT", "PATCH", "POST", "DELETE", "HEAD", "OPTIONS"})

	err = http.ListenAndServe("0.0.0.0:8082",
			handlers.CORS(originsOk, headersOk, methodsOk)(r))
	if err != nil {
		log.Fatalf("Cannot start HTTP server: %s", err.Error())
	}
}
