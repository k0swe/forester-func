package main

import (
	"github.com/xylo04/kellog-qrz-sync"
	"log"
	"net/http"
)

// Local development driver, not used by Cloud Functions
func main() {
	http.HandleFunc("/", kellog.HelloHTTP)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
