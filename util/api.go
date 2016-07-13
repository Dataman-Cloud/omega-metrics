package util

import (
	"time"
)

const (
	MonitorMasterMetrics = "MasterMetricsMar"
	MonitorAppMetrics    = "SlaveStateMar"
)

type MonitorResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Err  string      `json:"error"`
}

type RabbitMqMessage struct {
	ClusterId int                `json:"clusterId"`
	NodeId    string             `json:"nodeId"`
	Type      string             `json:"type"`
	Message   string             `json:"message"`
	Attached  string             `json:"attached"`
	Tags      map[string]*string `json:"tags"`
	Timestamp int64              `json:"timestamp"`
}

// master metrics
type MasterMetrics struct {
	CpuPercent float64 `json:"master/cpus_percent"`
	CpuShare   float64 `json:"master/cpus_used"`
	CpuTotal   float64 `json:"master/cpus_total"`
	DiskUsed   float64 `json:"master/disk_used"`
	DiskTotal  float64 `json:"master/disk_total"`
	MemUsed    float64 `json:"master/mem_used"`
	MemTotal   float64 `json:"master/mem_total"`
	Leader     float64 `json:"master/elected"`
}

type MasterMetricsMar struct {
	CpuPercent float64 `json:"cpuPercent"`
	CpuShare   float64 `json:"cpuShare"`
	CpuTotal   float64 `json:"cpuTotal"`
	MemTotal   float64 `json:"memTotal"`
	MemUsed    float64 `json:"memUsed"`
	DiskTotal  float64 `json:"diskTotal"`
	DiskUsed   float64 `json:"diskUsed"`
	Timestamp  int64   `json:"timestamp"`
	Leader     float64 `json:"leader"`
	ClusterId  string  `json:"clusterId"`
}

// cluster metrics
type ClusterMetrics struct {
	MasMetrics interface{} `json:"masMetrics"`
	AppMetrics []AppMetric `json:"appMetrics"`
}

type AppMetric struct {
	AppName     string  `json:"appName"`
	AppCpuShare float64 `json:"appCpuShare"`
	AppCpuUsed  float64 `json:"appCpuUsed"`
	AppMemShare uint64  `json:"appMemShare"`
	AppMemUsed  float64 `json:"appMemUsed"`
	Instances   int64   `json:"instances"`
	Status      uint8   `json:"status"`
}

type StatusAndTask struct {
	Cid       int64  `json:"cid"`
	Name      string `json:"name"`
	Alias     string `json:"alias"`
	Status    uint8  `json:"status"`
	Tasks     int64  `json:"tasks"`
	Instances int64  `json:"instances"`
}

type AppListResponse struct {
	Code int                      `json:"code"`
	Data map[string]StatusAndTask `json:"data"`
	Err  string                   `json:"error"`
}

// master state
type MasterStateMar struct {
	Timestamp   int64             `json:"timestamp"`
	ClusterId   string            `json:"clusterId"`
	Leader      int               `json:"leader"`
	AppAndTasks []AppAndTasks     `json:"appAndTasks"`
	Slaves      []MasterSlaveInfo `json:"slaves"`
}

type AppAndTasks struct {
	AppName string `json:"appName"`
	TaskId  string `json:"taskId"`
}

type MasterState struct {
	HostName   string            `json:"hostname"`
	Frameworks []Framework       `json:"frameworks"`
	Leader     string            `json:"leader"`
	Slaves     []MasterSlaveInfo `json:"slaves"`
}

// slave info in master info
type MasterSlaveInfo struct {
	Id       string `json:"id"`
	Hostname string `json:"hostname"`
	Active   bool   `json:"active"`
}

// slave state
type SlaveStateMar struct {
	Timestamp               time.Time `json:"timestamp"`
	App                     AppInfo   `json:"app"`
	ContainerId             string    `json:"containerId"`
	CpuUsedCores            float64   `json:"cpuUsedCores"`
	CpuShareCores           float64   `json:"cpuShareCores"`
	MemoryTotal             float64   `json:"memoryTotal"`
	MemoryUsed              float64   `json:"memoryUsed"`
	NetworkReceviedByteRate float64   `json:"nw_rx_bytes"`
	NetworkSentByteRate     float64   `json:"nw_tx_bytes"`
	DiskIOReadBytesRate     float64   `json: "disk_read_bytes"`
	DiskIOWriteBytesRate    float64   `json: "disk_write_bytes"`
}

