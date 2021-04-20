package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"os"
	//"net/http/httputil"
	"cloud.google.com/go/logging"
	"github.com/gorilla/mux"
	"golang.org/x/net/http2"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

var (
	loggingClient *logging.Client

	//https://cloud.google.com/run/docs/reference/container-contract#env-vars
	K_SERVICE       = os.Getenv("K_SERVICE")
	K_REVISION      = os.Getenv("K_REVISION")
	K_CONFIGURATION = os.Getenv("K_CONFIGURATION")
)

const (
	signatureHeader   = "X-Tfe-Notification-Signature"
	key               = "password1"
	verificationToken = ""
)

type TFNotification struct {
	Message      string `json:"message,omitempty"`
	Trigger      string `json:"trigger,omitempty"`
	RunStatus    string `json:"run_status,omitempty"`
	RunUpdatedAt string `json:"run_updated_at,omitempty"`
	RunUpdatedBy string `json:"run_updated_by,omitempty"`
}
type TFCloudEvent struct {
	PayloadVersion              int              `json:"payload_version,omitempty"`
	NotificationConfigurationId string           `json:"notification_configuration_id,omitempty"`
	RunUrl                      string           `json:"run_url,omitempty"`
	RunId                       string           `json:"run_id,omitempty"`
	RunMessage                  string           `json:"run_message,omitempty"`
	RunCreatedAt                string           `json:"run_created_at,omitempty"`
	RunCreatedBy                string           `json:"run_created_by,omitempty"`
	WorkspaceId                 string           `json:"workspace_id,omitempty"`
	WorkspaceName               string           `json:"workspace_name,omitempty"`
	OrganizationName            string           `json:"organization_name,omitempty"`
	Notifications               []TFNotification `json:"notifications,omitempty"`
}

func validMAC(message, messageMAC, key []byte) bool {

	mac := hmac.New(sha512.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}

type contextKey string

const contextEventKey contextKey = "event"

func eventsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		signatureHeaderBytes, err := hex.DecodeString(r.Header.Get(signatureHeader))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if !validMAC(b, signatureHeaderBytes, []byte(key)) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		requestDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println(string(requestDump))

		event := &TFCloudEvent{}
		err = json.Unmarshal(b, event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(r.Context(), contextEventKey, event)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func fronthandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("/  called")
	val := r.Context().Value(contextEventKey).(*TFCloudEvent)

	for _, v := range val.Notifications {
		if v.Trigger == "verification" {
			fmt.Println("Got Verification Request")
			fmt.Fprint(w, "ok")
			return
		}
	}

	logger := loggingClient.Logger("terraform")
	monitoredResource := &monitoredres.MonitoredResource{
		Type: "generic_task",
		Labels: map[string]string{
			"project_id": "mineral-minutia-820",
			"namespace":  "default",
			"job":        K_SERVICE,
			"task_id":    K_REVISION,
			"location":   "us-central1-a",
		},
	}
	//commonResource := logging.CommonResource(&monitoredResource)
	logger.Log(logging.Entry{
		Payload:  val,
		Severity: logging.Info,
		Resource: monitoredResource,
		Labels:   map[string]string{
			"project_id": "mineral-minutia-820", 
			"location": "us-central1-a", 
			"node_id": "run_instance_id"
		},
	})
	defer logger.Flush()
	fmt.Fprint(w, "ok")
}

func main() {

	router := mux.NewRouter()
	router.Methods(http.MethodPost).Path("/").HandlerFunc(fronthandler)

	ctx := context.Background()
	var err error
	loggingClient, err = logging.NewClient(ctx, "mineral-minutia-820")
	if err != nil {
		fmt.Errorf("Failed to create logging client: %v", err)
	}

	var server *http.Server
	server = &http.Server{
		Addr:    ":8080",
		Handler: eventsMiddleware(router),
	}
	http2.ConfigureServer(server, &http2.Server{})
	fmt.Println("Starting Server..")
	err = server.ListenAndServe()
	fmt.Printf("Unable to start Server %v", err)

}
