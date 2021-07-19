package main

import (
	"github.com/k0swe/forester-func"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
)

// Local development driver, not used by Cloud Functions
func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if os.Getenv("GCP_PROJECT") == "" {
		panic("GCP_PROJECT is not set")
	} else {
		log.Info().Str("GCP_PROJECT", os.Getenv("GCP_PROJECT"))
	}
	const addr = "localhost:8080"
	http.HandleFunc("/ImportQrz", forester.ImportQrz)
	http.HandleFunc("/ImportLotw", forester.ImportLotw)
	http.HandleFunc("/UpdateSecret", forester.UpdateSecret)
	log.Info().Str("address", "http://"+addr).Msg("Ready to serve")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal().Err(err).Msg("Startup failed")
	}
}
