package redis

import (
	"context"

	"github.com/google/uuid"
	"github.com/rmarathe-hub/StreamForce/shared/models"
)

type NotifyingProgressReporter struct {
	store    *ProgressStore
	events   *EventPublisher
	workerID string
}

func NewNotifyingProgressReporter(
	store *ProgressStore,
	events *EventPublisher,
	workerID string,
) *NotifyingProgressReporter {
	return &NotifyingProgressReporter{
		store:    store,
		events:   events,
		workerID: workerID,
	}
}

func (r *NotifyingProgressReporter) Set(ctx context.Context, videoID uuid.UUID, percent int) error {
	if err := r.store.Set(ctx, videoID, percent); err != nil {
		return err
	}

	workerID := r.workerID
	return r.events.Publish(ctx, VideoEvent{
		VideoID:         videoID.String(),
		Status:          models.StatusProcessing,
		ProgressPercent: &percent,
		ClaimedBy:       &workerID,
	})
}

func (r *NotifyingProgressReporter) Delete(ctx context.Context, videoID uuid.UUID) error {
	return r.store.Delete(ctx, videoID)
}

func VideoEventFromModel(video models.Video, progressPercent *int) VideoEvent {
	event := VideoEvent{
		VideoID:      video.ID.String(),
		Status:       video.Status,
		ClaimedBy:    video.ClaimedBy,
		HLSPath:      video.HLSPath,
		ThumbnailPath: video.ThumbnailPath,
		Codec:        video.Codec,
		Duration:     video.Duration,
		Width:        video.Width,
		Height:       video.Height,
		ErrorMessage: video.ErrorMessage,
	}
	if progressPercent != nil {
		event.ProgressPercent = progressPercent
	}
	return event
}
