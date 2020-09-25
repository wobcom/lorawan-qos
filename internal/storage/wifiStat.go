package storage

import (
	"context"
	"time"

	"github.com/brocaar/lorawan"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type WifiMessage struct {
	Count int `json:"wifi"`
}

type WifiStat struct {
	ID       int           `db:"wifi_id"`
	DevEUI   lorawan.EUI64 `db:"dev_eui"`
	Count    int           `db:"dev_count"`
	Recorded *time.Time    `db:"recorded"`
}

func NewWiFiStat(EUI lorawan.EUI64, c int, r *time.Time) *WifiStat {
	return &WifiStat{DevEUI: EUI, Count: c, Recorded: r}
}

func InsertWifiStat(ctx context.Context, db sqlx.Queryer, wst *WifiStat) {
	if wst.Recorded == nil {
		rnow := time.Now()
		wst.Recorded = &rnow
	}

	res, err := db.Query(`insert into wifi_stat(dev_eui, dev_count, recorded) values ($1, $2, $3) RETURNING wifi_id`, wst.DevEUI[:], wst.Count, wst.Recorded)
	if err != nil {
		log.Errorln("InsertGateway Error ", err)
	} else {
		for res.Next() {
			scanErr := res.Scan(&wst.ID)
			if scanErr != nil {
				log.Errorln("Error Scanning ", scanErr)
			}
		}
		res.Close()
	}

}
