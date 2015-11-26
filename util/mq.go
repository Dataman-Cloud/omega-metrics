package util

import (
	"fmt"
	"sync"

	"github.com/Dataman-Cloud/omega-metrics/config"
	log "github.com/cihub/seelog"
	"github.com/streadway/amqp"
)

const (
	Metrics_exchange         string = "cluster_info"
	Master_metrics_routing   string = "master_metrics"
	Master_state_routing     string = "master_state"
	Slave_metrics_routing    string = "slave_metrics"
	Slave_state_routing      string = "slave_state"
	Marathon_event_routing   string = "marathon_event"
	Marathon_metrics_routing string = "marathon_metrics"
	Marathon_info_routing    string = "marathon_info"
)

func failOnError(err error, msg string) error {
	if err != nil {
		log.Errorf("%s: %s", msg, err)
	}
	return err
}

var mq *amqp.Connection

func MQ() *amqp.Connection {
	if mq != nil {
		return mq
	}

	mutex := sync.Mutex{}
	mutex.Lock()
	InitMQ()
	defer mutex.Unlock()

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

func InitMQ() {
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
	log.Debug("initialized MQ")
}

func DestroyMQ() {
	log.Info("destroying MQ")
	if mq != nil {
		mq.Close()
	}
}
