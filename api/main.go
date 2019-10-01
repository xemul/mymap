package main

import (
	"log"
	"strconv"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
)

func handleListGeos(w http.ResponseWriter, r *http.Request) {
	load, err := storage.LoadGeos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(load)
}

func handleSaveGeos(w http.ResponseWriter, r *http.Request) {
	var sv SavePointReq

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&sv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = storage.SavePoint(&sv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleForgetGeos(w http.ResponseWriter, r *http.Request) {
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

	ok, err := storage.RemoveGeo(areaId, typ[0])
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

func handleGeos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleListGeos(w, r)
	case "POST":
		handleSaveGeos(w, r)
	case "DELETE":
		handleForgetGeos(w, r)
	}
}

func handleListVisits(w http.ResponseWriter, r *http.Request, id int) {
	viss, err := storage.LoadVisits(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(viss)
}

func handleSaveVisit(w http.ResponseWriter, r *http.Request, id int) {
	var sv SaveVisitReq

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&sv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = storage.SaveVisit(id, &sv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleDeleteVisit(w http.ResponseWriter, r *http.Request, id int) {
	vn, err := strconv.Atoi(r.URL.Query()["vn"][0])
	if err != nil {
		http.Error(w, "vn must be integer", http.StatusBadRequest)
		return
	}

	ok, err := storage.RemoveVisit(id, vn)
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

func handleVisits(w http.ResponseWriter, r *http.Request) {
	ptId, err := strconv.Atoi(r.URL.Query()["id"][0])
	if err != nil {
		http.Error(w, "id must be integer", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		handleListVisits(w, r, ptId)
	case "POST":
		handleSaveVisit(w, r, ptId)
	case "DELETE":
		handleDeleteVisit(w, r, ptId)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/geos", handleGeos).Methods("GET", "POST", "DELETE", "OPTIONS")
	r.HandleFunc("/visits", handleVisits).Methods("GET", "POST", "DELETE", "OPTIONS")

	headersOk := handlers.AllowedHeaders([]string{"Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "HEAD", "OPTIONS"})

	err := http.ListenAndServe("0.0.0.0:8082",
			handlers.CORS(originsOk, headersOk, methodsOk)(r))
	if err != nil {
		log.Fatalf("Cannot start HTTP server: %s", err.Error())
	}
}
