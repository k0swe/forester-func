package forester

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"fmt"
	"github.com/k0swe/qrz-api"
	"log"
)

// PubSubMessage is the payload of a Pub/Sub event.
// See the documentation for more details:
// https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// FillNewQsoFromQrz listens to Pub/Sub for new contacts in Firestore, and fills
// in missing QSO details for the contacted station from QRZ.com.
func FillNewQsoFromQrz(ctx context.Context, m PubSubMessage) error {
	var psMap map[string]string
	err := json.Unmarshal(m.Data, &psMap)
	if err != nil {
		return err
	}
	logbookId := psMap["logbookId"]
	contactId := psMap["contactId"]
	firebasePath := fmt.Sprintf("logbooks/%s/contacts/%s", logbookId, contactId)
	log.Printf("Got a new Firebase QSO at path %s", firebasePath)

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	defer client.Close()
	contact, err := client.Doc(firebasePath).Get(ctx)
	if err != nil {
		return err
	}
	qso, err := ParseFirestoreQso(contact)
	if err != nil {
		return err
	}
	contactedStationCall := qso.qsopb.ContactedStation.StationCall

	qrzUser, qrzPass, err := getQrzCreds(ctx, logbookId)
	if err != nil {
		return err
	}

	log.Printf("Querying QRZ.com for %v", contactedStationCall)
	lookupResp, err := qrz.Lookup(&qrzUser, &qrzPass, &contactedStationCall)
	if err != nil {
		return err
	}
	log.Printf("QRZ.com lookup: %v is %v", lookupResp.Callsign.Call, lookupResp.Callsign.Name)

	// TODO: copy fields into ADIFPB
	// TODO: fixcase on ADIFPB
	// TODO: merge
	// TODO: save

	return nil
}

func getQrzCreds(ctx context.Context, logbookId string) (string, string, error) {
	secretStore := NewSecretStore(ctx)
	username, err := secretStore.FetchSecret(logbookId, qrzUsername)
	if err != nil {
		return "", "", err
	}
	password, err := secretStore.FetchSecret(logbookId, qrzPassword)
	if err != nil {
		return "", "", err
	}
	return username, password, nil
}
