package forester

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	ql "github.com/k0swe/qrz-logbook"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

const qrzLastFetchedDate = "qrzLastFetchedDate"

// ImportQrz imports QSOs from QRZ logbook and merges them into Firestore. Called via GCP Cloud
// Functions.
func ImportQrz(w http.ResponseWriter, r *http.Request) {
	const isFixCase = true
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if handleCorsOptions(w, r) {
		return
	}
	log.Info().Msg("Starting ImportQrz")
	fb, err := MakeFirebaseManager(&ctx, r)
	if err != nil {
		writeError(500, "Error", err, w)
		return
	}
	_, err = fb.GetLogbookProperty(qrzLastFetchedDate)
	if err != nil {
		writeError(500, "Error fetching logbook properties from firestore", err, w)
		return
	}

	secretStore := NewSecretStore(ctx)
	qrzApiKey, err := secretStore.FetchSecret(fb.logbookId, qrzLogbookApiKey)
	if err != nil {
		writeError(500, "Error fetching QRZ API key from secret manager", err, w)
		return
	}
	qrzResponse, err := ql.Fetch(ctx, &qrzApiKey)
	if err != nil {
		writeError(500, "Error fetching QRZ.com data", err, w)
		return
	}
	log.Info().Uint64("recordCount", qrzResponse.Count).Msg("Fetched QRZ.com data")
	qrzAdi, err := adifToProto(qrzResponse.Adif, time.Now())
	if err != nil {
		writeError(500, "Failed parsing QRZ.com data", err, w)
		log.Debug().
			Str("payload", base64.StdEncoding.EncodeToString([]byte(qrzResponse.Adif))).
			Msg("Failed parsing QRZ.com data")
		return
	}
	if isFixCase {
		for _, qso := range qrzAdi.Qsos {
			fixCase(qso)
		}
	}
	fsContacts, err := fb.GetContacts()
	if err != nil {
		writeError(500, "Error fetching contacts from firestore", err, w)
		return
	}
	created, modified, noDiff := fb.MergeQsos(fsContacts, qrzAdi)

	var report = map[string]int{}
	report["qrz"] = len(qrzAdi.Qsos)
	report["firestore"] = len(fsContacts)
	report["created"] = created
	report["modified"] = modified
	report["noDiff"] = noDiff
	marshal, _ := json.Marshal(report)
	log.Info().RawJSON("report", marshal).Msg("Complete")
	_, _ = fmt.Fprint(w, string(marshal))
}
