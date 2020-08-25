// Package kellog provides a set of Cloud Functions samples.
package kellog

import (
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"fmt"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"os"
)

// GCP_PROJECT is a user-set environment variable.
var projectID = os.Getenv("GCP_PROJECT")

var app *firebase.App
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
	idToken := ""
	if len(strings) > 0 {
		idToken = strings[0]
	}
	log.Printf("token param is '%v'", idToken)
	if idToken == "" {
		w.WriteHeader(403)
		return
	}
	userToken, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		w.WriteHeader(500)
		_, _ = fmt.Fprintf(w, "Failed VerifyIDTokenAndCheckRevoked: %v", err)
		log.Printf("Failed VerifyIDTokenAndCheckRevoked: %v", err)
		return
	}
	log.Printf("Token UID is %v", userToken.UID)
	conf := &firebase.Config{ProjectID: projectID}
	userApp, err := firebase.NewApp(ctx, conf, option.WithTokenSource(
		oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: idToken,
			})))
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v", err)
		return
	}
	firestoreClient, err := userApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error getting firestoreClient: %v", err)
		return
	}

	docSnapshot, err := firestoreClient.Collection("users").Doc(userToken.UID).Get(ctx)
	if err != nil {
		w.WriteHeader(500)
		_, _ = fmt.Fprintf(w, "Failed getting test data: %v", err)
		log.Printf("Failed getting test data: %v", err)
		return
	}
	m := docSnapshot.Data()
	_, _ = fmt.Fprintf(w, "Document data: %#v\n", m)
}
