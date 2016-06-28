package logger

import (
	"github.com/Dataman-Cloud/omega-metrics/config"
	log "github.com/Sirupsen/logrus"
)

var conf = config.Pairs()

var (
	levels map[string]log.Level
	level  log.Level
)

func initLogger(logLevel string) {
	levels = map[string]log.Level{
		"debug": log.DebugLevel,
		"info":  log.InfoLevel,
		"warn":  log.WarnLevel,
		"error": log.ErrorLevel,
		"panic": log.PanicLevel,
		"fatal": log.FatalLevel,
	}
	if logLevel == "" {
		level = log.DebugLevel
	}
	log.SetLevel(level)
}

func LoadLogConfig() {
	logLevel := conf.Log.Level
	initLogger(logLevel)
}
