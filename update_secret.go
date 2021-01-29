package kellog

import (
	"context"
	"fmt"
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
	log.Print("Starting UpdateSecret")
	fb, err := MakeFirebaseManager(&ctx, r)
	if err != nil {
		writeError(500, "Error", err, w)
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
		log.Print("Nothing to do")
		w.WriteHeader(204)
		return
	}

	secretStore := NewSecretStore(ctx)
	if lotwUser != "" {
		log.Printf("Updating %v", lotwUsername)
		_, err = checkAndSetSecret(secretStore, fb, lotwUsername, lotwUser)
		if err != nil {
			writeError(500, "Error storing a secret", err, w)
			return
		}
	}
	if lotwPass != "" {
		log.Printf("Updating %v", lotwPassword)
		_, err = checkAndSetSecret(secretStore, fb, lotwPassword, lotwPass)
		if err != nil {
			writeError(500, "Error storing a secret", err, w)
			return
		}
	}
	if qrzKey != "" {
		log.Printf("Updating %v", qrzLogbookApiKey)
		_, err = checkAndSetSecret(secretStore, fb, qrzLogbookApiKey, qrzKey)
		if err != nil {
			writeError(500, "Error storing a secret", err, w)
			return
		}
	}
	w.WriteHeader(204)
}

func checkAndSetSecret(secretStore SecretStore, fb *FirebaseManager, key string, value string) (string, error) {
	// Rely on Firestore rules to check that user is an editor on the logbook
	err := fb.SetLogbookProperty(
		key+"_last_set",
		time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		// 403
		return "", fmt.Errorf("can't modify logbook, are you an editor? "+
			"not saving secret: %v", err)
	}

	return secretStore.SetSecret(fb.logbookId, key, value)
}
