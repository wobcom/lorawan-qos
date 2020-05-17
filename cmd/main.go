package main

import (
	"net/http"
	"network-qos/pkg/app"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	sendRequestErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "send_request_error",
		Help: "Count of all HTTP requests to FeedbackNow",
	})
)

// // NewConfig creates a Network-QoS configuration with defaults
// func NewConfig() *netqos.Config {
// 	return &netqos.Config{
// 		Endpoint: getEnv("ENDPOINT", "http://"),
// 		Timeout:  5 * time.Second,
// 	}
// }

func main() {
	// c := NewConfig()
	// log.SetLevel(log.DebugLevel)
	// log.Debug(c)

	// fc := netqos.NewClient(c)

	http.HandleFunc("/uplink", app.NewHTTPUplinkHandler())
	http.Handle("/metrics", promhttp.Handler())

	http.ListenAndServe(":8080", nil)
	log.Info("Starting Network-QoS-Service")
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
