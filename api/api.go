package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"

	"../db"
)

var allowed_epochs = []string{
	"3-hours",
	"6-hours",
	"12-hours",
	"1-day",
	"3-days",
	"7-days",
	"14-days",
	"1-month",
	"2-months",
	"3-months",
	"6-months",
	"1-year",
	"all",
}

func AllEpochs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	out, err := json.Marshal(allowed_epochs)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, string(out))
}

func AllProducers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	producers, err := db.AllProducers()

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	out, err := json.Marshal(producers)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, string(out))
}

func AllBenchmarks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	epoch := vars["epoch"]

	if stringInSlice(epoch, allowed_epochs) {
		benchmarks, err := db.AllBenchmarks(epoch)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		out, err := json.Marshal(benchmarks)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, string(out))
	} else {
		http.Error(w, "Illegal epoch provided.", 500)
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
