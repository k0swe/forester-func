package kellog

import (
	"context"
	"net/http"
)

// Import QSOs from QRZ logbook and merge into Firestore. Called via GCP Cloud Functions.
func UpdateSecret(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if handleCorsOptions(w, r) {
		return
	}
	fb, err := MakeFirebaseManager(&ctx, r)
	if err != nil {
		return
	}

	err = r.ParseMultipartForm(102400)
	if err != nil {
		writeError(500, "Error parsing form values", err, w)
		return
	}
	lotwUser := r.PostFormValue(lotwUsername)
	lotwPass := r.PostFormValue(lotwPassword)
	qrzKey := r.PostFormValue(qrzLogbookApiKey)
	if lotwUser == "" && lotwPass == "" && qrzKey == "" {
		w.WriteHeader(204)
		return
	}

	secretStore := NewSecretStore(ctx)
	if lotwUser != "" {
		_, err = secretStore.SetSecret(fb.GetUID(), lotwUsername, lotwUser)
		if err != nil {
			writeError(500, "Error storing a secret", err, w)
			return
		}
	}
	if lotwPass != "" {
		_, err = secretStore.SetSecret(fb.GetUID(), lotwPassword, lotwPass)
		if err != nil {
			writeError(500, "Error storing a secret", err, w)
			return
		}
	}
	if qrzKey != "" {
		_, err = secretStore.SetSecret(fb.GetUID(), qrzLogbookApiKey, qrzKey)
		if err != nil {
			writeError(500, "Error storing a secret", err, w)
			return
		}
	}
	// TODO: put a flag in firestore
	w.WriteHeader(204)
}
