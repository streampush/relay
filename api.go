package main

import (
	"encoding/json"
	"log"
	"net/http"
)

const _apiRoot = "/api/"

func StartAPIServer() {
	http.HandleFunc(_apiRoot+"reload", func(w http.ResponseWriter, r *http.Request) {
		if reloadConfigs() == 0 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	http.HandleFunc(_apiRoot+"stats", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(restreams.cfgs)
	})

	log.Fatal(http.ListenAndServe(API_LISTEN, nil))
}