type SlaveState struct {
	Hostname   string      `json:"hostname"`
	Id         string      `json:"id"`
	Frameworks []Framework `json:"frameworks"`
	Flags      flag        `json:"flags"`
}

type flag struct {
	Ip string `json:"ip"`
}

type Framework struct {
	Id        string     `json:"id"`
	Name      string     `json:"name"`
	Hostname  string     `json:"hostname"`
	Executors []executor `json:"executors"`
	Tasks     []tasks    `json:"tasks,omitempty"`
}

type executor struct {
	Container string  `json:"container"`
	Id        string  `json:"id"`
	Tasks     []tasks `json:"tasks"`
}

type tasks struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	SlaveId   string    `json:"slave_id"`
	Resources resources `json:"resources"`
}

type resources struct {
	Cpus  float64 `json:"cpus"`
	Disk  float64 `json:"disk"`
	Mem   float64 `json:"mem"`
	Ports string  `json:"ports,omitempty"`
}

type AppInfo struct {
	AppName       string    `json:"appName"`
	AppId         string    `json:"appId"`
	ClusterId     string    `json:"clusterId"`
	TaskId        string    `json:"taskId"`
	SlaveId       string    `json:"slaveId"`
	Resources     resources `json:"resources"`
	ContainerName string    `json:"containerName"`
}

// cadivsor
type ContainerInfo struct {
	ContainerReference

	// The direct subcontainers of the current container.
	Subcontainers []ContainerReference `json:"subcontainers,omitempty"`

	// The isolation used in the container.
	Spec ContainerSpec `json:"spec,omitempty"`

	// Historical statistics gathered from the container.
	Stats []*ContainerStats `json:"stats,omitempty"`
}

// Container reference contains enough information to uniquely identify a container
type ContainerReference struct {
	// The absolute name of the container. This is unique on the machine.
	Name string `json:"name"`
	// Other names by which the container is known within a certain namespace.
	// This is unique within that namespace.
	Aliases []string `json:"aliases,omitempty"`

	// Namespace under which the aliases of a container are unique.
	// An example of a namespace is "docker" for Docker containers.
	Namespace string `json:"namespace,omitempty"`
}

type ContainerSpec struct {
	// Time at which the container was created.
	CreationTime time.Time `json:"creation_time,omitempty"`

	// Metadata labels associated with this container.
	Labels map[string]string `json:"labels,omitempty"`

	HasCpu bool    `json:"has_cpu"`
	Cpu    CpuSpec `json:"cpu,omitempty"`

	HasMemory bool       `json:"has_memory"`
	Memory    MemorySpec `json:"memory,omitempty"`

	HasNetwork bool `json:"has_network"`

	HasFilesystem bool `json:"has_filesystem"`

	// HasDiskIo when true, indicates that DiskIo stats will be available.
	HasDiskIo bool `json:"has_diskio"`
}

type ContainerStats struct {
	// The time of this stat point.
	Timestamp time.Time    `json:"timestamp"`
	Cpu       CpuStats     `json:"cpu,omitempty"`
	DiskIo    DiskIoStats  `json:"diskio,omitempty"`
	Memory    MemoryStats  `json:"memory,omitempty"`
	Network   NetworkStats `json:"network,omitempty"`

	// Filesystem statistics
	Filesystem []FsStats `json:"filesystem,omitempty"`

	// Task load stats
	TaskStats LoadStats `json:"task_stats,omitempty"`
}

// All CPU usage metrics are cumulative from the creation of the container
type CpuStats struct {
	Usage CpuUsage `json:"usage"`
	// Smoothed average of number of runnable threads x 1000.
	// We multiply by thousand to avoid using floats, but preserving precision.
	// Load is smoothed over the last 10 seconds. Instantaneous value can be read
	// from LoadStats.NrRunning.
	LoadAverage int32 `json:"load_average"`
}

type PerDiskStats struct {
	Major uint64            `json:"major"`
	Minor uint64            `json:"minor"`
	Stats map[string]uint64 `json:"stats"`
}

