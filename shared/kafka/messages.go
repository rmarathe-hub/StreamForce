package kafka

import (
	"encoding/json"

	"github.com/google/uuid"
)

type VideoJobMessage struct {
	EventID    uuid.UUID `json:"eventId"`
	VideoID    uuid.UUID `json:"videoId"`
	SourcePath string    `json:"sourcePath"`
	Attempt    int       `json:"attempt"`
}

func (m VideoJobMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func ParseVideoJobMessage(data []byte) (VideoJobMessage, error) {
	var message VideoJobMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return VideoJobMessage{}, err
	}
	return message, nil
}
