#omega-metrics 安装环境

    文档信息
    创建人 韩路
    邮件地址 lhan@dataman-inc.com
    时间 2015-09-28

## 1. 安装GO的编译环境（Dockerfile）
    mkdir -p /data/tools/go1.5.linux-amd64
    vi /data/tools/go1.5.linux-amd64/Dockerfile
    
    ROM ubuntu:14.04
    MAINTAINER lhan lhan@dataman-inc.com

    #install go
    RUN mkdir -p /data/tools/go && cd /data/tools/go && \
        apt-get update && \
        apt-get install gcc automake autoconf libtool make -y && \
        apt-get install git-core wget -y && \
        rm -rf /var/lib/apt/lists/* && \
        wget https://storage.googleapis.com/golang/go1.5.linux-amd64.tar.gz && \
        tar xzf go* && \
        mv go /usr/local/.
        
## 2.1 编译 Omega-metrics Dockerfile
    mkdir -p /data/tools/omega-metrics
    vi /data/tools/omega-metrics/Dockerfile

    FROM go1.5.linux-amd64
    MAINTAINER lhan lhan@dataman-inc.com

    #install omega-metrics
    CMD ["/data/omega-metrics/build.sh"]
## 2.2 Omega-metrics 构建脚本
    mkdir -p /data/omega-metrics
    vi /data/omega-metrics/build.sh

    #!/bin/bash
    export PATH=$PATH:/usr/local/go/bin && \
    export GOROOT=/usr/local/go && \
    export PATH=$PATH:$GOROOT/bin && \
    export GOPATH=$HOME/go && \
    go get -u github.com/FiloSottile/gvt && \
    export GO15VENDOREXPERIMENT=1 && \
    # resolve gvt conflict
    echo 'alias ggvt='$GOPATH/bin/gvt >> ~/.bashrc && \
    mkdir -p $GOPATH/src/github.com/Dataman-Cloud/ && \
    cd $GOPATH/src/github.com/Dataman-Cloud/ && \
    git clone https://leonluhan:DataMan1234@github.com/Dataman-Cloud/omega-metrics && \
    cd $GOPATH/src/github.com/Dataman-Cloud/omega-metrics && \
    go build && mv omega-metrics /data/omega-metrics/
    
    chmod 777 /data/omega-metrics/build.sh
## 2.3 启动 Docker 编译出 Omega-metrics
    docker run -v /data/omega-metrics:/data/omega-metrics \
               -v /data/omega-metrics/omega-metrics.yaml:/omega-metrics.yaml \
                omega-metrics ./data/omega-metrics/omega-metrics
                
## 3. 测试Omega-metrics服务是否启动
    curl -X GET http://$host:9005/
返回值为pass。其中`$host`是配置文件`omega-metrics.yaml`中的`host`
