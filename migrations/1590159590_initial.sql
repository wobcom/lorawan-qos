-- +migrate Up
CREATE EXTENSION IF not exists timescaledb;
CREATE EXTENSION IF not exists postgis;

create table device (
    dev_eui bytea not null primary key,
    dev_name VARCHAR(50)
);

create table device_config (
    config_id serial not null,
    dev_eui bytea not null references device,
    config jsonb,
    recorded timestamp without time zone not null,
    primary key(config_id, recorded)
);
SELECT create_hypertable('device_config', 'recorded');

create table device_point (
    dev_gid serial not null,
    dev_eui  bytea not null references device,
    geom  geometry(POINT,4326),
    hdop real,
    sats smallint,
    recorded timestamp without time zone not null,
    primary key(dev_gid, recorded)
);
SELECT create_hypertable('device_point', 'recorded');

create table gateway (
    gw_eui bytea not null primary key,
    gw_name VARCHAR(50)
);

create table gateway_point (
    gw_gid serial not null,
    geom geometry(POINT,4326),
    gw_eui  bytea not null references gateway,
    recorded timestamp without time zone not null,
    primary key(gw_gid, recorded)
);
SELECT create_hypertable('gateway_point', 'recorded');

create table network_stat (
    tr_id serial not null,
    dev_eui  bytea not null references device,
    gw_eui bytea not null references gateway,
    fCnt integer,
    rssi real,
    snr real,
    frequency integer,
    dr smallint,
    gw_gid integer,
    dev_gid integer,
    recorded timestamp without time zone not null,
    -- foreign key (gw_gid, recorded) references gateway_point(gw_gid, recorded),
    -- foreign key (dev_gid, recorded) references device_point(dev_gid, recorded),
    primary key(tr_id, recorded)
);

SELECT create_hypertable('network_stat', 'recorded');

create table wifi_stat (
    wifi_id serial not null,
    dev_eui bytea references device,
    dev_count integer,
    recorded timestamp without time zone not null,
    primary key(wifi_id, recorded)
);
SELECT create_hypertable('wifi_stat', 'recorded');

create table ibis_line (
    line_id smallint primary key,
    line_name VARCHAR(50)
);

create table ibis_stop (
    stop_id smallint primary key,
    stop_name VARCHAR(50),
    stop_short_name VARCHAR(10),
    geom geometry(POINT,4326)
);

create table ibis_stat (
    ibis_id serial not null,
    line_id smallint references ibis_line,
    stop_id smallint references ibis_stop,
    recorded timestamp without time zone not null,
    primary key(ibis_id, recorded)
);
SELECT create_hypertable('ibis_stat', 'recorded');

-- +migrate Down
-- DROP EXTENSION IF EXISTS timescaledb;

drop table devices cascade;
drop table device_points;
drop table device_config;
drop table gateway cascade;
drop table gateway_points;
drop table tx_info;
drop table ibis_stat;
drop table ibis_line cascade;
drop table ibis_stop cascade;
drop table wifi_stat;
