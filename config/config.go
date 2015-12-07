package config

import (
	"log"
	"runtime"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

const (
	ContainerMonitorTimeout int = 30
	DefaultTimeout          int = 24 * 3600
	DefaultHost                 = "localhost"
	DefaultPort                 = 9005
	DefaultDebugging            = true
	DefaultLogLevel             = "debug"
)

type Config struct {
	CacheTimeout   int
	NumCPU         int
	Host           string
	Port           uint
	Debugging      bool
	Omega_app_host string
	Omega_app_port int
	Scale          *ScaleConfig
	Log            *LogConfig
	Cache          *CacheConfig
	Mq             *MqConfig
	Db             *DbConfig
}

type ScaleConfig struct {
	AutoScaling    bool
	MaxMemPercent  float64
	MaxCpuPercent  float64
	MinMemPercent  float64
	MinCpuPercent  float64
	WaitCheckTimes int
	MaxInstances   int
	MinInstances   int
}

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
	Port     uint
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
	User         string
	Password     string
	Host         string
	Port         uint
	Name         string
	MaxIdleConns int
	MaxOpenConns int
}

var pairs Config

func Pairs() Config {
	return pairs
}

func Init() {
	jww.SetLogThreshold(jww.LevelTrace)
	jww.SetStdoutThreshold(jww.LevelInfo)

	log.Println("initing config ...")

	viper.SetConfigName("omega-metrics")
	viper.AddConfigPath("./conf/")
	viper.AddConfigPath("$HOME/.omega/")
	viper.AddConfigPath("/etc/omega/")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Panicln("can't read config file:", err)
	}

	initDefault()

	if err := viper.Unmarshal(&pairs); err != nil {
		log.Panicln("can't covert to config pairs: ", err)
	}

	if !pairs.Debugging {
		jww.SetLogThreshold(jww.LevelError)
		jww.SetStdoutThreshold(jww.LevelError)
	}
	log.Printf("initialized config pairs: %q\n", pairs)

}

func init() {
	jww.SetLogThreshold(jww.LevelTrace)
	jww.SetStdoutThreshold(jww.LevelInfo)

	log.Println("initing config ...")

	viper.SetConfigName("omega-metrics")
	viper.AddConfigPath("./conf/")
	viper.AddConfigPath("$HOME/.omega/")
	viper.AddConfigPath("/etc/omega/")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Panicln("can't read config file:", err)
	}

	initDefault()

	if err := viper.Unmarshal(&pairs); err != nil {
		log.Panicln("can't covert to config pairs: ", err)
	}

	if !pairs.Debugging {
		jww.SetLogThreshold(jww.LevelError)
		jww.SetStdoutThreshold(jww.LevelError)
	}
	log.Printf("initialized config pairs: %q\n", pairs)
}

func initDefault() {
	viper.SetDefault("numCPU", runtime.NumCPU())
	viper.SetDefault("host", DefaultHost)
	viper.SetDefault("port", DefaultPort)
	viper.SetDefault("debugging", true)
}

func Get(name string) interface{} {
	return viper.Get(name)
}

func GetString(name string) string {
	return viper.GetString(name)
}

func GetInt(name string) int {
	return viper.GetInt(name)
}
