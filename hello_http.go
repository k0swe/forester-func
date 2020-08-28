// Package kellog provides a set of Cloud Functions samples.
package kellog

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"errors"
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
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ctx := context.Background()
	idToken, err := extractIdToken(w, r)
	if err != nil {
		return
	}
	userToken, err := verifyToken(w, err, ctx, idToken)
	if err != nil {
		return
	}
	firestoreClient, err := makeFirestoreClient(ctx, idToken)
	if err != nil {
		return
	}

	docSnapshot, err := firestoreClient.Collection("users").Doc(userToken.UID).Get(ctx)
	if err != nil {
		w.WriteHeader(500)
		_, _ = fmt.Fprintf(w, "Failed getting test data: %v", err)
		log.Printf("Failed getting test data: %v", err)
		return
	}
	enc := json.NewEncoder(w)
	_ = enc.Encode(docSnapshot.Data())
}

func extractIdToken(w http.ResponseWriter, r *http.Request) (string, error) {
	strings := r.URL.Query()["token"]
	idToken := ""
	if len(strings) > 0 {
		idToken = strings[0]
	}
	if idToken == "" {
		w.WriteHeader(403)
		_, _ = fmt.Fprintf(w, "missing token")
		return "", errors.New("missing token")
	}
	return idToken, nil
}

func verifyToken(w http.ResponseWriter, err error, ctx context.Context, idToken string) (*auth.Token, error) {
	userToken, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		w.WriteHeader(403)
		_, _ = fmt.Fprintf(w, "Failed VerifyIDToken: %v", err)
		log.Printf("Failed VerifyIDToken: %v", err)
		return nil, err
	}
	return userToken, nil
}

func makeFirestoreClient(ctx context.Context, idToken string) (*firestore.Client, error) {
	conf := &firebase.Config{ProjectID: projectID}
	userApp, err := firebase.NewApp(ctx, conf, option.WithTokenSource(
		oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: idToken,
			})))
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v", err)
		return nil, err
	}
	firestoreClient, err := userApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error getting firestoreClient: %v", err)
		return nil, err
	}
	return firestoreClient, nil
}
