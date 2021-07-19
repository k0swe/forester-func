package forester

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

// UpdateSecret allows users to add or change their stored secrets. Called via GCP Cloud Functions.
func UpdateSecret(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if handleCorsOptions(w, r) {
		return
	}
	log.Info().Msg("Starting UpdateSecret")
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
	qrzUser := r.PostFormValue(qrzUsername)
	qrzPass := r.PostFormValue(qrzPassword)
	qrzKey := r.PostFormValue(qrzLogbookApiKey)
	if lotwUser == "" && lotwPass == "" && qrzUser == "" && qrzPass == "" && qrzKey == "" {
		log.Info().Msg("Nothing to do")
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
	if qrzUser != "" {
		_, err = checkAndSetSecret(secretStore, fb, qrzUsername, qrzUser)
		if err != nil {
			writeError(500, "Error storing a secret", err, w)
			return
		}
	}
	if qrzPass != "" {
		_, err = checkAndSetSecret(secretStore, fb, qrzPassword, qrzPass)
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
	w.WriteHeader(204)
}

func checkAndSetSecret(secretStore SecretStore, fb *FirebaseManager, key string, value string) (string, error) {
	log.Info().Str("secret", key).Msgf("Updating")
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
