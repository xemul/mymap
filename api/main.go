package main

import (
	"log"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
)

type Area struct {
	Id	int		`json:"id"`
	Name	string		`json:"name"`
	Type	string		`json:"type"`
}

func handleListVisited(w http.ResponseWriter, r *http.Request) {
	log.Printf("-> GET")
	w.WriteHeader(http.StatusOK)
}

func handleSaveVisited(w http.ResponseWriter, r *http.Request) {
	var areas []Area

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&areas)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("-> POST [%v]", areas)
	w.WriteHeader(http.StatusOK)
}

func handleForgetVisited(w http.ResponseWriter, r *http.Request) {
	areaId := r.URL.Query()["id"]
	log.Printf("-> DELETE %s", areaId)
	w.WriteHeader(http.StatusNoContent)
}

func handleVisited(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleListVisited(w, r)
	case "POST":
		handleSaveVisited(w, r)
	case "DELETE":
		handleForgetVisited(w, r)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/visited", handleVisited).Methods("GET", "POST", "DELETE", "OPTIONS")

	headersOk := handlers.AllowedHeaders([]string{"Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "HEAD", "OPTIONS"})

	err := http.ListenAndServe("0.0.0.0:8082",
			handlers.CORS(originsOk, headersOk, methodsOk)(r))
	if err != nil {
		log.Fatalf("Cannot start HTTP server: %s", err.Error())
	}
}
