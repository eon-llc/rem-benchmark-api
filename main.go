package main

import (
	"./api"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/epochs", api.AllEpochs).Methods("GET")
	router.HandleFunc("/producers", api.AllProducers).Methods("GET")
	router.HandleFunc("/benchmarks", api.AllBenchmarks).Methods("GET").Queries("epoch", "{epoch}")

	srv := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
