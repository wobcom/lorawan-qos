SELECT 
  encode(network_stat.dev_eui, 'hex') as dev_eui,
  encode(network_stat.gw_eui, 'hex') as gw_eui,
  gateway.gw_name,
  device.dev_name,
  network_stat.recorded,
  st_aslatlontext(gateway_point.geom) gw_location,
  st_aslatlontext(device_point.geom) dev_location,
  st_distance_sphere(gateway_point.geom, device_point.geom)  as distance
FROM network_stat
join gateway_point on network_stat.gw_gid = gateway_point.gw_gid
join device_point on network_stat.gw_gid = device_point.dev_gid
join gateway on network_stat.gw_eui = gateway.gw_eui
join device on network_stat.dev_eui = device.dev_eui
where network_stat.recorded > now() - interval '12 day'
and network_stat.dev_gid is not null
and network_stat.gw_gid is not null
and dev_name = 'GRA00'
ORDER BY network_stat.recorded DESC LIMIT 500;


select distinct on (gw.gw_name)
  gw.gw_name,
  encode(ns.gw_eui, 'hex') as gw_eui,
  st_aslatlontext(gp.geom) gw_location,
  dev.dev_name,
  encode(ns.dev_eui, 'hex') as dev_eui,
  st_aslatlontext(dp.geom) dev_location,
  dp.sats,
  dp.hdop,
  round(st_distancesphere(gp.geom, dp.geom)::numeric) as distance,
  ns.dr,
  ns.rssi,
  ns.frequency,
  ns.recorded as stat_recorded
from network_stat ns
join gateway_point gp using(gw_gid)
join device_point dp using(dev_gid)
join gateway gw on ns.gw_eui = gw.gw_eui
join device dev on ns.dev_eui = dev.dev_eui
where dp.hdop < 2.5
order by gw.gw_name, distance desc limit 500;


           gw_name           |      gw_eui      |          gw_location          | dev_name |     dev_eui      |         dev_location          | sats | hdop | distance | dr | rssi | frequency |       stat_recorded        
-----------------------------+------------------+-------------------------------+----------+------------------+-------------------------------+------+------+----------+----+------+-----------+----------------------------
 LABHIUPS01                  | aa555a0000000101 | 52°9'2.139"N 9°57'56.035"E    | GRA00    | 3c71bff18808feff | 52°9'5.494"N 9°57'44.687"E    |    4 | 2.43 |      239 |  5 |  -91 | 868100000 | 2020-06-08 03:19:06.284352
 LABWNTMTCDT01               | 00800000a000419a | 52°25'39.864"N 10°47'31.380"E | GRA01    | 3c71bff18af0feff | 52°25'41.106"N 10°46'34.842"E |    8 | 1.19 |     1066 |  2 |  -85 | 868500000 | 2020-06-08 14:14:55.482815
 PRODWNTMTCDT01_WNT          | 00800000a0004632 | 52°25'41.088"N 10°47'30.048"E | GRA01    | 3c71bff18af0feff | 52°26'7.166"N 10°42'54.997"E  |    9 |  1.2 |     5242 |  5 | -114 | 868500000 | 2020-06-08 15:08:12.262367
 PRODWOBMTCDT03_Nordsteimke  | 00800000a00037b2 | 52°23'44.268"N 10°49'23.880"E | GRA03    | 3c71bff18b60feff | 52°26'3.509"N 10°44'19.162"E  |    8 | 1.61 |     7173 |  2 | -117 | 868300000 | 2020-06-08 12:27:58.979156
 PRODWOBMTCDT04_Fallersleben | 00800000a0003abe | 52°24'30.852"N 10°43'5.016"E  | GRA03    | 3c71bff18b60feff | 52°18'7.423"N 10°36'54.954"E  |    8 | 1.22 |    13748 |  1 | -118 | 868500000 | 2020-06-08 15:35:07.439515
 PRODWOBMTCDT07_Teichbreite  | 00800000a00026a7 | 52°26'44.641"N 10°48'50.002"E | GRA01    | 3c71bff18af0feff | 52°20'48.282"N 10°47'9.643"E  |    9 | 0.91 |    11168 |  5 | -115 | 868100000 | 2020-06-08 06:28:40.265673
 PROWOBMTCDT06_Rathaus       | 00800000a00037a8 | 52°25'13.080"N 10°47'12.660"E | GRA03    | 3c71bff18b60feff | 52°21'59.076"N 10°43'2.424"E  |    8 | 1.25 |     7626 |  2 | -119 | 868100000 | 2020-06-08 12:21:40.154614