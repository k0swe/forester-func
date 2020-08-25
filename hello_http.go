// Package kellog provides a set of Cloud Functions samples.
package kellog

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"fmt"
	"log"
	"net/http"
	"os"
)

// GCP_PROJECT is a user-set environment variable.
var projectID = os.Getenv("GCP_PROJECT")

// firestoreClient is a global client, initialized once per instance.
var firestoreClient *firestore.Client
var authClient *auth.Client

func init() {
	// Use the application default credentials
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v", err)
		return
	}

	firestoreClient, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error getting firestoreClient: %v", err)
		return
	}

	authClient, err = app.Auth(ctx)
	if err != nil {
		log.Fatalf("Error getting authClient: %v", err)
		return
	}
}

// Hello World function. Called via GCP Cloud Functions.
func HelloHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	strings := r.URL.Query()["token"]
	tokenString := ""
	if len(strings) > 0 {
		tokenString = strings[0]
	}
	log.Printf("token param is '%v'", tokenString)
	if tokenString == "" {
		w.WriteHeader(403)
		return
	}
	token, err := authClient.VerifyIDTokenAndCheckRevoked(ctx, tokenString)
	if err != nil {
		w.WriteHeader(500)
		_, _ = fmt.Fprintf(w, "Failed VerifyIDTokenAndCheckRevoked: %v", err)
		log.Printf("Failed VerifyIDTokenAndCheckRevoked: %v", err)
		return
	}
	log.Printf("Token UID is %v", token.UID)

	docSnapshot, err := firestoreClient.Collection("test").Doc("3bjCgRjsJB1BK3X1Zy8s").Get(ctx)
	if err != nil {
		w.WriteHeader(500)
		_, _ = fmt.Fprintf(w, "Failed getting test data: %v", err)
		log.Printf("Failed getting test data: %v", err)
		return
	}
	m := docSnapshot.Data()
	_, _ = fmt.Fprintf(w, "Document data: %#v\n", m)
}
