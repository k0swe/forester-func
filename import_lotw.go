package kellog

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"fmt"
	adifpb "github.com/k0swe/adif-json-protobuf/go"
	"github.com/k0swe/lotw-qsl"
	"net/http"
	"strings"
	"time"
)

// Import QSOs from Logbook of the World and merge into Firestore. Called via GCP Cloud Functions.
func ImportLotw(w http.ResponseWriter, r *http.Request) {
	const isFixCase = true
	ctx, userToken, firestoreClient, done, err := getUserFirestore(w, r)
	if done || err != nil {
		return
	}

	lotwUser, lotwPass, err := getLotwCreds(ctx, firestoreClient, userToken.UID)
	if err != nil {
		writeError(500, "Error fetching LotW creds", err, w)
		return
	}
	lotwResponse, err := lotw.Query(lotwUser, lotwPass, &lotw.QueryOpts{})
	if err != nil {
		writeError(500, "Error fetching LotW data", err, w)
		return
	}
	lotwResponse = removeEof(lotwResponse)
	lotwAdi, err := adifToProto(lotwResponse, time.Now())
	if err != nil {
		writeError(500, "Failed parsing LotW data", err, w)
		return
	}
	fixLotwQsls(lotwAdi)
	if isFixCase {
		for _, qso := range lotwAdi.Qsos {
			fixCase(qso)
		}
	}
	contactsRef := firestoreClient.Collection("users").Doc(userToken.UID).Collection("contacts")
	fsContacts, err := getContacts(ctx, contactsRef)
	if err != nil {
		writeError(500, "Error fetching contacts from firestore", err, w)
		return
	}
	created, modified, noDiff := mergeQsos(fsContacts, lotwAdi, contactsRef, ctx)

	var report = map[string]int{}
	report["lotw"] = len(lotwAdi.Qsos)
	report["firestore"] = len(fsContacts)
	report["created"] = created
	report["modified"] = modified
	report["noDiff"] = noDiff
	marshal, _ := json.Marshal(report)
	_, _ = fmt.Fprint(w, string(marshal))
}

func removeEof(response string) string {
	// This non-conformant tag screws up the parser
	return strings.ReplaceAll(response, "<APP_LoTW_EOF>", "")
}

func fixLotwQsls(lotwAdi *adifpb.Adif) {
	// LotW puts their QSL in the ADIF fields where cards should go
	for _, qso := range lotwAdi.Qsos {
		qso.Lotw = qso.Card
		qso.Card = nil
	}
}

func getLotwCreds(ctx context.Context, firestoreClient *firestore.Client, userUid string) (string, string, error) {
	// TODO: pull creds from Google Cloud Secret Manager
	return "", "", nil
}