type DiskIoStats struct {
	IoServiceBytes []PerDiskStats `json:"io_service_bytes,omitempty"`
	IoServiced     []PerDiskStats `json:"io_serviced,omitempty"`
	IoQueued       []PerDiskStats `json:"io_queued,omitempty"`
	Sectors        []PerDiskStats `json:"sectors,omitempty"`
	IoServiceTime  []PerDiskStats `json:"io_service_time,omitempty"`
	IoWaitTime     []PerDiskStats `json:"io_wait_time,omitempty"`
	IoMerged       []PerDiskStats `json:"io_merged,omitempty"`
	IoTime         []PerDiskStats `json:"io_time,omitempty"`
}

type MemoryStats struct {
	// Current memory usage, this includes all memory regardless of when it was
	// accessed.
	// Units: Bytes.
	Usage uint64 `json:"usage"`

	// The amount of working set memory, this includes recently accessed memory,
	// dirty memory, and kernel memory. Working set is <= "usage".
	// Units: Bytes.
	WorkingSet uint64 `json:"working_set"`

	ContainerData    MemoryStatsMemoryData `json:"container_data,omitempty"`
	HierarchicalData MemoryStatsMemoryData `json:"hierarchical_data,omitempty"`
}

type InterfaceStats struct {
	// The name of the interface.
	Name string `json:"name"`
	// Cumulative count of bytes received.
	RxBytes uint64 `json:"rx_bytes"`
	// Cumulative count of packets received.
	RxPackets uint64 `json:"rx_packets"`
	// Cumulative count of receive errors encountered.
	RxErrors uint64 `json:"rx_errors"`
	// Cumulative count of packets dropped while receiving.
	RxDropped uint64 `json:"rx_dropped"`
	// Cumulative count of bytes transmitted.
	TxBytes uint64 `json:"tx_bytes"`
	// Cumulative count of packets transmitted.
	TxPackets uint64 `json:"tx_packets"`
	// Cumulative count of transmit errors encountered.
	TxErrors uint64 `json:"tx_errors"`
	// Cumulative count of packets dropped while transmitting.
	TxDropped uint64 `json:"tx_dropped"`
}

type NetworkStats struct {
	InterfaceStats `json:",inline"`
	Interfaces     []InterfaceStats `json:"interfaces,omitempty"`
	// TCP connection stats (Established, Listen...)
	Tcp TcpStat `json:"tcp"`
	// TCP6 connection stats (Established, Listen...)
	Tcp6 TcpStat `json:"tcp6"`
}

type TcpStat struct {
	//Count of TCP connections in state "Established"
	Established uint64
	//Count of TCP connections in state "Syn_Sent"
	SynSent uint64
	//Count of TCP connections in state "Syn_Recv"
	SynRecv uint64
	//Count of TCP connections in state "Fin_Wait1"
	FinWait1 uint64
	//Count of TCP connections in state "Fin_Wait2"
	FinWait2 uint64
	//Count of TCP connections in state "Time_Wait
	TimeWait uint64
	//Count of TCP connections in state "Close"
	Close uint64
	//Count of TCP connections in state "Close_Wait"
	CloseWait uint64
	//Count of TCP connections in state "Listen_Ack"
	LastAck uint64
	//Count of TCP connections in state "Listen"
	Listen uint64
	//Count of TCP connections in state "Closing"
	Closing uint64
}

