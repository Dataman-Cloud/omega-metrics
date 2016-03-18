# Influxdb Docker image

We use the [tutum/influxdb](https://hub.docker.com/r/tutum/influxdb/).

## Running the container

Start your container:

```
docker run -d \
    --volume=/var/influxdb:/data
    -p 8083:5008 \
    -p 8086:5009 \
    -e "PRE_CREATE_DB=shurenyun"
    testregistry.dataman.io/zqdou/influxdb:0.9
```
