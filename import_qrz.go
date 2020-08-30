// Package kellog provides a set of Cloud Functions samples.
package kellog

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"fmt"
	ql "github.com/xylo04/qrz-logbook"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"
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
	idToken, err := extractIdToken(r)
	if err != nil {
		writeError(403, "Couldn't find authorization", err, w)
		return
	}
	userToken, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		writeError(403, "Couldn't verify authorization", err, w)
		return
	}
	firestoreClient, err := makeFirestoreClient(ctx, idToken)
	if err != nil {
		writeError(500, "Error creating firestore client", err, w)
		return
	}

	qrzApiKey, err := getQrzApiKey(ctx, firestoreClient, userToken.UID)
	if err != nil {
		writeError(500, "Error fetching QRZ API key from firestore", err, w)
		return
	}
	qrzResponse, err := ql.Fetch(&qrzApiKey)
	if err != nil {
		writeError(500, "Error fetching QRZ.com data", err, w)
		return
	}
	adi, err := adifToJson(qrzResponse.Adif)
	if err != nil {
		writeError(500, "Failed parsing QRZ.com data", err, w)
		return
	}
	_, _ = fmt.Fprintln(w, "[")
	for _, qso := range adi.Qsos {
		marshal, _ := protojson.Marshal(qso)
		_, _ = fmt.Fprint(w, string(marshal))
		_, _ = fmt.Fprintln(w, ",")
	}
	_, _ = fmt.Fprintln(w, "]")
}

func writeError(statusCode int, message string, err error, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	_, _ = fmt.Fprintf(w, message+": %v", err)
	if statusCode >= 500 {
		log.Fatalf(message+": %v", err)
	}
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

func extractIdToken(r *http.Request) (string, error) {
	idToken := strings.TrimSpace(r.Header.Get("Authorization"))
	if idToken == "" {
		return "", errors.New("requests must be authenticated with a Firebase JWT")
	}
	idToken = strings.TrimPrefix(idToken, "Bearer ")
	return idToken, nil
}

func makeFirestoreClient(ctx context.Context, idToken string) (*firestore.Client, error) {
	conf := &firebase.Config{ProjectID: projectID}
	userApp, err := firebase.NewApp(ctx, conf, option.WithTokenSource(
		oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: idToken,
			})))
	if err != nil {
		return nil, err
	}
	firestoreClient, err := userApp.Firestore(ctx)
	if err != nil {
		return nil, err
	}
	return firestoreClient, nil
}

func getQrzApiKey(ctx context.Context, firestoreClient *firestore.Client, userUid string) (string, error) {
	docSnapshot, err := firestoreClient.Collection("users").Doc(userUid).Get(ctx)
	if err != nil {
		return "", err
	}
	qrzApiKey := fmt.Sprint(docSnapshot.Data()["qrzLogbookApiKey"])
	if qrzApiKey == "" {
		return "", errors.New("user hasn't set up their QRZ.com API key")
	}
	return qrzApiKey, nil
}
