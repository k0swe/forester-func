package forester

import (
	"context"
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
	firebasePath := string(m.Data) // Automatically decoded from base64.
	log.Printf("Got a new Firebase contact at path %s", firebasePath)
	return nil
}
