package model

type LogConfig struct {
	Console   bool
	File      string
	FileNum   int
	FileSize  int
	Level     string
	Formatter string
}

type CacheConfig struct {
	Host     string
	Port     uint64
	Password string
	DB       int64
	PoolSize int
	Llen     int
}

type MqConfig struct {
	User     string
	Password string
	Host     string
	Port     int64
}

type DbConfig struct {
	User                   string
	Password               string
	Host                   string
	Port                   int64
	Database               string
	Query_Default_Duration string
}

type Config struct {
	CacheTimeout        int
	NumCPU              int
	Host                string
	Port                int
	Debugging           bool
	Omega_app_host      string
	Omega_app_port      int
	Log                 LogConfig
	Cache               CacheConfig
	Mq                  MqConfig
	Db                  DbConfig
	HealthCheckInterval int
}
