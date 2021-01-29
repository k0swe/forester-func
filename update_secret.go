package kellog

import (
	"context"
	"log"
	"net/http"
	"time"
)

// Import QSOs from QRZ logbook and merge into Firestore. Called via GCP Cloud Functions.
func UpdateSecret(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if handleCorsOptions(w, r) {
		return
	}
	log.Print("Serving UpdateSecret")
	fb, err := MakeFirebaseManager(&ctx, r)
	if err != nil {
		writeError(500, "", err, w)
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
		_, err = checkAndSetSecret(secretStore, fb, lotwUsername, lotwUser)
		if err != nil {
			writeError(500, "Error storing a secret", err, w)
			return
		}
	}
	if lotwPass != "" {
		_, err = checkAndSetSecret(secretStore, fb, lotwPassword, lotwPass)
		if err != nil {
			writeError(500, "Error storing a secret", err, w)
			return
		}
	}
	if qrzKey != "" {
		_, err = checkAndSetSecret(secretStore, fb, qrzLogbookApiKey, qrzKey)
		if err != nil {
			writeError(500, "Error storing a secret", err, w)
			return
		}
	}
	// TODO: put a flag in firestore
	w.WriteHeader(204)
}

func checkAndSetSecret(secretStore SecretStore, fb *FirebaseManager, key string, value string) (string, error) {
	// Rely on Firestore rules to check that user is an editor on the logbook
	err := fb.SetLogbookSetting(
		key+"_last_set",
		time.Now().UTC().String(),
	)
	if err != nil {
		return "", err
	}

	return secretStore.SetSecret(fb.logbookId, key, value)
}
