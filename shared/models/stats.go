package models

type SystemStats struct {
	TotalVideos      int            `json:"total_videos"`
	ByStatus         map[string]int `json:"by_status"`
	ProcessingCount  int            `json:"processing_count"`
	QueuedCount      int            `json:"queued_count"`
	ReadyCount       int            `json:"ready_count"`
	FailedCount      int            `json:"failed_count"`
	ActiveWorkers    []string       `json:"active_workers"`
}
