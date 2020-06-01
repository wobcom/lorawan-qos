package storage

import (
	"context"
	"time"

	"github.com/brocaar/lorawan"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type TransmissionInfo struct {
	ID         int           `db:"tr_id"`
	GatewayEUI lorawan.EUI64 `db:"gw_eui"`
	DeviceEUI  lorawan.EUI64 `db:"dev_eui"`
	DR         int           `db:"dr"`
	Frequency  int           `db:"frequency"`
	FCnt       uint32        `db:"fCnt"`
	Recorded   *time.Time    `db:"recorded"`
	RSSI       int           `db:"rssi"`
	LoRaSNR    float64       `db:"snr"`
	GatewayLoc *GatewayPoint `db:"gw_gid"`
	DeviceLoc  *DevicePoint  `db:"dev_gid"`
}

func InsertTransmissionInfo(ctx context.Context, db sqlx.Queryer, trInfo *TransmissionInfo) {
	log.Infoln("Insert ", trInfo)
	if trInfo.Recorded == nil {
		rnow := time.Now()
		trInfo.Recorded = &rnow
	}
	var deviceGID *int
	if trInfo.DeviceLoc != nil {
		deviceGID = &trInfo.DeviceLoc.ID
	}

	var gatewayGID *int
	if trInfo.GatewayLoc != nil {
		gatewayGID = &trInfo.GatewayLoc.ID
	}

	res, err := db.Query(`insert into transmission_info(dev_eui, gw_eui, fCnt, rssi, snr, frequency, dr, gw_gid, dev_gid, recorded) 
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9,$10) RETURNING tr_id`,
		trInfo.DeviceEUI[:], trInfo.GatewayEUI[:], trInfo.FCnt, trInfo.RSSI, trInfo.LoRaSNR,
		trInfo.Frequency, trInfo.DR, gatewayGID, deviceGID, trInfo.Recorded)
	if err != nil {
		log.Errorln("InsertGatewayPoint Error ", err)
	} else {
		for res.Next() {
			scanErr := res.Scan(&trInfo.ID)
			if scanErr != nil {
				log.Errorln("Error Scanning ", scanErr)
			}
		}
		res.Close()
	}

}
