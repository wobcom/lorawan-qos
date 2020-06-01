package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"network-qos/internal/integration"
	"network-qos/internal/storage"
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

func decodeUplink(payload []byte) (*integration.DataUpPayload, error) {
	var dp integration.DataUpPayload
	decodeError := json.Unmarshal(payload, &dp)
	return &dp, decodeError
}

func processEvent(ue *integration.DataUpPayload) error {
	ctx := context.Background()
	dt, checkDT := ue.Tags["device-type"]
	if !checkDT || dt != "network-qos" {
		return errors.New("wrong device-type")
	}

	devEUI := ue.DevEUI

	d := storage.Device{EUI: devEUI, Name: ue.DeviceName}
	storage.InsertDevice(ctx, storage.DB(), &d)

	jsonbody, err := json.Marshal(ue.Object)
	if err != nil {
		log.Errorln("Marshalling Error")
	}

	var dp *storage.DevicePoint
	n := time.Now()
	switch ue.FPort {
	case 1:
		var wifiMessage storage.WifiMessage
		WifiMessageErr := json.Unmarshal(jsonbody, &wifiMessage)
		if WifiMessageErr != nil {
			log.Errorln("No Wifi Message ", WifiMessageErr)
		}
		wst := storage.NewWiFiStat(devEUI, wifiMessage.Count, &n)
		storage.InsertWifiStat(ctx, storage.DB(), wst)
	case 4:
		var gnssMessage storage.GNSSMessage
		GNSSMessageErr := json.Unmarshal(jsonbody, &gnssMessage)
		if GNSSMessageErr != nil {
			log.Errorln("No GNSS Message ", GNSSMessageErr)
		}
		log.Infoln("GNSS Message ", gnssMessage)
		dp = storage.NewDevicePoint(devEUI, gnssMessage.Latitude,
			gnssMessage.Longitude, gnssMessage.Hdop, gnssMessage.Sats, &n)
		storage.InsertDevicePoint(ctx, storage.DB(), dp)
	case 12:
		var ibisMessage storage.IbisMessage
		IbisMessageErr := json.Unmarshal(jsonbody, &ibisMessage)
		if IbisMessageErr != nil {
			log.Errorln("No Ibis Message ", IbisMessageErr)
		}
		isStat := storage.NewIbisStat(ibisMessage.LineID, ibisMessage.StopID, &n)
		storage.InsertIbisStat(ctx, storage.DB(), isStat)
	default:
		log.Infoln("Unknown payload")
	}

	trInfo := make([]storage.TransmissionInfo, len(ue.RXInfo))
	for i := 0; i < len(ue.RXInfo); i++ {
		rxInfo := ue.RXInfo[i]

		gw := storage.NewGateway(rxInfo.GatewayID, rxInfo.Name)
		storage.InsertGateway(ctx, storage.DB(), gw)

		gp := storage.NewGatewayPoint(rxInfo.GatewayID, rxInfo.Location.Latitude, rxInfo.Location.Longitude, rxInfo.Time)
		storage.InsertGatewayPoint(ctx, storage.DB(), gp)

		trInfo[i].DeviceEUI = devEUI
		trInfo[i].GatewayLoc = gp
		trInfo[i].GatewayEUI = gw.EUI
		trInfo[i].Recorded = gp.Recorded
		trInfo[i].LoRaSNR = rxInfo.LoRaSNR
		trInfo[i].RSSI = rxInfo.RSSI
		trInfo[i].FCnt = ue.FCnt
		trInfo[i].DR = ue.TXInfo.DR
		trInfo[i].Frequency = ue.TXInfo.Frequency
		trInfo[i].DeviceLoc = dp
		storage.InsertTransmissionInfo(ctx, storage.DB(), &trInfo[i])
	}
	return nil
}
