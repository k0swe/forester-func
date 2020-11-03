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
	} else {
		log.Printf("GCP_PROJECT is %v", os.Getenv("GCP_PROJECT"))
	}
	const addr = "localhost:8080"
	http.HandleFunc("/ImportQrz", kellog.ImportQrz)
	http.HandleFunc("/ImportLotw", kellog.ImportLotw)
	log.Printf("Ready to serve on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
