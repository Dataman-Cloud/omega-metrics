package main

import (
	"fmt"
	"github.com/Dataman-Cloud/omega-metrics/cache"
	"github.com/Dataman-Cloud/omega-metrics/config"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/Sirupsen/logrus"
	redis "github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
)

func init() {
	conn := cache.Open()
	defer conn.Close()
}

func startC() {
	fmt.Println("start master metrics mq consumer")
	util.MetricsSubscribe(util.Metrics_exchange, util.Master_metrics_routing, handler)
	/*	fmt.Println("start master state mq consumer")
		util.MetricsSubscribe(util.Metrics_exchange, util.Master_state_routing, handler)*/
}

func startP() {
	json1 := `{"type": "masterState", "clusterId": "1", "nodeId": "1ad3derad1234451dggdq3e3e3", "timestamp": 1442900989, "message":{"allocator\/event_queue_dispatches":0,"master\/cpus_percent":2.77555756156289e-17,"master\/cpus_revocable_percent":0,"master\/cpus_revocable_total":0,"master\/cpus_revocable_used":0,"master\/cpus_total":2,"master\/cpus_used":5.55111512312578e-17,"master\/disk_percent":0,"master\/disk_revocable_percent":0,"master\/disk_revocable_total":0,"master\/disk_revocable_used":0,"master\/disk_total":14908,"master\/disk_used":0,"master\/dropped_messages":14,"master\/elected":1,"master\/event_queue_dispatches":21,"master\/event_queue_http_requests":0,"master\/event_queue_messages":0,"master\/frameworks_active":2,"master\/frameworks_connected":2,"master\/frameworks_disconnected":0,"master\/frameworks_inactive":0,"master\/invalid_framework_to_executor_messages":0,"master\/invalid_status_update_acknowledgements":0,"master\/invalid_status_updates":0,"master\/mem_percent":0,"master\/mem_revocable_percent":0,"master\/mem_revocable_total":0,"master\/mem_revocable_used":0,"master\/mem_total":2929,"master\/mem_used":0,"master\/messages_authenticate":0,"master\/messages_deactivate_framework":0,"master\/messages_decline_offers":1099789,"master\/messages_exited_executor":2,"master\/messages_framework_to_executor":0,"master\/messages_kill_task":11,"master\/messages_launch_tasks":76,"master\/messages_reconcile_tasks":16369,"master\/messages_register_framework":3,"master\/messages_register_slave":5,"master\/messages_reregister_framework":2,"master\/messages_reregister_slave":1,"master\/messages_resource_request":0,"master\/messages_revive_offers":2,"master\/messages_status_update":152,"master\/messages_status_update_acknowledgement":152,"master\/messages_unregister_framework":3,"master\/messages_unregister_slave":0,"master\/messages_update_slave":2,"master\/outstanding_offers":0,"master\/recovery_slave_removals":0,"master\/slave_registrations":1,"master\/slave_removals":1,"master\/slave_removals\/reason_registered":1,"master\/slave_removals\/reason_unhealthy":0,"master\/slave_removals\/reason_unregistered":0,"master\/slave_reregistrations":1,"master\/slave_shutdowns_canceled":0,"master\/slave_shutdowns_completed":0,"master\/slave_shutdowns_scheduled":0,"master\/slaves_active":1,"master\/slaves_connected":1,"master\/slaves_disconnected":0,"master\/slaves_inactive":0,"master\/tasks_error":0,"master\/tasks_failed":4,"master\/tasks_finished":61,"master\/tasks_killed":11,"master\/tasks_lost":0,"master\/tasks_running":0,"master\/tasks_staging":0,"master\/tasks_starting":0,"master\/uptime_secs":3267541.51136102,"master\/valid_framework_to_executor_messages":0,"master\/valid_status_update_acknowledgements":152,"master\/valid_status_updates":152,"registrar\/queued_operations":0,"registrar\/registry_size_bytes":260,"registrar\/state_fetch_ms":80.187904,"registrar\/state_store_ms":4.573184,"registrar\/state_store_ms\/count":2,"registrar\/state_store_ms\/max":4.573184,"registrar\/state_store_ms\/min":4.406784,"registrar\/state_store_ms\/p50":4.489984,"registrar\/state_store_ms\/p90":4.556544,"registrar\/state_store_ms\/p95":4.564864,"registrar\/state_store_ms\/p99":4.57152,"registrar\/state_store_ms\/p999":4.5730176,"registrar\/state_store_ms\/p9999":4.57316736,"system\/cpus_total":2,"system\/load_15min":0.07,"system\/load_1min":0.03,"system\/load_5min":0.04,"system\/mem_free_bytes":626851840,"system\/mem_total_bytes":4145487872}}`

	fmt.Println("start mq publisher")
	util.MetricsPublish(util.Metrics_exchange, []byte(json1))
}

func handler(routingKey string, messageBody []byte) {
	switch routingKey {
	case util.Master_metrics_routing:
		id, json := util.MasterMetricsJson(string(messageBody))
		if id != "" && json != "" {
			label := id + "_" + routingKey
			fmt.Println(label)
			writeToRedis(label, json)
		}
	case util.Slave_metrics_routing:
		id, json := util.SlaveMetricsJson(string(messageBody))
		if id != "" && json != "" {
			label := id + "_" + routingKey
			writeToRedis(label, json)
		}
	}
}

func writeToRedis(id string, json string) {
	conn := cache.Open()
	defer conn.Close()
	fmt.Println("write to redis")
	conn.Send("LPUSH", id, json)
	conn.Send("EXPIRE", id, config.DefaultTimeout)
	_, err := conn.Do("LTRIM", id, 0, 19)
	if err != nil {
		log.Errorf("LPUSH key:%s value:%s is wrong", id, json)
		log.Errorln("[writeToRedis] error is ", err)
	}
	strs, _ := redis.Strings(conn.Do("LRANGE", "1_master_metrics", 0, -1))
	fmt.Println("Data: ", strs)
}

func masterMetrics(ctx *gin.Context) {
	conn := cache.Open()
	defer conn.Close()
	cluster_id := ctx.Param("cluster_id") + "_" + util.Master_metrics_routing
	log.Debug("cluster_id ", cluster_id)
	strs, err := redis.Strings(conn.Do("LRANGE", cluster_id, 0, -1))
	if err != nil {
		log.Error("[Master Metrics] got error ", err)
		jsoninterface := util.ReturnMessage("1", nil, "[MasterMetrics] got error")
		ctx.JSON(http.StatusOK, jsoninterface)
	}
	jsoninterface := util.ReturnMessage("0", strs, "")
	log.Info("Got master metrics", jsoninterface)
	fmt.Println(reflect.TypeOf(jsoninterface))
	ctx.JSON(http.StatusOK, jsoninterface)
}
