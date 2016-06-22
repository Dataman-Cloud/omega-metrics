package config

import (
	"bufio"
	"errors"
	"flag"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/Dataman-Cloud/omega-metrics/model"
	log "github.com/Sirupsen/logrus"
)

const (
	ContainerMonitorTimeout    int = 30 // seconds
	SessionInfoTimeout         int = 60 // seconds
	DefaultTimeout             int = 24 * 3600
	DefaultHost                    = "localhost"
	DefaultPort                    = 9005
	DefaultDebugging               = true
	DefaultLogLevel                = "debug"
	DefaultHealthCheckInterval     = 60
	ContainerMonitorSerie          = "Slave_state"
	AppRequestInfoSerie            = "app_req_rate"
)

type EnvEntry struct {
	METRICS_CACHETIMEOUT              int    `required:"true"`
	METRICS_NUMCPU                    int    `required:"true"`
	METRICS_HOST                      string `required:"true"`
	METRICS_PORT                      int    `required:"true"`
	METRICS_DEBUGGING                 bool   `required:"true"`
	METRICS_OMEGA_APP_HOST            string `required:"true"`
	METRICS_OMEGA_APP_PORT            int    `required:"true"`
	METRICS_HEALTHCHECK               int    `required:"true"`
	METRICS_LOG_CONSOLE               bool   `required:"true"`
	METRICS_LOG_FILE                  string `required:"true"`
	METRICS_LOG_LEVEL                 string `required:"true"`
	METRICS_LOG_FORMATTER             string `required:"true"`
	METRICS_LOG_FILESIZE              int    `required:"true"`
	METRICS_LOG_FILENUM               int    `required:"true"`
	METRICS_CACHE_HOST                string `required:"true"`
	METRICS_CACHE_PORT                uint64 `required:"true"`
	METRICS_CACHE_PASSWORD            string `required:"true"`
	METRICS_CACHE_DB                  int64  `required:"true"`
	METRICS_CACHE_LLEN                int    `required:"true"`
	METRICS_CACHE_POOLSIZE            int    `required:"true"`
	METRICS_MQ_USER                   string `required:"true"`
	METRICS_MQ_PASSWORD               string `required:"true"`
	METRICS_MQ_HOST                   string `required:"true"`
	METRICS_MQ_PORT                   int64  `required:"true"`
	METRICS_DB_USER                   string `required:"true"`
	METRICS_DB_PASSWORD               string `required:"true"`
	METRICS_DB_HOST                   string `required:"true"`
	METRICS_DB_PORT                   int64  `required:"true"`
	METRICS_DB_DATABASE               string `required:"true"`
	METRICS_DB_QUERY_DEFAULT_DURATION string `required:"true"`
}

var config model.Config

func Pairs() model.Config {
	return config
}

func initDefault(config *model.Config) {
	config.NumCPU = runtime.NumCPU()
	config.Host = DefaultHost
	config.Port = DefaultPort
	config.Debugging = true
}

