package main

import (
	"github.com/k0swe/kellog-func"
	"log"
	"net/http"
	"os"
)

// Local development driver, not used by Cloud Functions
func main() {
	if os.Getenv("GCP_PROJECT") == "" {
		panic("GCP_PROJECT is not set")
	}
	const addr = "localhost:8080"
	const path = "/ImportQrz"
	http.HandleFunc(path, kellog.ImportQrz)
	log.Printf("Ready to serve on http://%s%s", addr, path)
	log.Fatal(http.ListenAndServe(addr, nil))
}
