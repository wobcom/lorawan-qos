package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

var uplinkHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "uplink_seconds",
	Help:    "Time take to process uplink",
	Buckets: []float64{1, 2, 5, 6, 10}, //defining small buckets as this app should not take more than 1 sec to respond
}, []string{"code"}) // this will be partitioned by the HTTP code.

var (
	processEventCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "processEvent",
		Help: "Count of all processEvent requests",
	})
)

// NewHTTPUplinkHandler is
func NewHTTPUplinkHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			start := time.Now()
			defer r.Body.Close()
			code := http.StatusInternalServerError

			defer func() {
				httpDuration := time.Since(start)
				uplinkHistogram.WithLabelValues(fmt.Sprintf("%d", code)).Observe(httpDuration.Seconds())
			}()
			var payload []byte
			r.Body.Read(payload)
			ue, err := decodeUplink(payload)
			if err != nil {

			}
			processEventError := processEvent(ue)
			if processEventError != nil {
				log.Error(processEventError)
				code = http.StatusBadGateway
				http.Error(w, processEventError.Error(), code)
				return
			}
			code = http.StatusOK
			w.WriteHeader(code)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func connect(clientId string, uri *url.URL) mqtt.Client {
	opts := createClientOptions(clientId, uri)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	return client
}

func createClientOptions(clientId string, uri *url.URL) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("ssl://%s", uri.Host))
	opts.SetUsername(uri.User.Username())
	password, _ := uri.User.Password()
	opts.SetPassword(password)
	opts.SetClientID(clientId)
	return opts
}

func Listen(uri *url.URL, topic string) {
	client := connect("sub", uri)
	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		go doSubscribe(client, msg)
	})
}

func doSubscribe(client mqtt.Client, msg mqtt.Message) {
	log.Debugf("* [%s] %s\n", msg.Topic(), string(msg.Payload()))
	ue, decodeError := decodeUplink(msg.Payload())
	if decodeError != nil {
		log.Error("decoding error")
	}
	processError := processEvent(ue)
	if processError != nil {
		log.Error("process error ", processError)
	}
}

func decodeUplink(payload []byte) (*DataUpPayload, error) {
	var dp DataUpPayload
	decodeError := json.Unmarshal(payload, &dp)
	return &dp, decodeError
}

func processEvent(ue *DataUpPayload) error {
	log.Infoln(ue)
	dt, checkDT := ue.Tags["device-type"]
	if !checkDT || dt != "network-qos" {
		return errors.New("wrong device-type")
	}

	rxInfo := ue.RXInfo
	log.Infoln("txInfo[0].Location", rxInfo[0].Location)
	log.Infoln("payload-object", ue.Object)
	return nil
}
