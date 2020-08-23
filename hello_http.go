// Package kellog provides a set of Cloud Functions samples.
package kellog

import (
	"context"
	firebase "firebase.google.com/go"
	"fmt"
	"log"
	"net/http"
)

// Called via GCP Cloud Functions
func HelloHTTP(w http.ResponseWriter, r *http.Request) {

	// Use the application default credentials
	projectID := "k0swe-kellog"
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		w.WriteHeader(500)
		_, _ = fmt.Fprintf(w, "Error initializing Firebase app: %v", err)
		log.Printf("Error initializing Firebase app: %v", err)
		return
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		w.WriteHeader(500)
		_, _ = fmt.Fprintf(w, "Error getting Firestore client: %v", err)
		log.Printf("Error getting Firestore client: %v", err)
		return
	}
	defer client.Close()

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
