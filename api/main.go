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

func qInt(r *http.Request, pn string) (int, error) {
	x := r.URL.Query()[pn]
	if len(x) == 1 {
		return strconv.Atoi(x[0])
	} else {
		return -1, noValue
	}
}

type Claims struct {
	*jwt.StandardClaims
	UserId	string		`json:"id"`
	viewmap	string		`json:"-"`
}

func getMap(c *Claims, w http.ResponseWriter, r *http.Request) Geos {
	x := c.viewmap
	if x == "" {
		x = r.Header.Get("X-MapId")
		if x == "" {
			http.Error(w, "no mapid", http.StatusBadRequest)
			return nil
		}
	}

	mapid, err := strconv.Atoi(x)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}

	db := openDB(c)
	defer db.Close()

	mp, err := db.Geos(mapid)
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

func handleListGeos(mp Geos, w http.ResponseWriter, r *http.Request) {
	load, err := mp.LoadGeos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(load)
}

func handleSaveGeos(mp Geos, w http.ResponseWriter, r *http.Request) {
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

func handleForgetGeos(mp Geos, w http.ResponseWriter, r *http.Request) {
	areaId, err := qInt(r, "id")
	if err != nil {
		http.Error(w, "id must be integer", http.StatusBadRequest)
		return
	}

	typ := r.URL.Query()["type"]
	if len(typ) == 0 {
		http.Error(w, "need type", http.StatusBadRequest)
		return
	}

	ok, err := mp.RemoveGeo(areaId, typ[0])
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

func handleGeos(c *Claims, w http.ResponseWriter, r *http.Request) {
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
	case "DELETE":
		handleForgetGeos(mp, w, r)
	}
}

func handleListVisits(mp Geos, w http.ResponseWriter, r *http.Request, id int) {
	viss, err := mp.LoadVisits(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(viss)
}

func handleSaveVisit(mp Geos, w http.ResponseWriter, r *http.Request, id int) {
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

func handleDeleteVisit(mp Geos, w http.ResponseWriter, r *http.Request, id int) {
	vn, err := qInt(r, "vn")
	if err != nil {
		http.Error(w, "vn must be integer", http.StatusBadRequest)
		return
	}

	log.Printf("[-v] %d:%d\n", id, vn)

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

func handleVisits(c *Claims, w http.ResponseWriter, r *http.Request) {
	mp := getMap(c, w, r)
	if mp == nil {
		return
	}

	defer mp.Close()

	ptid, err := qInt(r, "id")
	if err != nil {
		if !(r.Method == "GET" && err == noValue) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	switch r.Method {
	case "GET":
		handleListVisits(mp, w, r, ptid)
	case "POST":
		handleSaveVisit(mp, w, r, ptid)
	case "DELETE":
		handleDeleteVisit(mp, w, r, ptid)
	}
}

func handleListMaps(c *Claims, w http.ResponseWriter, r *http.Request) {
	if c.UserId == "" {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	db := openDB(c)
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

	db := openDB(c)
	defer db.Close()

	err = db.Create(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&m)
}

func handleDeleteMap(c *Claims, w http.ResponseWriter, r *http.Request) {
	mapid, err := qInt(r, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := openDB(c)
	defer db.Close()

	err = db.Remove(mapid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleMaps(c *Claims, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleListMaps(c, w, r)
	case "POST":
		handleCreateMap(c, w, r)
	case "DELETE":
		handleDeleteMap(c, w, r)
	}
}

type auth func(*Claims, http.ResponseWriter, *http.Request)

func (fn auth)ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var c *Claims

	if r.Method == "GET" {
		viewmap := r.URL.Query()["viewmap"]
		if len(viewmap) == 1 {
			c = &Claims{ viewmap: viewmap[0] }
		}
	}

	if c == nil {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "no token", http.StatusUnauthorized)
			return
		}

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
	flag.Parse()

	tokenKey, err = base64.StdEncoding.DecodeString(os.Getenv("JWT_SIGN_KEY"))
	if err != nil || len(tokenKey) == 0 {
		log.Printf("No JWT key provided")
		return
	}

	r := mux.NewRouter()
	r.Handle("/maps", auth(handleMaps)).Methods("GET", "POST", "DELETE", "OPTIONS")
	r.Handle("/geos", auth(handleGeos)).Methods("GET", "POST", "DELETE", "OPTIONS")
	r.Handle("/visits", auth(handleVisits)).Methods("GET", "POST", "DELETE", "OPTIONS")

	headersOk := handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-MapId"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "HEAD", "OPTIONS"})

	err = http.ListenAndServe("0.0.0.0:8082",
			handlers.CORS(originsOk, headersOk, methodsOk)(r))
	if err != nil {
		log.Fatalf("Cannot start HTTP server: %s", err.Error())
	}
}
