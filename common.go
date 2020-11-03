package kellog

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/secretmanager/apiv1"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/imdario/mergo"
	"github.com/jinzhu/copier"
	adifpb "github.com/k0swe/adif-json-protobuf/go"
	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// GCP_PROJECT is a user-set environment variable.
var projectID = os.Getenv("GCP_PROJECT")

func getUserFirestore(w http.ResponseWriter, r *http.Request) (context.Context, *auth.Token, *firestore.Client, bool, error) {
	// Use the application default credentials
	ctx := context.Background()
	if projectID == "" {
		panic("GCP_PROJECT is not set")
	}
	conf := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		writeError(500, "Error initializing Firebase app", err, w)
		return nil, nil, nil, true, err
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		writeError(500, "Error getting authClient", err, w)
		return nil, nil, nil, true, err
	}
	if handleCorsOptions(w, r) {
		return nil, nil, nil, true, nil
	}

	idToken, err := extractIdToken(r)
	if err != nil {
		writeError(403, "Couldn't find authorization", err, w)
		return nil, nil, nil, true, err
	}
	userToken, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		writeError(403, "Couldn't verify authorization", err, w)
		return nil, nil, nil, true, err
	}
	firestoreClient, err := makeFirestoreClient(ctx, idToken)
	if err != nil {
		writeError(500, "Error creating firestore client", err, w)
		return nil, nil, nil, true, err
	}
	return ctx, userToken, firestoreClient, false, err
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

type FirestoreQso struct {
	qsopb  *adifpb.Qso
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
		retval = append(retval, FirestoreQso{&qso, qsoDoc.Ref})
	}
	return retval, nil
}

func mergeQsos(firebaseQsos []FirestoreQso, remoteAdi *adifpb.Adif, contactsRef *firestore.CollectionRef, ctx context.Context) (int, int, int) {
	var created = 0
	var modified = 0
	var noDiff = 0
	m := map[string]FirestoreQso{}
	for _, fsQso := range firebaseQsos {
		hash := hashQso(fsQso.qsopb)
		m[hash] = fsQso
	}

	for _, remoteQso := range remoteAdi.Qsos {
		hash := hashQso(remoteQso)
		if _, ok := m[hash]; ok {
			diff := mergeQso(m[hash].qsopb, remoteQso)
			if diff {
				log.Printf("Updating QSO with %v on %v",
					remoteQso.ContactedStation.StationCall,
					remoteQso.TimeOn.String())
				err := update(m[hash].qsopb, m[hash].docref, ctx)
				if err != nil {
					continue
				}
				modified++
			} else {
				log.Printf("No difference for QSO with %v on %v",
					remoteQso.ContactedStation.StationCall,
					remoteQso.TimeOn.String())
				noDiff++
			}
		} else {
			log.Printf("Creating QSO with %v on %v",
				remoteQso.ContactedStation.StationCall,
				remoteQso.TimeOn.String())
			err := create(remoteQso, contactsRef, ctx)
			if err != nil {
				continue
			}
			created++
		}
	}
	return created, modified, noDiff
}

func hashQso(qsopb *adifpb.Qso) string {
	timeOn, _ := ptypes.Timestamp(qsopb.TimeOn)
	// Some providers (QRZ.com) only have minute precision
	timeOn = timeOn.Truncate(time.Minute)
	payload := []byte(qsopb.LoggingStation.StationCall +
		qsopb.ContactedStation.StationCall +
		strconv.FormatInt(timeOn.Unix(), 10))
	return fmt.Sprintf("%x", sha256.Sum256(payload))
}

// Given two QSO objects, replace missing values in `base` with those from `backfill`. Values
// already present in `base` should be preserved.
func mergeQso(base *adifpb.Qso, backfill *adifpb.Qso) bool {
	original := &adifpb.Qso{}
	_ = copier.Copy(original, base)
	cleanQsl(base)
	cleanQsl(backfill)
	_ = mergo.Merge(base, backfill)
	return !proto.Equal(original, base)
}

func create(remoteQso *adifpb.Qso, contactsRef *firestore.CollectionRef, ctx context.Context) error {
	buf, err := qsoToJson(remoteQso)
	if err != nil {
		log.Printf("Problem unmarshaling for create: %v", err)
		return err
	}
	_, err = contactsRef.NewDoc().Create(ctx, buf)
	if err != nil {
		log.Printf("Problem creating: %v", err)
		return err
	}
	return nil
}

func update(remoteQso *adifpb.Qso, ref *firestore.DocumentRef, ctx context.Context) error {
	buf, err := qsoToJson(remoteQso)
	if err != nil {
		log.Printf("Problem unmarshaling for update: %v", err)
		return err
	}
	_, err = ref.Set(ctx, buf)
	if err != nil {
		log.Printf("Problem updating: %v", err)
		return err
	}
	return nil
}

func qsoToJson(qso *adifpb.Qso) (map[string]interface{}, error) {
	jso, _ := protojson.Marshal(qso)
	var buf map[string]interface{}
	err := json.Unmarshal(jso, &buf)
	return buf, err
}

func fetchSecret(key string, client *secretmanager.Client, ctx context.Context) (string, error) {
	fullKey := "projects/" + projectID + "/secrets/" + key + "/versions/latest"
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fullKey,
	}
	secretResp, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return "", err
	}
	return string(secretResp.GetPayload().GetData()), nil
}

func cleanQsl(qso *adifpb.Qso) {
	// If QSO has LotW QSL with status=N or date=0001-01-01T00:00:00Z,
	// remove those to make way for the merge
	if qso.Lotw != nil {
		l := qso.Lotw
		if l.SentStatus == "N" {
			l.SentStatus = ""
		}
		if l.SentDate != nil &&
			(l.SentDate.Seconds == -62135596800 || l.SentDate.Seconds == 0) {
			l.SentDate = nil
		}
		if l.ReceivedStatus == "N" {
			l.ReceivedStatus = ""
		}
		if l.ReceivedDate != nil &&
			(l.ReceivedDate.Seconds == -62135596800 || l.ReceivedDate.Seconds == 0) {
			l.ReceivedDate = nil
		}
	}
}
