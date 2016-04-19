# Influxdb Docker image

Based on the [tutum/influxdb](https://hub.docker.com/r/tutum/influxdb/), update the reserver ports in the configuration file.

## Running the container

Start your container:

```
docker run -d \
    --volume=/data/volumn/influxdb:/data
    -p 5007:5007 \
    -p 5008:5008 \
    -e "PRE_CREATE_DB=shurenyun"
    testregistry.dataman.io/zqdou/influxdb:0.10
```

## Enable the UDP listener

Refer to [Influxdb: UDP Service](https://docs.influxdata.com/influxdb/v0.12/write_protocols/udp/)

Add this line in the file /etc/sysctl.conf
```
net.core.rmem_max=8388608
```
