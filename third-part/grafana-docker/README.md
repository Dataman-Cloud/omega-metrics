# Grafana Docker image

Base on the [grafana/grafana](https://github.com/grafana/grafana-docker).
This image start the service with the data source and dashboard creation.

## Running the container

Start your container:

```
docker run -d \
   -v /opt/data/grafana:/var/lib/grafana \
   -p 5010:5010 \
   -e "GF_SECURITY_ADMIN_PASSWORD=dataman" \
   -e "GF_AUTH_ANONYMOUS_ENABLED=true" \
   -e "INFLUXDB_HOST=192.168.1.189" \
   -e "INFLUXDB_PORT=5008" \
   -e "INFLUXDB_USER=root" \
   -e "INFLUXDB_PASS=root" \
   -e "INFLUXDB_NAME=shurenyun" \
   testregistry.dataman.io/zqdou/grafana:2.6.0.4
```
