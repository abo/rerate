package main

import (
	"encoding/json"
	"net/http"
	"time"

	redis "gopkg.in/redis.v5"

	"github.com/abo/rerate"
	"github.com/gorilla/mux"
)

var counter *rerate.Counter

func init() {
	buckets := rerate.NewRedisV5Buckets(redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}))
	counter = rerate.NewCounter(buckets, "rerate:sparkline", 20*time.Second, 500*time.Millisecond)
}

func main() {
	r := mux.NewRouter()
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	r.HandleFunc("/histogram/{key}", histogram)
	r.HandleFunc("/inc/{key}", inc)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("tpl")))

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	srv.ListenAndServe()
}

func inc(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	counter.Inc(vars["key"])
}

func histogram(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if hist, err := counter.Histogram(vars["key"]); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if resp, err := json.Marshal(hist); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	}
}
