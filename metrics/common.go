package metrics

import (
	"encoding/json"

	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/Sirupsen/logrus"
)

func ParserMqMessage(messgae *[]byte) *util.RabbitMqMessage {
	var mqMessage *util.RabbitMqMessage = &util.RabbitMqMessage{}
	err := json.Unmarshal(*messgae, mqMessage)
	if err != nil {
		log.Error("Parser mq message has error: ", err)
		return nil
	}
	return mqMessage
}
