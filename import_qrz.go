package kellog

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ql "github.com/k0swe/qrz-logbook"
	"net/http"
	"time"
)

// Import QSOs from QRZ logbook and merge into Firestore. Called via GCP Cloud Functions.
func ImportQrz(w http.ResponseWriter, r *http.Request) {
	const isFixCase = true
	ctx, userToken, firestoreClient, done, err := getUserFirestore(w, r)
	if done || err != nil {
		return
	}

	qrzApiKey, err := getQrzApiKey(ctx, firestoreClient, userToken.UID)
	if err != nil {
		writeError(500, "Error fetching QRZ API key from firestore", err, w)
		return
	}
	qrzResponse, err := ql.Fetch(&qrzApiKey)
	if err != nil {
		writeError(500, "Error fetching QRZ.com data", err, w)
		return
	}
	qrzAdi, err := adifToProto(qrzResponse.Adif, time.Now())
	if err != nil {
		writeError(500, "Failed parsing QRZ.com data", err, w)
		return
	}
	if isFixCase {
		for _, qso := range qrzAdi.Qsos {
			fixCase(qso)
		}
	}
	contactsRef := firestoreClient.Collection("users").Doc(userToken.UID).Collection("contacts")
	fsContacts, err := getContacts(ctx, contactsRef)
	if err != nil {
		writeError(500, "Error fetching contacts from firestore", err, w)
		return
	}
	created, modified, noDiff := mergeQsos(fsContacts, qrzAdi, contactsRef, ctx)

	var report = map[string]int{}
	report["qrz"] = len(qrzAdi.Qsos)
	report["firestore"] = len(fsContacts)
	report["created"] = created
	report["modified"] = modified
	report["noDiff"] = noDiff
	marshal, _ := json.Marshal(report)
	_, _ = fmt.Fprint(w, string(marshal))
}

func getQrzApiKey(ctx context.Context, firestoreClient *firestore.Client, userUid string) (string, error) {
	docSnapshot, err := firestoreClient.Collection("users").Doc(userUid).Get(ctx)
	if err != nil {
		return "", err
	}
	qrzApiKey := fmt.Sprint(docSnapshot.Data()["qrzLogbookApiKey"])
	if qrzApiKey == "" {
		return "", errors.New("user hasn't set up their QRZ.com API key")
	}
	return qrzApiKey, nil
}