type FsStats struct {
	// The block device name associated with the filesystem.
	Device string `json:"device,omitempty"`

	// Number of bytes that can be consumed by the container on this filesystem.
	Limit uint64 `json:"capacity"`

	// Number of bytes that is consumed by the container on this filesystem.
	Usage uint64 `json:"usage"`

	// Number of bytes available for non-root user.
	Available uint64 `json:"available"`

	// Number of reads completed
	// This is the total number of reads completed successfully.
	ReadsCompleted uint64 `json:"reads_completed"`

	// Number of reads merged
	// Reads and writes which are adjacent to each other may be merged for
	// efficiency.  Thus two 4K reads may become one 8K read before it is
	// ultimately handed to the disk, and so it will be counted (and queued)
	// as only one I/O.  This field lets you know how often this was done.
	ReadsMerged uint64 `json:"reads_merged"`

	// Number of sectors read
	// This is the total number of sectors read successfully.
	SectorsRead uint64 `json:"sectors_read"`

	// Number of milliseconds spent reading
	// This is the total number of milliseconds spent by all reads (as
	// measured from __make_request() to end_that_request_last()).
	ReadTime uint64 `json:"read_time"`

	// Number of writes completed
	// This is the total number of writes completed successfully.
	WritesCompleted uint64 `json:"writes_completed"`

	// Number of writes merged
	// See the description of reads merged.
	WritesMerged uint64 `json:"writes_merged"`

	// Number of sectors written
	// This is the total number of sectors written successfully.
	SectorsWritten uint64 `json:"sectors_written"`

	// Number of milliseconds spent writing
	// This is the total number of milliseconds spent by all writes (as
	// measured from __make_request() to end_that_request_last()).
	WriteTime uint64 `json:"write_time"`

	// Number of I/Os currently in progress
	// The only field that should go to zero. Incremented as requests are
	// given to appropriate struct request_queue and decremented as they finish.
	IoInProgress uint64 `json:"io_in_progress"`

	// Number of milliseconds spent doing I/Os
	// This field increases so long as field 9 is nonzero.
	IoTime uint64 `json:"io_time"`

	// weighted number of milliseconds spent doing I/Os
	// This field is incremented at each I/O start, I/O completion, I/O
	// merge, or read of these stats by the number of I/Os in progress
	// (field 9) times the number of milliseconds spent doing I/O since the
	// last update of this field.  This can provide an easy measure of both
	// I/O completion time and the backlog that may be accumulating.
	WeightedIoTime uint64 `json:"weighted_io_time"`
}

type CpuSpec struct {
	Limit    float64 `json:"limit"`
	MaxLimit float64 `json:"max_limit"`
	Mask     string  `json:"mask,omitempty"`
}

type MemorySpec struct {
	// The amount of memory requested. Default is unlimited (-1).
	// Units: bytes.
	Limit float64 `json:"limit,omitempty"`

	// The amount of guaranteed memory.  Default is 0.
	// Units: bytes.
	Reservation float64 `json:"reservation,omitempty"`

	// The amount of swap space requested. Default is unlimited (-1).
	// Units: bytes.
	SwapLimit float64 `json:"swap_limit,omitempty"`
}

// CPU usage time statistics.
type CpuUsage struct {
	// Total CPU usage.
	// Units: nanoseconds
	Total float64 `json:"total"`

	// Per CPU/core usage of the container.
	// Unit: nanoseconds.
	PerCpu []float64 `json:"per_cpu_usage,omitempty"`

	// Time spent in user space.
	// Unit: nanoseconds
	User float64 `json:"user"`

	// Time spent in kernel space.
	// Unit: nanoseconds
	System float64 `json:"system"`
}

type MemoryStatsMemoryData struct {
	Pgfault    uint64 `json:"pgfault"`
	Pgmajfault uint64 `json:"pgmajfault"`
}

// This mirrors kernel internal structure.
type LoadStats struct {
	// Number of sleeping tasks.
	NrSleeping uint64 `json:"nr_sleeping"`

	// Number of running tasks.
	NrRunning uint64 `json:"nr_running"`

	// Number of tasks in stopped state
	NrStopped uint64 `json:"nr_stopped"`

	// Number of tasks in uninterruptible state
	NrUninterruptible uint64 `json:"nr_uninterruptible"`

	// Number of tasks waiting on IO
	NrIoWait uint64 `json:"nr_io_wait"`
}

// ContainerInfoQuery is used when users check a container info from the REST api.
// It specifies how much data users want to get about a container
type ContainerInfoRequest struct {
	// Max number of stats to return. Specify -1 for all stats currently available.
	// If start and end time are specified this limit is ignored.
	// Default: 60
	NumStats int `json:"num_stats,omitempty"`

	// Start time for which to query information.
	// If ommitted, the beginning of time is assumed.
	Start time.Time `json:"start,omitempty"`

	// End time for which to query information.
	// If ommitted, current time is assumed.
	End time.Time `json:"end,omitempty"`
}

