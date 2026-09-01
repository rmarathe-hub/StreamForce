package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type JobPublisher struct {
	producer *Producer
}

func NewJobPublisher(producer *Producer) *JobPublisher {
	return &JobPublisher{producer: producer}
}

func (p *JobPublisher) PublishVideoJob(ctx context.Context, videoID uuid.UUID, sourcePath string, attempt int) error {
	message := VideoJobMessage{
		EventID:    uuid.New(),
		VideoID:    videoID,
		SourcePath: sourcePath,
		Attempt:    attempt,
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal video job: %w", err)
	}

	if err := p.producer.Publish(ctx, videoID.String(), payload); err != nil {
		return fmt.Errorf("publish video job: %w", err)
	}

	return nil
}

func (p *JobPublisher) Close() error {
	return p.producer.Close()
}
