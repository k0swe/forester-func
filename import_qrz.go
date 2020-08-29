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
	ql "github.com/xylo04/qrz-logbook"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"os"
	"strings"
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

// Import QSOs from QRZ logbook and merge into Firestore. Called via GCP Cloud Functions.
func ImportQrz(w http.ResponseWriter, r *http.Request) {
	if handleCorsOptions(w, r) {
		return
	}

	ctx := context.Background()
	idToken, err := extractIdToken(w, r)
	if err != nil {
		return
	}
	userToken, err := verifyToken(idToken, ctx, w)
	if err != nil {
		return
	}
	firestoreClient, err := makeFirestoreClient(ctx, idToken, w)
	if err != nil {
		return
	}

	qrzApiKey, err := getQrzApiKey(w, firestoreClient, userToken, ctx)
	if err != nil {
		return
	}
	fetchResponse, err := ql.Fetch(&qrzApiKey)
	if err != nil {
		w.WriteHeader(500)
		_, _ = fmt.Fprintf(w, "Failed getting QRZ.com data: %v", err)
		log.Printf("Failed getting QRZ.com data: %v", err)
		return
	}
	log.Printf("Got QRZ.com data, count %d", fetchResponse.Count)
	enc := json.NewEncoder(w)
	_ = enc.Encode(fetchResponse)
}

// Write CORS headers to the response. Returns true if this is an OPTIONS request; false otherwise.
func handleCorsOptions(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if strings.Contains(origin, "log.k0swe.radio") || strings.Contains(origin, "localhost") {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func extractIdToken(w http.ResponseWriter, r *http.Request) (string, error) {
	idToken := strings.TrimSpace(r.Header.Get("Authorization"))
	if idToken == "" {
		w.WriteHeader(403)
		_, _ = fmt.Fprintf(w, "requests must be authenticated")
		return "", errors.New("requests must be authenticated")
	}
	idToken = strings.TrimPrefix(idToken, "Bearer ")
	return idToken, nil
}

func verifyToken(idToken string, ctx context.Context, w http.ResponseWriter) (*auth.Token, error) {
	userToken, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		w.WriteHeader(403)
		_, _ = fmt.Fprintf(w, "Failed to verify JWT: %v", err)
		log.Printf("Failed to verify JWT: %v", err)
		return nil, err
	}
	return userToken, nil
}

func makeFirestoreClient(ctx context.Context, idToken string, w http.ResponseWriter) (*firestore.Client, error) {
	conf := &firebase.Config{ProjectID: projectID}
	userApp, err := firebase.NewApp(ctx, conf, option.WithTokenSource(
		oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: idToken,
			})))
	if err != nil {
		w.WriteHeader(500)
		_, _ = fmt.Fprintf(w, "Error initializing Firebase user app: %v", err)
		log.Fatalf("Error initializing Firebase user app: %v", err)
		return nil, err
	}
	firestoreClient, err := userApp.Firestore(ctx)
	if err != nil {
		w.WriteHeader(500)
		_, _ = fmt.Fprintf(w, "Error getting firestoreClient: %v", err)
		log.Fatalf("Error getting firestoreClient: %v", err)
		return nil, err
	}
	return firestoreClient, nil
}

func getQrzApiKey(w http.ResponseWriter, firestoreClient *firestore.Client, userToken *auth.Token, ctx context.Context) (string, error) {
	docSnapshot, err := firestoreClient.Collection("users").Doc(userToken.UID).Get(ctx)
	if err != nil {
		w.WriteHeader(500)
		_, _ = fmt.Fprintf(w, "Failed getting Kellog user data: %v", err)
		log.Printf("Failed getting Kellog user data: %v", err)
		return "", err
	}
	qrzApiKey := fmt.Sprint(docSnapshot.Data()["qrzLogbookApiKey"])
	log.Printf("Got qrzApiKey %v", qrzApiKey)
	return qrzApiKey, nil
}
