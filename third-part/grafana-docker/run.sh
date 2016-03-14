#!/bin/bash -e

admin_user="admin"
admin_pass="admin"

start_grafana() {
    chown -R grafana:grafana /var/lib/grafana /var/log/grafana

    exec gosu grafana /usr/sbin/grafana-server  \
      --homepath=/usr/share/grafana             \
      --config=/etc/grafana/grafana.ini         \
      cfg:default.paths.data=/var/lib/grafana   \
      cfg:default.paths.logs=/var/log/grafana  
}

check_conn_influxdb() {
    time=0
    check=1
    while [ $time==3 ] || [ $check==0 ] 
    do 
       curl -sG  "http://$INFLUXDB_HOST:$INFLUXDB_PORT/query?" \
          --data-urlencode "u=$INFLUXDB_USER" \
          --data-urlencode "p=$INFLUXDB_PASS" \
          --data-urlencode "q=SHOW DATABASES"  >> /dev/null
       check=$?
       if [ $check == 0 ];
       then echo "Success connect to the influxdb server"
            return 0
       else 
            echo $check $time
            if [ $time -ge 3 ]
            then
                echo "Failed to connect to the influxdb server in 15s"
                exit 1
            else
                echo "Connection Failed, Will try 5s later.."
                time=$(($time+1))
                sleep 5
            fi
       fi
    done
} 

check_env_variable() {
    if [[ $INFLUXDB_HOST ]];
    then echo "INFLUXDB_HOST is $INFLUXDB_HOST"; 
    else echo "Need set the Environment Variable \"INFLUXDB_HOST\"" ; exit 1
    fi
    if [[ $INFLUXDB_PORT ]];
    then echo "INFLUXDB_PORT is $INFLUXDB_PORT"; 
    else echo "Need set the Environment Variable \"INFLUXDB_PORT\"" ; exit 1
    fi
    if [[ $INFLUXDB_USER ]];
    then echo "INFLUXDB_USER is $INFLUXDB_USER"; 
    else echo "Need set the Environment Variable \"INFLUXDB_USER\"" ; exit 1
    fi
    if [[ $INFLUXDB_PASS ]];
    then echo "INFLUXDB_PASS is $INFLUXDB_PASS"; 
    else echo "Need set the Environment Variable \"INFLUXDB_PASS\"" ; exit 1
    fi
    if [[ $INFLUXDB_NAME ]];
    then echo "INFLUXDB_NAME is $INFLUXDB_NAME"; 
    else echo "Need set the Environment Variable \"INFLUXDB_NAME\"" ; exit 1
    fi
    if [[ $GF_SECURITY_ADMIN_PASSWORD ]];
    then echo "The password of grafana user was passed!"; 
         admin_pass=$GF_SECURITY_ADMIN_PASSWORD
         echo "admin_pass=$admin_pass"
    fi
    
}

create_grafana_datasource() {
    url="http://$INFLUXDB_HOST:$INFLUXDB_PORT"
    curl -sX POST -i "http://$admin_user:$admin_pass@localhost:3000/api/datasources" \
       -H "Content-Type: application/json" \
       -H "Accept: application/json" \
       -d '{"name":"influxdb",
            "type":"influxdb",
            "access":"proxy",
            "url":"'"$url"'",
            "password":"'"$INFLUXDB_PASS"'",
            "user":"'"$INFLUXDB_USER"'",
            "database":"'"$INFLUXDB_NAME"'",
            "basicAuth":true,
            "basicAuthUser":"'"$INFLUXDB_USER"'",
            "basicAuthPassword":"'"$INFLUXDB_PASS"'",
            "withCredentials":false,
            "isDefault":true}' 
}

create_grafana_dashboard() {
    if [ ! -f "dataman.json" ] 
    then echo "Lost the file dataman.json"
         exit 1 
    fi
    url="http://$INFLUXDB_HOST:$INFLUXDB_PORT"
    echo $dashboard
    curl -sX POST -i "http://$admin_user:$admin_pass@localhost:3000/api/dashboards/db" \
       -H "Content-Type: application/json" \
       -H "Accept: application/json" \
       -d @dataman.json
}

 
check_env_variable
check_conn_influxdb
/etc/init.d/grafana-server start
create_grafana_datasource
create_grafana_dashboard
/etc/init.d/grafana-server stop
start_grafana

