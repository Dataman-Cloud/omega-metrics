package util

const (
	MonitorMasterMetrics = "MasterMetricsMar"
	MonitorMarathonEvent = "MarathonEventMar"
)

type MasterMetricsMar struct {
	CpuPercent float64 `json:"cpuPercent"`
	MemTotal   int     `json:"memTotal"`
	MemUsed    int     `json:"memUsed"`
	DiskTotal  int     `json:"diskTotal"`
	DiskUsed   int     `json:"diskUsed"`
	Timestamp  int64   `json:"timestamp"`
	Leader     int     `json:"leader"`
}

type MarathonEvent struct {
	EventType   string      `json:"eventType"`
	Timestamp   string      `json:"timestamp"`
	Id          string      `json:"id,omitempty"`
	Plan        plan        `json:"plan,omitempty"`
	CurrentStep currentStep `json:"currentStep,omitempty"`
}

type MonitorResponse struct {
	Code string      `json:"code"`
	Data interface{} `json:"data"`
	Err  string      `json:"error"`
}

type MarathonEventMar struct {
	EventType   string `json:"eventType"`
	Timestamp   string `json:"timestamp"`
	IdOrApp     string `json:"idOrApp"`
	CurrentType string `json:"currentType"`
	TaskId      string `json:"taskId"`
}
