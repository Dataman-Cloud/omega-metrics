package main

import (
	"github.com/Dataman-Cloud/omega-metrics/metrics/master"
	"github.com/Dataman-Cloud/omega-metrics/metrics/slave"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/cihub/seelog"
)

func startC() {
	log.Debug("start master metrics mq consumer")
	go util.MetricsSubscribe(util.Metrics_exchange, util.Master_state_routing, master.MasterStateHandler)
	go util.MetricsSubscribe(util.Metrics_exchange, util.Master_metrics_routing, master.MasterMetricHandler)
	go util.MetricsSubscribe(util.Metrics_exchange, util.Slave_state_routing, slave.SlaveStateHandler)
	go util.MetricsSubscribe(util.Metrics_exchange, util.Slave_metrics_routing, func(messageBody *[]byte) {})
	go util.MetricsSubscribe(util.Metrics_exchange, util.Marathon_info_routing, func(messageBody *[]byte) {})
	go util.MetricsSubscribe(util.Metrics_exchange, util.Marathon_metrics_routing, func(messageBody *[]byte) {})
	go util.MetricsSubscribe(util.Metrics_exchange, util.Slave_monitor_routing, func(messageBody *[]byte) {})
}
