package logger

import (
	"fmt"
	"strings"

	"github.com/Dataman-Cloud/omega-metrics/config"
	seelog "github.com/cihub/seelog"
)

var conf = config.Pairs()

var (
	levels       map[string]string
	logfile      string
	level        string
	logToScreen  bool
	formatString string
	fileNum      int
	fileSize     int
)

func initLogger() {
	levels = map[string]string{
		"debug": "debug",
		"info":  "info",
		"warn":  "warn",
		"error": "error",
		"crit":  "critical",
		"none":  "off",
	}

	if level == "" {
		level = config.DefaultLogLevel
	}

	if logfile == "" {
		logfile = "/var/log/omeaga/omega-metrics.log"
	}

	if fileNum <= 0 {
		fileNum = 10
	}

	if fileSize <= 0 {
		fileSize = 5000000
	}

	if formatString == "" {
		formatString = "%Date(2006-01-02 15:04:05Z07:00) [%LEVEL] %Msg%n"
	}

	fmt.Println(fileSize)
	fmt.Println(fileNum)
	SetLevel(level)
}

func SetLevel(logLevel string) {
	parsedLevel, ok := levels[strings.ToLower(logLevel)]
	if ok {
		level = parsedLevel
		reloadLogConfig()
	}
}

func reloadLogConfig() {
	fmt.Println(loggerConfig())
	logger, err := seelog.LoggerFromConfigAsString(loggerConfig())

	if err == nil {
		seelog.ReplaceLogger(logger)
	} else {
		seelog.Error(err)
	}
}

func LoadLogConfig() {
	logfile = conf.Log.File
	level = conf.Log.Level
	logToScreen = conf.Log.Console
	fileNum = conf.Log.FileNum
	fileSize = conf.Log.FileSize
	formatString = conf.Log.Formatter
	initLogger()
}
