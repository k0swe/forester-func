package kellog

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/antihax/optional"
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

	lotwUser, lotwPass, err := getLotwCreds(ctx, userToken.UID)
	if err != nil {
		writeError(500, "Error fetching LotW creds", err, w)
		return
	}
	lotwResponse, err := lotw.Query(lotwUser, lotwPass, &lotw.QueryOpts{
		QsoQsl: optional.NewInterface(lotw.NO),
		// TODO: Last query time should be stored and fast-forward only
		QsoQsorxsince: optional.NewString("1970-01-01"),
		QsoMydetail:   optional.NewInterface(lotw.YES),
	})
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

func getLotwCreds(ctx context.Context, userUid string) (string, string, error) {
	secretStore := NewSecretStore(ctx)
	username, err := secretStore.fetchSecret(userUid + "_lotw_username")
	if err != nil {
		return "", "", err
	}
	password, err := secretStore.fetchSecret(userUid + "_lotw_password")
	if err != nil {
		return "", "", err
	}
	return username, password, nil
}
