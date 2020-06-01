package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/brocaar/lorawan"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"github.com/cridenour/go-postgis"
	"github.com/jmoiron/sqlx"
)

type Device struct {
	EUI  lorawan.EUI64 `db:"dev_eui"`
	Name string        `db:"dev_name"`
}

// GetDevice returns the device matching the given DevEUI.
// When forUpdate is set to true, then db must be a db transaction.
// When localOnly is set to true, no call to the network-server is made to
// retrieve additional device data.
func GetDevice(ctx context.Context, db sqlx.Queryer, devEUI lorawan.EUI64, forUpdate, localOnly bool) (Device, error) {
	var fu string
	if forUpdate {
		fu = " for update"
	}

	var d Device
	err := sqlx.Get(db, &d, "select * from device where dev_eui = $1"+fu, devEUI[:])
	if err != nil {
		return d, handlePSQLError(Select, err, "select error")
	}

	if localOnly {
		return d, nil
	}

	return d, nil
}

func NewDevice(EUI lorawan.EUI64, name string) *Device {
	return &Device{EUI: EUI, Name: name}
}

func InsertDevice(ctx context.Context, db sqlx.Queryer, d *Device) {
	res, err := db.Query(`insert into device(dev_eui, dev_name) values ($1, $2) ON CONFLICT DO NOTHING`, d.EUI[:], d.Name)
	if err != nil {
		log.Errorln("InsertDevice Error ", err)
	}
	res.Close()
}

type DeviceConfig struct {
	ID       int
	EUI      lorawan.EUI64
	Config   *json.RawMessage
	Recorded *time.Time
}

type GNSSMessage struct {
	Hdop      float32 `json:"hdop"`
	Altitude  float64 `json:"altitude"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Sats      int16   `json:"sats"`
}
type DevicePoint struct {
	ID       int             `db:"dev_gid"`
	EUI      lorawan.EUI64   `db:"dev_eui"`
	Geom     *postgis.PointS `db:"geom"`
	Hdop     float32         `db:"hdop"`
	Sats     int16           `db:"sats"`
	Recorded *time.Time      `db:"recorded"`
}

func NewDevicePoint(EUI lorawan.EUI64, latitude float64, longitude float64,
	hdop float32, sats int16, recorded *time.Time) *DevicePoint {
	return &DevicePoint{EUI: EUI,
		Geom:     &postgis.PointS{SRID: 4326, X: latitude, Y: longitude},
		Hdop:     hdop,
		Sats:     sats,
		Recorded: recorded,
	}
}

func InsertDevicePoint(ctx context.Context, db sqlx.Ext, dp *DevicePoint) {
	res, err := db.Query(`insert into device_point(dev_eui, geom, hdop, sats, recorded) values ($1, GeomFromEWKB($2), $3, $4, $5) RETURNING dev_gid`, dp.EUI[:], dp.Geom, dp.Hdop, dp.Sats, dp.Recorded)
	if err != nil {
		log.Errorln("CreatePoint Error ", err)
	} else {
		for res.Next() {
			scanErr := res.Scan(&dp.ID)
			if scanErr != nil {
				log.Errorln("Error Scanning ", scanErr)
			}
		}
		res.Close()
	}

}
