FROM tutum/influxdb:0.10
MAINTAINER Zhongqiang Dou <zqdou@dataman-inc.com>

ADD config.toml /config/config.toml
ADD run.sh /run.sh
RUN chmod +x /*.sh

# Admin server WebUI
EXPOSE 5008

# HTTP API
EXPOSE 5009

# Raft port (for clustering, don't expose publicly!)
#EXPOSE 8090

# Protobuf port (for clustering, don't expose publicly!)
#EXPOSE 8099

VOLUME ["/data"]

CMD ["/run.sh"]
