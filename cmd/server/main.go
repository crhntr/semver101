package main

import (
	"cmp"
	"log"
	"net/http"
	"os"

	"github.com/crhntr/semver101"
)

func main() {
	addr := ":" + cmp.Or(os.Getenv("PORT"), "8080")
	log.Fatal(http.ListenAndServe(addr, routes()))
}

func routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", semver101.HandleGet("/"))
	mux.HandleFunc("POST /", semver101.HandlePost("/"))
	return mux
}
