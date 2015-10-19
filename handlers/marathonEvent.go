package handlers

import (
	"encoding/json"

	log "github.com/cihub/seelog"
)

type MarathonEventResponse struct {
	EventType   string `json:"eventType"`
	Timestamp   string `json:"timestamp"`
	IdOrApp     string `json:"idOrApp"`
	CurrentType string `json:"currentType"`
	TaskId      string `json:"taskId"`
}

func (req *MarathonEventResponse) String() string {
	reqString, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Marshal MarathonEventResponse has error: ", err)
		return ""
	}

	return string(reqString)
}
