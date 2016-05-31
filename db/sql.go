package db

const (
	QueryMetricFormatSql = `
		select
			time,
			appname,
			instance,
			CpuShareCores as cpuUsed,
			CpuShareCores as cpuTotal,
			DiskIOReadBytesRate as diskReadRate,
			DiskIOWriteBytesRate as diskWriteRate,
			MemoryTotal as memoryTotal,
			MemoryUsed  as memoryUsed,
			NetworkReceviedByteRate  as networkRecevied,
			NetworkSentByteRate as networkSend
			from  Slave_state
			where clusterid = '%s'
			and appname = '%s'
			and time >= %d
			and time <= %d
			order by time desc
	`
)
