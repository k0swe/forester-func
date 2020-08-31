// Package kellog provides a set of Cloud Functions samples.
package kellog

import (
	"cloud.google.com/go/firestore"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"fmt"
	"github.com/xylo04/kellog-qrz-sync/generated/adifpb"
	ql "github.com/xylo04/qrz-logbook"
	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"
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
	qrzAdi, err := adifToJson(qrzResponse.Adif)
	if err != nil {
		writeError(500, "Failed parsing QRZ.com data", err, w)
		return
	}
	contactsRef := firestoreClient.Collection("users").Doc(userToken.UID).Collection("contacts")
	fsContacts, err := getContacts(ctx, contactsRef)
	if err != nil {
		writeError(500, "Error fetching contacts from firestore", err, w)
		return
	}
	mergeQsos(fsContacts, qrzAdi, contactsRef, ctx)

	var report = map[string]int{}
	report["qrz"] = len(qrzAdi.Qsos)
	report["firestore"] = len(fsContacts)
	marshal, _ := json.Marshal(report)
	_, _ = fmt.Fprint(w, string(marshal))
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

type FirestoreQso struct {
	qsopb  adifpb.Qso
	docref *firestore.DocumentRef
}

func getContacts(ctx context.Context, contactsRef *firestore.CollectionRef) ([]FirestoreQso, error) {
	docItr := contactsRef.Documents(ctx)
	var retval = make([]FirestoreQso, 0, 100)
	for i := 0; ; i++ {
		qsoDoc, err := docItr.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		// I want to just qsoDoc.DataTo(&qso), but timestamps don't unmarshal
		buf := qsoDoc.Data()
		marshal, _ := json.Marshal(buf)
		var qso adifpb.Qso
		err = protojson.Unmarshal(marshal, &qso)
		if err != nil {
			log.Printf("Skipping qso %d: unmarshaling error: %v", i, err)
			continue
		}
		retval = append(retval, FirestoreQso{qso, qsoDoc.Ref})
	}
	return retval, nil
}

func mergeQsos(firebaseQsos []FirestoreQso, qrzAdi *adifpb.Adif, contactsRef *firestore.CollectionRef, ctx context.Context) {
	m := map[string]FirestoreQso{}
	for _, fsQso := range firebaseQsos {
		hash := hashQso(fsQso.qsopb)
		m[hash] = fsQso
	}

	for _, qrzQso := range qrzAdi.Qsos {
		hash := hashQso(*qrzQso)
		if _, ok := m[hash]; ok {
			// TODO: merge
			log.Printf("Found a match for %v on %v",
				qrzQso.ContactedStation.StationCall,
				qrzQso.TimeOn.String())
		} else {
			log.Printf("Creating QSO with %v on %v",
				qrzQso.ContactedStation.StationCall,
				qrzQso.TimeOn.String())
			jso, _ := protojson.Marshal(qrzQso)
			var buf map[string]interface{}
			err := json.Unmarshal(jso, &buf)
			if err != nil {
				log.Printf("Problem unmarshaling for create: %v", err)
				continue
			}
			_, err = contactsRef.NewDoc().Create(ctx, buf)
			if err != nil {
				log.Printf("Problem creating: %v", err)
				continue
			}
		}
	}
}

func hashQso(qsopb adifpb.Qso) string {
	payload := []byte(qsopb.ContactedStation.StationCall + qsopb.TimeOn.String())
	return fmt.Sprintf("%x", sha256.Sum256(payload))
}
