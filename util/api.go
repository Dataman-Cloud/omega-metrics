package util

import (
	"time"
)

const (
	MonitorMasterMetrics = "MasterMetricsMar"
	MonitorMarathonEvent = "MarathonEventMar"
	MonitorAppMetrics    = "SlaveStateMar"
)

type MonitorResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Err  string      `json:"error"`
}

type RabbitMqMessage struct {
	ClusterId int    `json:"clusterId"`
	NodeId    string `json:"nodeId"`
	Message   string `json:"message"`
	Attached  string `json:"attached"`
	Timestamp int64  `json:"timestamp"`
}

// master metrics
type MasterMetrics struct {
	CpuPercent float64 `json:"master/cpus_percent"`
	DiskUsed   int     `json:"master/disk_used"`
	DiskTotal  int     `json:"master/disk_total"`
	MemUsed    int     `json:"master/mem_used"`
	MemTotal   int     `json:"master/mem_total"`
	Leader     int     `json:"master/elected"`
}

type MasterMetricsMar struct {
	CpuPercent float64 `json:"cpuPercent"`
	MemTotal   int     `json:"memTotal"`
	MemUsed    int     `json:"memUsed"`
	DiskTotal  int     `json:"diskTotal"`
	DiskUsed   int     `json:"diskUsed"`
	Timestamp  int64   `json:"timestamp"`
	Leader     int     `json:"leader"`
	ClusterId  string  `json:"clusterId"`
}

// marathon event
type MarathonEvent struct {
	EventType   string      `json:"eventType"`
	Timestamp   string      `json:"timestamp"`
	Id          string      `json:"id,omitempty"`
	Plan        plan        `json:"plan,omitempty"`
	CurrentStep currentStep `json:"currentStep,omitempty"`
}

type MarathonEventMar struct {
	EventType   string       `json:"eventType"`
	Timestamp   string       `json:"timestamp"`
	App         appNameAndId `json:"app"`
	CurrentType string       `json:"currentType"`
	TaskId      string       `json:"taskId"`
	ClusterId   string       `json:"clusterId"`
}

type StatusUpdate struct {
	EventType  string `json:"eventType"`
	Timestamp  string `json:"timestamp"`
	AppId      string `json:"appId"`
	Host       string `json:"host"`
	Ports      []int  `json:"ports"`
	TaskStatus string `json:"taskStatus"`
}

type DestroyApp struct {
	EventType string `json:"eventType"`
	Timestamp string `json:"timestamp"`
	AppId     string `json:"appId"`
}

type currentStep struct {
	Actions []actions
}

type actions struct {
	Type string
	App  string
}

type plan struct {
	Id string
}

// slave state
type SlaveStateMar struct {
	Timestamp     int64        `json:"timestamp"`
	ClusterId     string       `json:"clusterId"`
	App           appNameAndId `json:"app"`
	ContainerId   string       `json:"containerId"`
	CpuUsed       uint64       `json:"cpuUsed"`
	CpuTotal      uint64       `json:"cpuTotal"`
	CpuShareCores float64      `json:"cpuShareCores"`
	MemoryTotal   uint64       `json:"memoryTotal"`
	MemoryUsed    uint64       `json:"memoryUsed"`
}

type SlaveState struct {
	Hostname   string      `json:"hostname"`
	Id         string      `json:"id"`
	Frameworks []framework `json:"frameworks"`
	Flags      flag        `json:"flags"`
}

type flag struct {
	Ip string `json:"ip"`
}

type framework struct {
	Id        string     `json:"id"`
	Name      string     `json:"name"`
	Hostname  string     `json:"hostname"`
	Executors []executor `json:"executors"`
}

type executor struct {
	Container string    `json:"container"`
	Id        string    `json:"id"`
	Directory string    `json:"directory"`
	Resources resources `json:"resources"`
}

type resources struct {
	Ports string `json:"ports,omitempty"`
}

type appNameAndId struct {
	AppName string `json:"appName"`
	AppId   string `json:"appId"`
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

type NetworkStats struct {
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
	Limit    uint64 `json:"limit"`
	MaxLimit uint64 `json:"max_limit"`
	Mask     string `json:"mask,omitempty"`
}

type MemorySpec struct {
	// The amount of memory requested. Default is unlimited (-1).
	// Units: bytes.
	Limit uint64 `json:"limit,omitempty"`

	// The amount of guaranteed memory.  Default is 0.
	// Units: bytes.
	Reservation uint64 `json:"reservation,omitempty"`

	// The amount of swap space requested. Default is unlimited (-1).
	// Units: bytes.
	SwapLimit uint64 `json:"swap_limit,omitempty"`
}

// CPU usage time statistics.
type CpuUsage struct {
	// Total CPU usage.
	// Units: nanoseconds
	Total uint64 `json:"total"`

	// Per CPU/core usage of the container.
	// Unit: nanoseconds.
	PerCpu []uint64 `json:"per_cpu_usage,omitempty"`

	// Time spent in user space.
	// Unit: nanoseconds
	User uint64 `json:"user"`

	// Time spent in kernel space.
	// Unit: nanoseconds
	System uint64 `json:"system"`
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
