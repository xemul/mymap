package main

import (
	"os"
	"log"
	"flag"
	"sync"
	"errors"
	"strconv"
	"net/http"
	"math/rand"
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

func withGeos(c *Claims, mapid Id, af func(Collection) error, pf func(Collection) error) error {
	var wg sync.WaitGroup

	wg.Add(2)

	var aerr error
	go func() {
		acol := storage.Col(c.areasCol(mapid))
		defer acol.Close()
		aerr = af(acol)
		wg.Done()
	}()

	var perr error

	go func() {
		pcol := storage.Col(c.pointsCol(mapid))
		defer pcol.Close()
		perr = pf(pcol)
		wg.Done()
	}()

	wg.Wait()

	err := aerr
	if err == nil {
		err = perr
	}

	return err
}

func handleMaps(c *Claims, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleListMaps(c, w, r)
	case "POST":
		handleCreateMap(c, w, r)
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

	err := mcol.Iter(func() Obj { return &Map{} }, func(id Id, x Obj) error {
		nm := x.(*Map)
		maps = append(maps, nm)
		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(maps) == 0 {
		m := &Map{Name: "default"}
		err := createMap(c, m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		maps = append(maps, m)
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

	err = createMap(c, &m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&m)
}

func createMap(c *Claims, m *Map) error {
	for i := 0; i < 128; i++ {
		id := Id(rand.Int() % 10000000)
		ac := storage.Col(c.areasCol(id))
		err := ac.Create()
		ac.Close()

		if err == nil {
			m.Id = id
			break
		} else if err != IdExistsErr {
			return err
		}
	}

	if m.Id == -1 {
		return errors.New("cannot generate map ID")
	}

	if m.Copy != nil {
		err := withGeos(c, m.Id,
			func(acol Collection) error {
				oc := storage.Col(c.areasCol(*m.Copy))
				defer oc.Close()

				areas, err := oc.Raw()
				if err == nil {
					err = acol.Write(areas)
				}

				return err
			},
			func(pcol Collection) error {
				oc := storage.Col(c.pointsCol(*m.Copy))
				defer oc.Close()

				points, err := oc.Raw()
				if err == nil {
					err = pcol.Write(points)
				}

				return err
			})
		if err != nil {
			log.Printf("Cannot copy map: %s", err.Error())
		}
	}

	mcol := storage.Col(c.mapsCol())
	defer mcol.Close()

	err := mcol.Add(m.Id, &m)
	if err != nil {
		return err
	}

	return nil
}

type mapH func(*Claims, Id, http.ResponseWriter, *http.Request)
type geoH func(*Claims, Collection, Id, http.ResponseWriter, *http.Request)

func withMap(handle mapH) auth {
	return func(c *Claims, w http.ResponseWriter, r *http.Request) {
		mapid, err := qId(mux.Vars(r)["mapid"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if r.Method != "GET" && !c.checkMapCol(mapid) {
			http.Error(w, "no such map", http.StatusNotFound)
			return
		}

		handle(c, mapid, w, r)
	}
}

func withPoint(handle geoH) mapH {
	return func(c *Claims, mapid Id, w http.ResponseWriter, r *http.Request) {
		pid, err := qId(mux.Vars(r)["pid"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		col := storage.Col(c.pointsCol(mapid))
		defer col.Close()

		handle(c, col, pid, w, r)
	}
}

func withArea(handle geoH) mapH {
	return func(c *Claims, mapid Id, w http.ResponseWriter, r *http.Request) {
		aid, err := qId(mux.Vars(r)["aid"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		col := storage.Col(c.areasCol(mapid))
		defer col.Close()

		handle(c, col, aid, w, r)
	}
}

func handleMap(c *Claims, mapid Id, w http.ResponseWriter, r *http.Request) {
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

func handlePutMap(c *Claims, mapid Id, w http.ResponseWriter, r *http.Request) {
	var geos RawGeos

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&geos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = withGeos(c, mapid,
		func(acol Collection) error { return acol.Write(geos.Areas) },
		func(pcol Collection) error { return pcol.Write(geos.Points) })
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleGetMap(c *Claims, mapid Id, w http.ResponseWriter, r *http.Request) {
	acol := storage.Col(c.areasCol(mapid))
	defer acol.Close()
	areas, err := acol.Raw()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pcol := storage.Col(c.pointsCol(mapid))
	defer pcol.Close()
	points, err := pcol.Raw()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ret := RawGeos {
		Areas: areas,
		Points: points,
	}

	data, err := json.MarshalIndent(&ret, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
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

func handleMapGeos(c *Claims, mapid Id, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleListGeos(c, mapid, w, r)
	case "POST":
		handleSaveGeos(c, mapid, w, r)
	}
}

func handleListGeos(c *Claims, mapid Id, w http.ResponseWriter, r *http.Request) {
	var geos LoadGeosResp

	err := withGeos(c, mapid,
		func(acol Collection) error {
			return acol.Iter(func() Obj { return &Area{} }, func(id Id, x Obj) error {
				a := x.(*Area)
				geos.Areas = append(geos.Areas, a)
				return nil
			})
		},
		func(pcol Collection) error {
			return pcol.Iter(func() Obj { return &Point{} }, func(id Id, x Obj) error {
				pt := x.(*Point)
				geos.Points = append(geos.Points, pt)
				return nil
			})
		})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&geos)
}

func handleSaveGeos(c *Claims, mapid Id, w http.ResponseWriter, r *http.Request) {
	var sv SaveGeoReq

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&sv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = withGeos(c, mapid,
		func(acol Collection) error {
			if len(sv.Areas) > 0 {
				i := -1
				return acol.AddMany(func() (Id, Obj) {
					i++
					if i >= len(sv.Areas) {
						return -1, nil
					}
					a := sv.Areas[i]
					return a.Id, a
				})
			} else {
				return nil
			}
		},
		func(pcol Collection) error {
			if sv.Point != nil {
				return pcol.Add(sv.Point.Id, sv.Point)
			} else {
				return nil
			}
		})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleMapVisits(c *Claims, mapid Id, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		pcol := storage.Col(c.pointsCol(mapid))
		defer pcol.Close()
		handleListVisits(c, pcol, -1, w)
	}
}

func handleListVisits(c *Claims, pcol Collection, pid Id, w http.ResponseWriter) {
	var pts []*PointX
	var err error

	if pid != -1 {
		var pt PointX
		err = pcol.Get(pid, &pt)
		pts = append(pts, &pt)
	} else {
		err = pcol.Iter(func() Obj { return &PointX{} }, func(id Id, x Obj) error {
			pt := x.(*PointX)
			pts = append(pts, pt)
			return nil
		})
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var viss LoadVisitsResp
	for _, pt := range pts {
		for _, v := range pt.Vis {
			v.PId = &pt.Id
			viss.A = append(viss.A, v)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&viss)
}

func handleMapPoint(c *Claims, pcol Collection, pid Id, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PATCH":
		handleUpdatePoint(c, pcol, pid, w, r)
	case "DELETE":
		handleRemovePoint(c, pcol, pid, w)
	}
}

func handleUpdatePoint(c *Claims, pcol Collection, pid Id, w http.ResponseWriter, r *http.Request) {
	var npt Point

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&npt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

func handleRemovePoint(c *Claims, pcol Collection, pid Id, w http.ResponseWriter) {
	err := pcol.Del(pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handlePointVisits(c *Claims, pcol Collection, pid Id, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleListVisits(c, pcol, pid, w)
	case "POST":
		handleSaveVisit(c, pcol, pid, w, r)
	}
}

func handleSaveVisit(c *Claims, pcol Collection, pid Id, w http.ResponseWriter, r *http.Request) {
	var sv Visit

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&sv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

func handlePointVisit(c *Claims, pcol Collection, pid Id, w http.ResponseWriter, r *http.Request) {
	vn, err := qInt(mux.Vars(r)["vn"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch r.Method {
	case "DELETE":
		handleDeleteVisit(c, pcol, pid, vn, w)
	}
}

func handleDeleteVisit(c *Claims, pcol Collection, pid Id, vn int, w http.ResponseWriter) {
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

func handleMapArea(c *Claims, acol Collection, aid Id, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "DELETE":
		handleRemoveArea(c, acol, aid, w)
	}
}

func handleRemoveArea(c *Claims, acol Collection, aid Id, w http.ResponseWriter) {
	err := acol.Del(aid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
	r.Handle("/maps",
			auth(handleMaps)).Methods("GET", "POST", "OPTIONS")
	r.Handle("/maps/{mapid}",
			auth(withMap(handleMap))).Methods("PUT", "PATCH", "DELETE", "GET", "OPTIONS")
	r.Handle("/maps/{mapid}/geos",
			auth(withMap(handleMapGeos))).Methods("GET", "POST", "OPTIONS")
	r.Handle("/maps/{mapid}/visits",
			auth(withMap(handleMapVisits))).Methods("GET", "OPTIONS")
	r.Handle("/maps/{mapid}/geos/points/{pid}",
			auth(withMap(withPoint(handleMapPoint)))).Methods("DELETE", "PATCH", "OPTIONS")
	r.Handle("/maps/{mapid}/geos/points/{pid}/visits",
			auth(withMap(withPoint(handlePointVisits)))).Methods("GET", "POST", "OPTIONS")
	r.Handle("/maps/{mapid}/geos/points/{pid}/visits/{vn}",
			auth(withMap(withPoint(handlePointVisit)))).Methods("DELETE", "OPTIONS")
	r.Handle("/maps/{mapid}/geos/areas/{aid}",
			auth(withMap(withArea(handleMapArea)))).Methods("DELETE", "OPTIONS")

	headersOk := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "PUT", "PATCH", "POST", "DELETE", "HEAD", "OPTIONS"})

	err = http.ListenAndServe("0.0.0.0:8082",
			handlers.CORS(originsOk, headersOk, methodsOk)(r))
	if err != nil {
		log.Fatalf("Cannot start HTTP server: %s", err.Error())
	}
}
