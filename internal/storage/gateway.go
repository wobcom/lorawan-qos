package storage

import (
	"context"
	"time"

	"github.com/brocaar/lorawan"
	"github.com/cridenour/go-postgis"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type Gateway struct {
	EUI  lorawan.EUI64 `db:"gw_eui"`
	Name string        `db:"gw_name"`
}

func NewGateway(EUI lorawan.EUI64, name string) *Gateway {
	return &Gateway{EUI: EUI, Name: name}
}

func InsertGateway(ctx context.Context, db sqlx.Queryer, gw *Gateway) {
	res, err := db.Query(`insert into gateway(gw_eui, gw_name) values ($1, $2) ON CONFLICT DO NOTHING`, gw.EUI[:], gw.Name)
	if err != nil {
		log.Errorln("InsertGateway Error ", err)
	} else {
		res.Close()
		log.Infoln("Gateway Insertd")
	}
}

type GatewayPoint struct {
	ID       int             `db:"gw_gid"`
	EUI      lorawan.EUI64   `db:"gw_eui"`
	Geom     *postgis.PointS `db:"geom"`
	Recorded *time.Time      `db:"recorded"`
}

func NewGatewayPoint(EUI lorawan.EUI64, latitude float64, longitude float64, recorded *time.Time) *GatewayPoint {
	return &GatewayPoint{EUI: EUI,
		Geom:     &postgis.PointS{SRID: 4326, X: longitude, Y: latitude},
		Recorded: recorded,
	}
}

func InsertGatewayPoint(ctx context.Context, db sqlx.Queryer, gp *GatewayPoint) {
	if gp.Recorded == nil {
		rnow := time.Now()
		gp.Recorded = &rnow
	}
	res, err := db.Query(`insert into gateway_point(gw_eui, geom, recorded) values ($1, GeomFromEWKB($2), $3) RETURNING gw_gid`, gp.EUI[:], gp.Geom, gp.Recorded)
	if err != nil {
		log.Errorln("InsertGatewayPoint Error ", err)
	} else {
		for res.Next() {
			scanErr := res.Scan(&gp.ID)
			if scanErr != nil {
				log.Errorln("Error Scanning ", scanErr)
			}
		}
		res.Close()
	}

}