func InitConfig() {
	log.Info("initing config ...")
	envFile := flag.String("config", "env", "")
	flag.Parse()
	loadEnvFile(*envFile)
	initDefault(&config)
	envEntry := NewEnvEntry()
	config.CacheTimeout = envEntry.METRICS_CACHETIMEOUT
	config.NumCPU = envEntry.METRICS_NUMCPU
	config.Host = envEntry.METRICS_HOST
	config.Port = envEntry.METRICS_PORT
	config.Debugging = envEntry.METRICS_DEBUGGING
	config.Omega_app_host = envEntry.METRICS_OMEGA_APP_HOST
	config.Omega_app_port = envEntry.METRICS_OMEGA_APP_PORT
	config.HealthCheckInterval = envEntry.METRICS_HEALTHCHECK
	config.Log.Console = envEntry.METRICS_LOG_CONSOLE
	config.Log.File = envEntry.METRICS_LOG_FILE
	config.Log.Level = envEntry.METRICS_LOG_LEVEL
	config.Log.Formatter = envEntry.METRICS_LOG_FORMATTER
	config.Log.FileSize = envEntry.METRICS_LOG_FILESIZE
	config.Log.FileNum = envEntry.METRICS_LOG_FILENUM
	config.Cache.Host = envEntry.METRICS_CACHE_HOST
	config.Cache.Port = envEntry.METRICS_CACHE_PORT
	config.Cache.Password = envEntry.METRICS_CACHE_PASSWORD
	config.Cache.DB = envEntry.METRICS_CACHE_DB
	config.Cache.PoolSize = envEntry.METRICS_CACHE_POOLSIZE
	config.Cache.Llen = envEntry.METRICS_CACHE_LLEN
	config.Mq.User = envEntry.METRICS_MQ_USER
	config.Mq.Password = envEntry.METRICS_MQ_PASSWORD
	config.Mq.Host = envEntry.METRICS_MQ_HOST
	config.Mq.Port = envEntry.METRICS_MQ_PORT
	config.Db.User = envEntry.METRICS_DB_USER
	config.Db.Password = envEntry.METRICS_DB_PASSWORD
	config.Db.Host = envEntry.METRICS_DB_HOST
	config.Db.Port = envEntry.METRICS_DB_PORT
	config.Db.Database = envEntry.METRICS_DB_DATABASE
	config.Db.Query_Default_Duration = envEntry.METRICS_DB_QUERY_DEFAULT_DURATION
	log.Info("initialized config : %q\n", config)
}
func NewEnvEntry() *EnvEntry {
	envEntry := &EnvEntry{}

	val := reflect.ValueOf(envEntry).Elem()

	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		required := typeField.Tag.Get("required")

		env := os.Getenv(typeField.Name)

		if env == "" && required == "true" {
			exitMissingEnv(typeField.Name)
		}

		var envEntryValue interface{}
		var err error
		valueFiled := val.Field(i).Interface()
		value := val.Field(i)
		switch valueFiled.(type) {
		case int:
			envEntryValue, err = strconv.Atoi(env)

		case int64:
			envEntryValue, err = strconv.ParseInt(env, 10, 64)

		case int16:
			envEntryValue, err = strconv.ParseInt(env, 10, 16)
			_, ok := envEntryValue.(int64)
			if !ok {
				exitCheckEnv(typeField.Name, err)
			}
			envEntryValue = int16(envEntryValue.(int64))
		case uint16:
			envEntryValue, err = strconv.ParseUint(env, 10, 16)

			_, ok := envEntryValue.(uint64)
			if !ok {
				exitCheckEnv(typeField.Name, err)
			}
			envEntryValue = uint16(envEntryValue.(uint64))
		case uint64:
			envEntryValue, err = strconv.ParseUint(env, 10, 64)
		case bool:
			envEntryValue, err = strconv.ParseBool(env)
		default:
			envEntryValue = env
		}

		if err != nil {
			exitCheckEnv(typeField.Name, err)
		}
		value.Set(reflect.ValueOf(envEntryValue))
	}
	return envEntry
}

// helper function to parse a "key=value" environment variable string.
func parseln(line string) (key string, val string, err error) {
	line = removeComments(line)
	if len(line) == 0 {
		return
	}
	splits := strings.SplitN(line, "=", 2)

	if len(splits) < 2 {
		err = errors.New("missing delimiter '='")
		return
	}

	key = strings.Trim(splits[0], " ")
	val = strings.Trim(splits[1], ` "'`)
	return

}
func loadEnvFile(envfile string) {
	// load the environment file
	f, err := os.Open(envfile)
	if err == nil {
		defer f.Close()

		r := bufio.NewReader(f)
		for {
			line, _, err := r.ReadLine()
			if err != nil {
				break
			}
			if len(line) == 0 {
				continue
			}
			key, val, err := parseln(string(line))
			if err != nil {
				continue
			}

			if len(os.Getenv(strings.ToUpper(key))) == 0 {
				err1 := os.Setenv(strings.ToUpper(key), val)
				if err1 != nil {
					log.Errorln(err1.Error())
				}
			}
		}
	}
}

// helper function to trim comments and whitespace from a string.
func removeComments(s string) (_ string) {
	if len(s) == 0 || string(s[0]) == "#" {
		return
	} else {
		index := strings.Index(s, " #")
		if index > -1 {
			s = strings.TrimSpace(s[0:index])
		}
	}
	return s
}

func exitMissingEnv(env string) {
	log.Fatalf("program exit missing config for env %s", env)
	os.Exit(1)
}

func exitCheckEnv(env string, err error) {
	log.Fatalf("Check env %s, %s", env, err.Error())

}
