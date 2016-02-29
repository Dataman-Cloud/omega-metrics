FROM testregistry.dataman.io/zqdou/ubuntu-go:1.5.1

MAINTAINER zqdou zqdou@dataman-inc.com

RUN apt-get install gcc automake autoconf libtool git -y

RUN mkdir -p /var/log/omega
RUN mkdir -p /etc/omega

RUN mkdir -p /usr/lib/go/src/github.com/Dataman-Cloud

ENV GOPATH=/usr/share/go

RUN go get -u github.com/FiloSottile/gvt

ENV GO15VENDOREXPERIMENT=1

RUN echo 'alias ggvt='$GOPATH/src/github.com/FiloSottile/gvt >> ~/.bashrc

ADD . /usr/lib/go/src/github.com/Dataman-Cloud/omega-metrics
WORKDIR /usr/lib/go/src/github.com/Dataman-Cloud/omega-metrics

RUN go build .
RUN cp omega-metrics.yaml.sample /etc/omega/omega-metrics.yaml

EXPOSE 9005

CMD ./omega-metrics
