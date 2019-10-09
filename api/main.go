package main

import (
	"os"
	"log"
	"flag"
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

func getMap(c *Claims, w http.ResponseWriter, r *http.Request) MDB {
	mapid, err := qInt(mux.Vars(r)["mapid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return nil
	}

	db, err := storage.openUDB(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	defer db.Close()

	mp, err := db.openMDB(mapid, r.Method != "GET")
	if mp == nil {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, "no such map", http.StatusNotFound)
		}

		return nil
	}

	return mp
}

func getPoint(c *Claims, w http.ResponseWriter, r *http.Request) (MDB, int) {
	mp := getMap(c, w, r)
	if mp == nil {
		return nil, -1
	}

	pid, err := qInt(mux.Vars(r)["pid"])
	if err != nil {
		mp.Close()
		http.Error(w, err.Error(), http.StatusNotFound)
		return nil, pid
	}

	return mp, pid
}

func handleListGeos(mp MDB, w http.ResponseWriter, r *http.Request) {
	load, err := mp.LoadGeos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(load)
}

func handleSaveGeos(mp MDB, w http.ResponseWriter, r *http.Request) {
	var sv SaveGeoReq

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&sv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = mp.SavePoint(&sv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleUpdatePoint(mp MDB, pid int, w http.ResponseWriter, r *http.Request) {
	var pt Point

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&pt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = mp.PatchPoint(pid, &pt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleMapPoint(c *Claims, w http.ResponseWriter, r *http.Request) {
	mp, pid := getPoint(c, w, r)
	if mp == nil {
		return
	}

	defer mp.Close()

	switch r.Method {
	case "PATCH":
		handleUpdatePoint(mp, pid, w, r)
	case "DELETE":
		handleRemoveGeo(mp, pid, "point", w)
	}
}

func handleMapArea(c *Claims, w http.ResponseWriter, r *http.Request) {
	mp := getMap(c, w, r)
	if mp == nil {
		return
	}

	defer mp.Close()

	aid, err := qInt(mux.Vars(r)["aid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch r.Method {
	case "DELETE":
		handleRemoveGeo(mp, aid, "area", w)
	}
}

func handleRemoveGeo(mp MDB, gid int, gtype string, w http.ResponseWriter) {
	ok, err := mp.RemoveGeo(gid, gtype)
	if err != nil {
		if ok {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, "no such area", http.StatusNotFound)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleMapGeos(c *Claims, w http.ResponseWriter, r *http.Request) {
	mp := getMap(c, w, r)
	if mp == nil {
		return
	}

	defer mp.Close()

	switch r.Method {
	case "GET":
		handleListGeos(mp, w, r)
	case "POST":
		handleSaveGeos(mp, w, r)
	}
}

func handleListVisits(mp MDB, w http.ResponseWriter, id int) {
	viss, err := mp.LoadVisits(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(viss)
}

func handleSaveVisit(mp MDB, w http.ResponseWriter, r *http.Request, id int) {
	var sv SaveVisitReq

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&sv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = mp.SaveVisit(id, &sv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handlePointVisit(c *Claims, w http.ResponseWriter, r *http.Request) {
	mp, pid := getPoint(c, w, r)
	if mp == nil {
		return
	}

	defer mp.Close()

	vn, err := qInt(mux.Vars(r)["vn"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch r.Method {
	case "DELETE":
		handleDeleteVisit(mp, w, pid, vn)
	}
}

func handleDeleteVisit(mp MDB, w http.ResponseWriter, id, vn int) {
	ok, err := mp.RemoveVisit(id, vn)
	if err != nil {
		if ok {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, "no such visit", http.StatusNotFound)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleMapVisits(c *Claims, w http.ResponseWriter, r *http.Request) {
	mp := getMap(c, w, r)
	if mp == nil {
		return
	}

	defer mp.Close()

	switch r.Method {
	case "GET":
		handleListVisits(mp, w, -1)
	}
}

func handlePointVisits(c *Claims, w http.ResponseWriter, r *http.Request) {
	mp, pid := getPoint(c, w, r)
	if mp == nil {
		return
	}

	defer mp.Close()

	switch r.Method {
	case "GET":
		handleListVisits(mp, w, pid)
	case "POST":
		handleSaveVisit(mp, w, r, pid)
	}
}

func handleGetMap(mp MDB, w http.ResponseWriter, r *http.Request) {
	data, err := mp.Raw()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func handleListMaps(c *Claims, w http.ResponseWriter, r *http.Request) {
	if c.UserId == "" {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	db, err := storage.openUDB(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer db.Close()

	maps, err := db.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	db, err := storage.openUDB(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer db.Close()

	err = db.Create(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&m)
}

func handleDeleteMap(c *Claims, mp MDB, w http.ResponseWriter, r *http.Request) {
	db, err := storage.openUDB(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer db.Close()

	err = db.Remove(mp.Id())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handlePutMap(mp MDB, w http.ResponseWriter, r *http.Request) {
	log.Printf("Upload map %d\n", mp.Id())

	err := mp.Put(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleMaps(c *Claims, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleListMaps(c, w, r)
	case "POST":
		handleCreateMap(c, w, r)
	}
}

func handleMap(c *Claims, w http.ResponseWriter, r *http.Request) {
	mp := getMap(c, w, r)
	if mp == nil {
		return
	}

	defer mp.Close()

	switch r.Method {
	case "PUT":
		handlePutMap(mp, w, r)
	case "GET":
		handleGetMap(mp, w, r)
	case "DELETE":
		handleDeleteMap(c, mp, w, r)
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
	r.Handle("/maps/{mapid}", auth(handleMap)).Methods("PUT", "DELETE", "GET", "OPTIONS")
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
