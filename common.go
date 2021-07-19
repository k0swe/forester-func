package forester

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"fmt"
	adifpb "github.com/k0swe/adif-json-protobuf/go"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/encoding/protojson"
	"net/http"
	"os"
	"strings"
)

// GCP_PROJECT is a user-set environment variable.
var projectID = os.Getenv("GCP_PROJECT")

func SetupLogging(ctx context.Context) {
	//output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	//output.FormatTimestamp = func(i interface{}) string {
	//	return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	//}
	//output.FormatLevel = func(i interface{}) string {
	//	return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	//}
	//output.FormatMessage = func(i interface{}) string {
	//	return fmt.Sprintf("***%s****", i)
	//}
	//output.FormatFieldName = func(i interface{}) string {
	//	return fmt.Sprintf("%s:", i)
	//}
	//output.FormatFieldValue = func(i interface{}) string {
	//	return strings.ToUpper(fmt.Sprintf("%s", i))
	//}
	//
	//log.Logger = zerolog.New(output).With().Timestamp().Logger()
}

// Write CORS headers to the response. Returns true if this is an OPTIONS request; false otherwise.
func handleCorsOptions(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if strings.Contains(origin, "forester.radio") || strings.Contains(origin, "localhost") {
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

func writeError(statusCode int, message string, err error, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	_, _ = fmt.Fprintf(w, message+": %v", err)
	if statusCode >= 500 {
		log.Error().Err(err).Msg(message)
	}
}

func ParseFirestoreQso(qsoDoc *firestore.DocumentSnapshot) (FirestoreQso, error) {
	// I want to just qsoDoc.DataTo(&qso), but timestamps don't unmarshal
	buf := qsoDoc.Data()
	marshal, _ := json.Marshal(buf)
	var qso adifpb.Qso
	err := protojson.Unmarshal(marshal, &qso)
	return FirestoreQso{&qso, qsoDoc.Ref}, err
}
