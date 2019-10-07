package main

import (
	"os"
	"log"
	"errors"
	"strconv"
	"net/http"
	"encoding/json"
	"encoding/base64"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	*jwt.StandardClaims
	UserId	string		`json:"id"`
}

func getMap(c *Claims, w http.ResponseWriter, r *http.Request) Geos {
	mp, err := geos(c)

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
	areaId, err := strconv.Atoi(r.URL.Query()["id"][0])
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
	vn, err := strconv.Atoi(r.URL.Query()["vn"][0])
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

	id := r.URL.Query()["id"]
	ptid := -1

	if len(id) == 0 {
		if r.Method != "GET" {
			http.Error(w, "no id given", http.StatusBadRequest)
			return
		}
	} else {

		var err error

		ptid, err = strconv.Atoi(id[0])
		if err != nil {
			http.Error(w, "id must be integer", http.StatusBadRequest)
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

type auth func(*Claims, http.ResponseWriter, *http.Request)

func (fn auth)ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var c *Claims

	if r.Method == "GET" {
		viewmap := r.URL.Query()["viewmap"]
		if len(viewmap) == 1 {
			c = &Claims{ UserId: viewmap[0] }
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
	fn(c, w, r)
}

var tokenKey []byte

func main() {
	var err error

	tokenKey, err = base64.StdEncoding.DecodeString(os.Getenv("JWT_SIGN_KEY"))
	if err != nil || len(tokenKey) == 0 {
		log.Printf("No JWT key provided")
		return
	}

	r := mux.NewRouter()
	r.Handle("/geos", auth(handleGeos)).Methods("GET", "POST", "DELETE", "OPTIONS")
	r.Handle("/visits", auth(handleVisits)).Methods("GET", "POST", "DELETE", "OPTIONS")

	headersOk := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "HEAD", "OPTIONS"})

	err = http.ListenAndServe("0.0.0.0:8082",
			handlers.CORS(originsOk, headersOk, methodsOk)(r))
	if err != nil {
		log.Fatalf("Cannot start HTTP server: %s", err.Error())
	}
}
