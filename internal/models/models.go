package models

import "time"

type UploadResponse struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

type JobStatus string

const (
	JobStatusInProgress JobStatus = "IN_PROGRESS"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
)

type ProcessingJob struct {
	ID                string    `json:"id"`
	Status            JobStatus `json:"status"`
	OriginalFileName  string    `json:"originalFileName"`
	ProcessedFilePath string    `json:"processedFilePath"`
	CreatedAt         time.Time `json:"createdAt"`
}

func DSProcessingJob(id, originalFileName string) *ProcessingJob {
	return &ProcessingJob{
		ID:               id,
		OriginalFileName: originalFileName,
		Status:           JobStatusInProgress,
		CreatedAt:        time.Now(),
	}
}
