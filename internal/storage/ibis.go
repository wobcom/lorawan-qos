package storage

import (
	"context"
	"time"

	"github.com/cridenour/go-postgis"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type IbisMessage struct {
	LineID int `json:"line"`
	StopID int `json:"nextStop"`
}

type IbisLine struct {
	ID   int
	Name string
}

type IbisStop struct {
	ID        int             `db:"stop_id"`
	Name      string          `db:"stop_name"`
	ShortName string          `db:"stop_short_name"`
	Geom      *postgis.PointS `db:"geom"`
}

func NewIbisStop(id int, name string, shortName string, latitude float64, longitude float64) *IbisStop {
	return &IbisStop{
		ID:        id,
		Name:      name,
		ShortName: shortName,
		Geom:      &postgis.PointS{SRID: 4326, X: latitude, Y: longitude},
	}
}

func InsertIbisStop(ctx context.Context, db sqlx.Queryer, ibisStop *IbisStop) {
	res, err := db.Query(`insert into ibis_stop(stop_id, stop_name, stop_short_name, geom) values ($1, $2, $3, GeomFromEWKB($4))`, ibisStop.ID, ibisStop.Name, ibisStop.ShortName, ibisStop.Geom)
	if err != nil {
		log.Errorln("InsertGateway Error ", err)
	} else {
		res.Close()
	}
}

func GetIbisStop(ctx context.Context, db sqlx.Queryer, ID int, forUpdate, localOnly bool) (IbisStop, error) {
	var fu string
	if forUpdate {
		fu = " for update"
	}

	var d IbisStop
	err := sqlx.Get(db, &d, "select * from ibis_stop where stop_id = $1"+fu, ID)
	if err != nil {
		return d, handlePSQLError(Select, err, "select error")
	}

	if localOnly {
		return d, nil
	}

	return d, nil
}

type IbisStat struct {
	ID       int        `db:"ibis_id"`
	LineID   int        `db:"line_id"`
	StopID   int        `db:"stop_id"`
	Recorded *time.Time `db:"recorded"`
}

func NewIbisStat(lid int, sid int, r *time.Time) *IbisStat {
	return &IbisStat{LineID: lid, StopID: sid, Recorded: r}
}

func InsertIbisStat(ctx context.Context, db sqlx.Queryer, ibisS *IbisStat) {
	res, err := db.Query(`insert into ibis_stat(line_id, stop_id, recorded) values ($1, $2, $3) RETURNING ibis_id`, ibisS.LineID, ibisS.StopID, ibisS.Recorded)
	if err != nil {
		log.Errorln("InsertGateway Error ", err)
	} else {
		for res.Next() {
			scanErr := res.Scan(&ibisS.ID)
			if scanErr != nil {
				log.Errorln("Error Scanning ", scanErr)
			}
		}
		res.Close()
	}

}
