package main

import (
	"github.com/k0swe/kellog-func"
	"log"
	"net/http"
)

// Local development driver, not used by Cloud Functions
func main() {
	http.HandleFunc("/ImportQrz", kellog.ImportQrz)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
