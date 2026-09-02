package redis

import (
	"context"
	"encoding/json"

	goredis "github.com/redis/go-redis/v9"
)

const VideoEventsChannel = "streamforge:video:events"

type VideoEvent struct {
	VideoID         string   `json:"video_id"`
	Status          string   `json:"status"`
	ProgressPercent *int     `json:"progress_percent,omitempty"`
	ClaimedBy       *string  `json:"claimed_by,omitempty"`
	HLSPath         *string  `json:"hls_path,omitempty"`
	ThumbnailPath   *string  `json:"thumbnail_path,omitempty"`
	Codec           *string  `json:"codec,omitempty"`
	Duration        *float64 `json:"duration,omitempty"`
	Width           *int     `json:"width,omitempty"`
	Height          *int     `json:"height,omitempty"`
	ErrorMessage    *string  `json:"error_message,omitempty"`
}

type EventPublisher struct {
	client *goredis.Client
}

func NewEventPublisher(client *goredis.Client) *EventPublisher {
	return &EventPublisher{client: client}
}

func (p *EventPublisher) Publish(ctx context.Context, event VideoEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.client.Publish(ctx, VideoEventsChannel, payload).Err()
}

func SubscribeVideoEvents(ctx context.Context, client *goredis.Client) *goredis.PubSub {
	return client.Subscribe(ctx, VideoEventsChannel)
}