type HaproxySession struct {
	Bin        int64  `json:"bin"`
	Bout       int64  `json:"bout"`
	Dreq       int64  `json:"dreq"`
	Dresp      int64  `json:"dresp"`
	Ereq       int64  `json:"ereq"`
	HttpResp1  int64  `json:"http_response.1xx"`
	HttpResp2  int64  `json:"http_response.2xx"`
	HttpResp3  int64  `json:"http_response.3xx"`
	HttpResp4  int64  `json:"http_response.4xx"`
	HttpResp5  int64  `json:"http_response.5xx`
	ProxyName  string `json:"proxyname"`
	Rate       int64  `json:"rate"`
	RateMax    int64  `json:"rate_max"`
	ReqRate    int64  `json:"req_rate"`
	ReqRateMax int64  `json:"req_rate_max"`
	ReqTot     int64  `json:"req_tot"`
	Scur       int64  `json:"scur"`
	ServerName string `json:"servername"`
	Smax       int64  `json:"smax"`
	Stot       int64  `json:"stot"`
}

type AppRequestInfo struct {
	AppName string `json:"appname"`
	ReqRate int64  `json:"reqrate"`
}

type InfluxAppRequestInfo struct {
	ClusterId string `json:"clusterid" influx:"tag"`
	AppName   string `json:"appname" influx:"tag"`
	ReqRate   int64  `json:"reqrate" influx:"field"`
}

type HostInstance struct {
	ClusterId     string  `json:"clusterId"`
	Instance      string  `json:"instance"`
	AppName       string  `json:"appname"`
	ContainerName string  `json:"containerName"`
	CpuUsed       float64 `json:"cpuUsed"`
	MemoryUsed    float64 `json:"memoryUsed"`
	StartTime     int64   `json:"startTime"`
}

type AppConfig struct {
	Id              int64               `json:"id"`
	Name            string              `json:"name"`
	Alias           string              `json:"alias"`
	Cid             int64               `json:"cid"`
	Instances       float64             `json:"instances"`
	Tasks           float64             `json:"tasks"`
	Cpus            float64             `json:"cpus"`
	Mem             float64             `json:"mem"`
	Cmd             string              `json:"cmd"`
	Envs            []map[string]string `json:"envs"`
	ImageName       string              `json:"imageName"`
	ImageVersion    string              `json:"imageVersion"`
	ForceImage      bool                `json:"forceImage"`
	Network         string              `json:"network"`
	Ports           []AppPort           `json:"ports"`
	Volumes         []AppVolume         `json:"volumes"`
	Unique          bool                `json:"unique"`
	Iplist          []string            `json:"iplist"`
	Update          *time.Time          `json:"update"`
	LogPaths        []string            `json:"logPaths"`
	Parameters      []map[string]string `json:"parameters"`
	Canary          uint8               `json:"canary"`
	CanaryInstances int64               `json:"canaryInstances"`
	Uid             int64               `json:"uid"`
}

// AppPort portmapping of application
type AppPort struct {
	AppID     int64  `db:"app_id" json:"appId,omitempty" structs:",omitempty"`
	Port      int64  `json:"appPort" structs:"appPort"`
	Type      uint8  `db:"svc_type" json:"type" structs:"type"`
	URI       string `db:"uri" json:"uri" structs:"uri"`
	Cid       string `json:"cid,omitempty" structs:",omitempty"`
	Proto     int64  `json:"protocol" structs:"protocol"`
	HasURI    int64  `db:"has_uri" json:"isUri" structs:"isUri"`
	MapPort   int64  `db:"bind_port" json:"mapPort" structs:"mapPort"`
	VersionId int64  `db:"ver_id" json:"versionId,omitempty" structs:",omitempty"`
	Status    int    `db:"status" json:"status,omitempty" structs:",omitempty"`
}

type AppVolume struct {
	HostPath      string `json:"hostPath"`
	ContainerPath string `json:"containerPath"`
}

type ClusterAppList struct {
	App   []AppConfig `json:"App"`
	Count int         `json:"Count"`
}

// response of omega-app /clusters/:cid/apps REST API
type ClusterAppListResp struct {
	Code int64          `json:"code"`
	Data ClusterAppList `json:"data"`
}

// app status struct
type AppStatus struct {
	Id          int64  `json:"id"`
	Cid         int64  `json:"cid"`
	Name        string `json:"name"`
	Alias       string `json:"alias"`
	Status      uint8  `json:"status"`
	Tasks       int64  `json:"tasks"`
	Instances   int64  `json:"instances"`
	LastFailure uint8  `json:"lastfailure"`
}

// response of omega-app /apps/status REST API
type AppStatusResp struct {
	Code int64                `json:"code"`
	Data map[string]AppStatus `json:"data"`
}
