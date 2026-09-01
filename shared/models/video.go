package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusUploaded   = "UPLOADED"
	StatusQueued     = "QUEUED"
	StatusProcessing = "PROCESSING"
	StatusReady      = "READY"
	StatusFailed     = "FAILED"
)

type Video struct {
	ID           uuid.UUID  `json:"id"`
	Filename     string     `json:"filename"`
	Status       string     `json:"status"`
	SourcePath   string     `json:"source_path"`
	HLSPath      *string    `json:"hls_path"`
	Codec        *string    `json:"codec"`
	Duration     *float64   `json:"duration"`
	Width        *int       `json:"width"`
	Height       *int       `json:"height"`
	ClaimedBy    *string    `json:"claimed_by,omitempty"`
	ClaimedAt    *time.Time `json:"claimed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ErrorMessage *string    `json:"error_message"`
}
