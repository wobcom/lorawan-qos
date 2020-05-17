package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

// // Location details.
// type Location struct {
// 	Latitude  float64 `json:"latitude"`
// 	Longitude float64 `json:"longitude"`
// 	Altitude  float64 `json:"altitude"`
// }
// 8000
// // RXInfo contains the RX information.
// type RXInfo struct {
// 	GatewayID lorawan.EUI64 `json:"gatewayID"`
// 	UplinkID  uuid.UUID     `json:"uplinkID"`
// 	Name      string        `json:"name"`
// 	Time      *time.Time    `json:"time,omitempty"`
// 	RSSI      int           `json:"rssi"`
// 	LoRaSNR   float64       `json:"loRaSNR"`
// 	Location  *Location     `json:"location"`
// }

// // TXInfo contains the TX information.
// type TXInfo struct {
// 	Frequency int `json:"frequency"`
// 	DR        int `json:"dr"`
// }

// // DataUpPayload represents a data-up payload.
// type DataUpPayload struct {
// 	ApplicationID   int64             `json:"applicationID,string"`
// 	ApplicationName string            `json:"applicationName"`
// 	DeviceName      string            `json:"deviceName"`
// 	DevEUI          lorawan.EUI64     `json:"devEUI"`
// 	RXInfo          []RXInfo          `json:"rxInfo,omitempty"`
// 	TXInfo          TXInfo            `json:"txInfo"`
// 	ADR             bool              `json:"adr"`
// 	FCnt            uint32            `json:"fCnt"`
// 	FPort           uint8             `json:"fPort"`
// 	Data            []byte            `json:"data"`
// 	Object          interface{}       `json:"object,omitempty"`
// 	Tags            map[string]string `json:"tags,omitempty"`
// 	Variables       map[string]string `json:"-"`
// }

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

//NewHTTPUplinkHandler is
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

			decoder := json.NewDecoder(r.Body)
			var upLinkEvent integration.UplinkEvent
			decodeError := decoder.Decode(&upLinkEvent)
			if decodeError != nil {
				log.Error(decodeError)
				code = http.StatusBadRequest
				http.Error(w, decodeError.Error(), code)
				return
			}

			processEventError := processEvent(&upLinkEvent)
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

func processEvent(ue *integration.UplinkEvent) error {
	log.Debugln(ue)
	dt, checkDT := ue.Tags["divice-type"]

	if !checkDT || dt != "network-qos" {
		return errors.New("wrong device-type")
	}
	prettyJSON, err := json.MarshalIndent(ue, "", "    ")
	if err != nil {
		log.Fatal("Failed to generate json", err)
	}
	fmt.Printf("%s\n", string(prettyJSON))

	return nil
}
