FROM grafana/grafana:2.6.0

ENV INFLUXDB_HOST influxdb
ENV INFLUXDB_PORT 8086
ENV INFLUXDB_USER root
ENV INFLUXDB_PASS root
ENV INFLUXDB_NAME shurenyun

RUN apt-get update -y && apt-get install curl -y

ADD grafana.ini /etc/grafana/grafana.ini 

COPY ./run.sh /run.sh
COPY ./dataman.json /dataman.json

EXPOSE 5010

ENTRYPOINT ["/run.sh"]
