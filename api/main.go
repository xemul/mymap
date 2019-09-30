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
	load, err := storage.Load()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(load)
}

func handleSaveGeos(w http.ResponseWriter, r *http.Request) {
	var sv SaveReq

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&sv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = storage.Save(&sv)
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

	ok, err := storage.Remove(areaId, typ[0])
	if err != nil {
		if ok {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, "no such area", http.StatusNotFound)
		}
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

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/geos", handleGeos).Methods("GET", "POST", "DELETE", "OPTIONS")

	headersOk := handlers.AllowedHeaders([]string{"Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "HEAD", "OPTIONS"})

	err := http.ListenAndServe("0.0.0.0:8082",
			handlers.CORS(originsOk, headersOk, methodsOk)(r))
	if err != nil {
		log.Fatalf("Cannot start HTTP server: %s", err.Error())
	}
}
