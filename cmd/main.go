package main

import (
	"net/http"
	"net/url"
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

func main() {
	http.Handle("/metrics", promhttp.Handler())
	uri, err := url.Parse(getEnv("ENDPOINT", "mqtt://network-qos:password@localhost:1883/application/1337/device/+/rx"))
	if err != nil {
		log.Fatal(err)
	}

	topic := uri.Path[1:len(uri.Path)]
	log.Debugf("topic %s \n", topic)

	go app.Listen(uri, topic)

	log.Info("Starting Network-QoS-Service")
	http.ListenAndServe(":8080", nil)
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
