package util

import (
	"fmt"
	"sync"
	"time"

	"github.com/Dataman-Cloud/omega-metrics/config"
	log "github.com/cihub/seelog"
	"github.com/streadway/amqp"
)

const (
	Metrics_exchange             string = "cluster_info"
	Master_metrics_routing       string = "master_metrics"
	Master_state_routing         string = "master_state"
	Slave_metrics_routing        string = "slave_metrics"
	Slave_state_routing          string = "slave_state"
	Marathon_event_routing       string = "marathon_event"
	Marathon_apps_routing        string = "marathon_apps"
	Marathon_metrics_routing     string = "marathon_metrics"
	Marathon_deployments_routing string = "marathon_deployments"
)

func failOnError(err error, msg string) error {
	if err != nil {
		log.Errorf("%s: %s", msg, err)
	}
	return err
}

var mq *amqp.Connection
var killChan <-chan bool
var defaultHandlers func(string, []byte)
var reconChan chan *amqp.Error = make(chan *amqp.Error)

func MQ() *amqp.Connection {
	if mq != nil {
		return mq
	}

	InitMQ(killChan, defaultHandlers)

	return mq
}

func MetricsSubscribe(exchange string, routingkey string, handler func(string, []byte)) error {
	mq := MQ()
	channel, err := mq.Channel()
	failOnError(err, "can't get channel")

	err = channel.ExchangeDeclare(exchange, "direct", true, false, false, false, nil)
	failOnError(err, "can't declare exchange")

	err = channel.QueueBind(routingkey, routingkey, exchange, false, nil)
	failOnError(err, "can't bind queue")

	messages, err := channel.Consume(routingkey, "", true, false, false, false, nil)
	failOnError(err, "can't consume")

	go func() {
		defer channel.Close()
		for message := range messages {
			handler(message.RoutingKey, message.Body)
		}
	}()

	return nil
}

func InitMQ(exitChan <-chan bool, handler func(string, []byte)) {
	mutex := sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()
	killChan = exitChan
	defaultHandlers = handler

	conf := config.Pairs()
	opts := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		conf.Mq.User, conf.Mq.Password, conf.Mq.Host, conf.Mq.Port)
	var err error
	mq, err = amqp.Dial(opts)
	if err != nil {
		log.Error("got err", err)
		log.Error("can't dial mq server: ", opts)
		panic(-1)
	}

	go onNotifyClose()
	mq.NotifyClose(reconChan)

	startC(handler)

	log.Debug("initialized MQ")
}

func onNotifyClose() {
	for {
		select {
		case err := <-reconChan:
			log.Error("mq connect closed with error", err)
			time.Sleep(time.Second * 3)
			InitMQ(killChan, defaultHandlers)
		case <-killChan:
			log.Info("close mq reconnector")
		}
	}
}

func startC(handler func(string, []byte)) {
	log.Debug("start master metrics mq consumer")
	MetricsSubscribe(Metrics_exchange, Master_metrics_routing, handler)
	MetricsSubscribe(Metrics_exchange, Marathon_event_routing, handler)
	MetricsSubscribe(Metrics_exchange, Slave_state_routing, handler)
	MetricsSubscribe(Metrics_exchange, Master_state_routing, handler)
	MetricsSubscribe(Metrics_exchange, Slave_metrics_routing, func(routingKey string, messageBody []byte) {})
	MetricsSubscribe(Metrics_exchange, Marathon_apps_routing, func(routingKey string, messageBody []byte) {})
	MetricsSubscribe(Metrics_exchange, Marathon_metrics_routing, func(routingKey string, messageBody []byte) {})
	MetricsSubscribe(Metrics_exchange, Marathon_deployments_routing, func(routingKey string, messageBody []byte) {})

}

func DestroyMQ() {
	log.Info("destroying MQ")
	if mq != nil {
		mq.Close()
	}
}
