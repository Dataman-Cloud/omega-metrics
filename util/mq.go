package util

import (
	"fmt"
	"sync"

	"github.com/Dataman-Cloud/omega-metrics/config"
	"github.com/Dataman-Cloud/omega-metrics/logger"
	log "github.com/cihub/seelog"
	"github.com/streadway/amqp"
)

const (
	Metrics_exchange         string = "cluster_info"
	Master_metrics_routing   string = "master_metrics"
	Master_state_routing     string = "master_state"
	Slave_metrics_routing    string = "slave_metrics"
	Slave_state_routing      string = "slave_state"
	Slave_monitor_routing    string = "slave_monitor"
	Marathon_event_routing   string = "marathon_event"
	Marathon_metrics_routing string = "marathon_metrics"
	Marathon_info_routing    string = "marathon_info"
	ExchangeCluster          string = "cluster"
	RoutingHealth            string = "health"
)

func failOnError(err error, msg string) error {
	if err != nil {
		log.Errorf("%s: %s", msg, err)
	}
	return err
}

var mq *amqp.Connection
var reconChan = make(chan *amqp.Error, 1)

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

func MetricsSubscribe(exchange string, routingkey string, handler func(messageBody *[]byte)) error {
	defer func() {
		if err := recover(); err != nil {
			stack := logger.Stack(5)
			log.Errorf("Panic recovery -> %s\n%s\n", err, stack)
			log.Flush()
			MetricsSubscribe(exchange, routingkey, handler)
		}
	}()

	mq := MQ()
	channel, err := mq.Channel()
	failOnError(err, "can't get channel")

	err = channel.ExchangeDeclare(exchange, "direct", true, false, false, false, nil)
	failOnError(err, "can't declare exchange")

	queue, err := declareQueue(channel, routingkey)
	failOnError(err, "can't declare queue")

	err = channel.QueueBind(queue.Name, routingkey, exchange, false, nil)
	failOnError(err, "can't bind queue")

	messages, err := channel.Consume(routingkey, "", true, false, false, false, nil)
	failOnError(err, "can't consume")

	defer channel.Close()
	for {
		select {
		case message, ok := <-messages:
			handler(&message.Body)
			if !ok {
				log.Errorf("channel of queue with exchange:%s and routingkey:%s quit!", exchange, routingkey)

				panic(-2)
			}
		}
	}
	return nil
}

func Publish(exchange, routing string, message string) error {
	mq := MQ()
	channel, err := mq.Channel()
	failOnError(err, "can't get channel")

	defer channel.Close()
	err = channel.ExchangeDeclare(exchange, "direct", true, false, false, false, nil)
	failOnError(err, "can't declare exchange")

	err = channel.Publish(
		exchange,
		routing,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	failOnError(err, "can't publish message")
	return nil
}

func declareQueue(channel *amqp.Channel, name string) (amqp.Queue, error) {
	args := amqp.Table{
		"x-message-ttl": int64(300000),
		"x-expires":     int64(1000 * 60 * 60 * 24 * 1),
	}
	return channel.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		args,
	)
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

	go onNotifyClose()
	mq.NotifyClose(reconChan)

	log.Debug("initialized MQ")
}

func onNotifyClose() {
	for {
		select {
		case err := <-reconChan:
			fmt.Println("mq connect closed with error", err)
			panic(-1)
		}
	}
}

func DestroyMQ() {
	log.Info("destroying MQ")
	if mq != nil {
		mq.Close()
	}
}
