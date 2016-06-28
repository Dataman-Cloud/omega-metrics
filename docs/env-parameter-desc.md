##omega-metrics环境变量说明
注释括号里面代表原来相对应配置文件的字段

	METRICS_CACHETIMEOUT=60														#metrics超时时间设置
	METRICS_NUMCPU=1															#metrics监控机器CPU数目
	METRICS_HOST="0.0.0.0"														#metrics监控机器ip
	METRICS_PORT=9005															#metrics进程占用端口号
	METRICS_DEBUGGING=false														#是否为debug模式
	METRICS_OMEGA_APP_HOST="http://app"											#app地址
	METRICS_OMEGA_APP_PORT=6080													#app端口
	METRICS_HEALTHCHECK=60														#健康监测间隔
	METRICS_LOG_CONSOLE=true													#是否在终端打印日志	
	METRICS_LOG_FILE="/var/log/omega/omega-metrics.log"							#日志存储路径
	METRICS_LOG_LEVEL="debug"													#日志级别
	METRICS_LOG_FORMATTER="%Date(2006-01-02 15:04:05Z07:00) [%LEVEL] %Msg%n"	#日志时间格式
	METRICS_LOG_FILESIZE=5000000												#日志文件大小
	METRICS_LOG_FILENUM=10														#日志文件个数
	METRICS_CACHE_HOST="redis"													#redis地址
	METRICS_CACHE_PORT=6379														#redis端口
	METRICS_CACHE_PASSWORD="123"												#redis密码														
	METRICS_CACHE_POOLSIZE=100													#redis连接池大小
	METRICS_MQ_USER="guest"														#rabbitmq用户名
	METRICS_MQ_PASSWORD="guest"													#rabbitmq密码 
	METRICS_MQ_HOST="rmq"														#rabbitmq地址 
	METRICS_MQ_PORT=5672														#rabbitmq端口 
	METRICS_DB_USER="root"														#influxdb用户名
	METRICS_DB_PASSWORD="root"													#influxdb密码 
	METRICS_DB_HOST="influxdb"													#influxdb地址 
	METRICS_DB_PORT=5008														#influxdb端口 
	METRICS_DB_DATABASE="shurenyun"												#influxdb数据库名
	METRICS_DB_QUERY_DEFAULT_DURATION="15m"  # 's' seconds, 'm' minutes, 'h' hours, 'd' days, 'w' weeks	#influxdb默认时间格式
