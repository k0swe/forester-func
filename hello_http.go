// Package kellog provides a set of Cloud Functions samples.
package kellog

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"fmt"
	"log"
	"net/http"
	"os"
)

// GCP_PROJECT is a user-set environment variable.
var projectID = os.Getenv("GCP_PROJECT")

// client is a global Firestore client, initialized once per instance.
var client *firestore.Client

func init() {
	// Use the application default credentials
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v", err)
		return
	}

	client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error getting Firestore client: %v", err)
		return
	}
}

// Hello World function. Called via GCP Cloud Functions.
func HelloHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	docSnapshot, err := client.Collection("test").Doc("3bjCgRjsJB1BK3X1Zy8s").Get(ctx)
	if err != nil {
		w.WriteHeader(500)
		_, _ = fmt.Fprintf(w, "Failed getting test data: %v", err)
		log.Printf("Failed getting test data: %v", err)
		return
	}
	m := docSnapshot.Data()
	_, _ = fmt.Fprintf(w, "Document data: %#v\n", m)
}
