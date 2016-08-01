#!/bin/bash
set -x

curl -v -X PUT $MARATHON_API_URL/v2/apps/shurenyun-$TASKENV-$SERVICE -H Content-Type:application/json -d \
'{
      "id": "shurenyun-'$TASKENV'-'$SERVICE'",
      "cpus": '$CPUS',
      "mem": '$MEM',
      "instances": '$INSTANCES',
      "constraints": [["hostname", "LIKE", "'$DEPLOYIP'"]],
      "container": {
                     "type": "DOCKER",
                     "docker": {
                                     "image": "'$SERVICE_IMAGE'",
                                     "network": "BRIDGE",
                                     "parameters": [
                                                 {"key": "dns","value": "'$DNS_IP'"}
                                                   ],
                                     "privileged": '$PRIVILEGED',
                                     "forcePullImage": '$FORCEPULLIMAGE',
                                     "portMappings": [
                                             { "containerPort": '$METRICS_PORT', "hostPort": 0, "protocol": "tcp"}
                                     ]
                                }
                   },
      "healthChecks": [{
               "path": "/api/v3/health/metrics",
               "protocol": "HTTP",
               "gracePeriodSeconds": 300,
               "intervalSeconds": 60,
               "portIndex": 0,
               "timeoutSeconds": 20,
               "maxConsecutiveFailures": 3
           }],
      "env": {
                    "BAMBOO_TCP_PORT": "'$BAMBOO_TCP_PORT'",
                    "BAMBOO_PRIVATE": "'$BAMBOO_PRIVATE'",
                    "BAMBOO_PROXY":"'$BAMBOO_PROXY'",
                    "BAMBOO_BRIDGE": "'$BAMBOO_BRIDGE'",
                    "BAMBOO_HTTP": "'$BAMBOO_HTTP'"
             },
      "uris": [
               "'$CONFIGSERVER'/config/demo/config/registry/docker.tar.gz"
       ]
}'

