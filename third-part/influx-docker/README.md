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
