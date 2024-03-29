package v1

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type DebugHandler struct {
	Build string
	Log   *zap.SugaredLogger
	DB    *sqlx.DB
}

// Readiness checks if the database is ready and if not will return a 500 status.
// Do not respond by just returning an error because further up in the call
// stack it will interpret that as a non-trusted error.
func (dbug *DebugHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	}

	statusCode := http.StatusOK

	if err := response(w, statusCode, data); err != nil {
		dbug.Log.Errorln("READINESS: ", err)

	}

	dbug.Log.Infoln("READINESS: ", "statusCode: ", statusCode, " method: ", r.Method, " path: ", r.URL.Path, " remoteaddr: ", r.RemoteAddr)
}

// Liveness returns simple status info if the service is alive. If the
// app is deployed to a Kubernetes cluster, it will also return pod, node, and
// namespace details via the Downward API. The Kubernetes environment variables
// need to be set within your Pod/Deployment manifest.
func (dbug *DebugHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}

	data := struct {
		Status    string `json:"status,omitempty"`
		Build     string `json:"build,omitempty"`
		Host      string `json:"host,omitempty"`
		Name      string `json:"name,omitempty"`
		PodIP     string `json:"podIP,omitempty"`
		Node      string `json:"node,omitempty"`
		Namespace string `json:"namespace,omitempty"`
	}{
		Status:    "up",
		Build:     dbug.Build,
		Host:      host,
		Name:      os.Getenv("KUBERNETES_NAME"),
		PodIP:     os.Getenv("KUBERNETES_POD_IP"),
		Node:      os.Getenv("KUBERNETES_NODE_NAME"),
		Namespace: os.Getenv("KUBERNETES_NAMESPACE"),
	}

	statusCode := http.StatusOK
	if err := response(w, statusCode, data); err != nil {
		dbug.Log.Errorln("LIVENESS: ", err)
	}

	// THIS IS A FREE TIMER. WE COULD UPDATE THE METRIC GOROUTINE COUNT HERE.

	dbug.Log.Infoln("LIVENESS: ", " statusCode: ", statusCode, " method: ", r.Method, " path: ", r.URL.Path, " remoteaddr: ", r.RemoteAddr)
}

func response(w http.ResponseWriter, statusCode int, data any) error {

	// Convert the response value to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Set the content type and headers once we know marshaling has succeeded.
	w.Header().Set("Content-Type", "application/json")

	// Write the status code to the response.
	w.WriteHeader(statusCode)

	// Send the result back to the client.
	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}
