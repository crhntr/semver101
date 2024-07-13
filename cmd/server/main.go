package main

import (
	"cmp"
	"log"
	"net/http"
	"os"

	"github.com/crhntr/semver101"
)

func main() {
	log.Fatal(http.ListenAndServe(":"+cmp.Or(os.Getenv("PORT"), "8080"), semver101.Handler("")))
}
