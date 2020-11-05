package kellog

import (
	"net/http"
)

// Import QSOs from QRZ logbook and merge into Firestore. Called via GCP Cloud Functions.
func UpdateSecret(w http.ResponseWriter, r *http.Request) {
	ctx, userToken, _, done, err := getUserFirestore(w, r)
	if done || err != nil {
		return
	}

	err = r.ParseMultipartForm(102400)
	if err != nil {
		writeError(500, "Error parsing form values", err, w)
		return
	}
	lotwUser := r.PostFormValue(lotwUsername)
	lotwPass := r.PostFormValue(lotwPassword)
	if lotwUser == "" && lotwPass == "" {
		w.WriteHeader(204)
		return
	}

	secretStore := NewSecretStore(ctx)
	if lotwUser != "" {
		_, err = secretStore.SetSecret(userToken.UID, lotwUsername, lotwUser)
		if err != nil {
			writeError(500, "Error storing a secret", err, w)
			return
		}
	}
	if lotwPass != "" {
		_, err = secretStore.SetSecret(userToken.UID, lotwPassword, lotwPass)
		if err != nil {
			writeError(500, "Error storing a secret", err, w)
			return
		}
	}
	// TODO: put a flag in firestore
	w.WriteHeader(204)
}
