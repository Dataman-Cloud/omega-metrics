package util

import (
	"io"
	"os"
	"strings"

	"github.com/Dataman-Cloud/omega-metrics/config"
	log "github.com/Sirupsen/logrus"
)

const (
	LevelsEnv = "LOGXI"
	FormatEnv = "LOGXI_FORMAT"
	ColorsEnv = "LOGXI_COLORS"
)

var logfile *os.File
var conf = config.Pairs()

func InitLog() {
	if conf.Log.Formatter == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{})
	}

	level, err := log.ParseLevel(strings.ToLower(conf.Log.Level))
	if err == nil {
		log.SetLevel(level)
	} else {
		log.SetLevel(log.DebugLevel)
	}

	logPath := conf.Log.File
	writers := []io.Writer{}

	if conf.Log.Console || len(logPath) < 1 {
		writers = append(writers, os.Stdout)
	}

	logInfo, err := os.Stat(logPath)
	if err == nil && logInfo.Size() > conf.Log.MaxSize {
		println("truncating logfile")
		os.Truncate(conf.Log.File, 0)
	}

	logfile, err = os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		writers = append(writers, logfile)
	}
	multiWriter := io.MultiWriter(writers...)
	log.SetOutput(multiWriter)

	log.Debugf("initialized logger: %s", conf.Log)

}

func DestroyLog() {
	log.Info("destroying log")
	if logfile != nil {
		log.Debug("logfile will be closed")
		logfile.Close()
	}
}
