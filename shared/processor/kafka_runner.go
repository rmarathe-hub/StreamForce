package processor

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/rmarathe-hub/StreamForce/shared/kafka"
	"github.com/rmarathe-hub/StreamForce/shared/models"
	"github.com/rmarathe-hub/StreamForce/shared/repository"
	segmentio "github.com/segmentio/kafka-go"
)

const (
	claimStaleAfter    = 10 * time.Minute
	claimRefreshEvery  = 2 * time.Minute
)

type KafkaRunner struct {
	consumer      *kafka.Consumer
	repo          *repository.VideoRepository
	processor     *Processor
	workerID      string
	consumerGroup string
}

func NewKafkaRunner(
	consumer *kafka.Consumer,
	repo *repository.VideoRepository,
	processor *Processor,
	workerID string,
	consumerGroup string,
) *KafkaRunner {
	return &KafkaRunner{
		consumer:      consumer,
		repo:          repo,
		processor:     processor,
		workerID:      workerID,
		consumerGroup: consumerGroup,
	}
}

func (r *KafkaRunner) Run(ctx context.Context) error {
	log.Printf(
		`{"worker_id":"%s","event":"kafka_worker_started","consumer_group":%q}`,
		r.workerID,
		r.consumerGroup,
	)

	for {
		message, err := r.consumer.Fetch(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Printf(`{"worker_id":"%s","event":"kafka_worker_stopped"}`, r.workerID)
				return nil
			}
			return err
		}

		commit, err := r.handleMessage(ctx, message)
		if err != nil {
			log.Printf(
				`{"worker_id":"%s","event":"kafka_message_failed","error":%q,"commit":%t}`,
				r.workerID,
				err.Error(),
				commit,
			)
		}

		if !commit {
			continue
		}

		if err := r.consumer.Commit(ctx, message); err != nil {
			return err
		}
	}
}

func (r *KafkaRunner) handleMessage(ctx context.Context, message segmentio.Message) (bool, error) {
	job, err := kafka.ParseVideoJobMessage(message.Value)
	if err != nil {
		return true, err
	}

	log.Printf(
		`{"worker_id":"%s","video_id":"%s","event":"kafka_message_received","event_id":"%s","attempt":%d,"partition":%d,"offset":%d}`,
		r.workerID,
		job.VideoID,
		job.EventID,
		job.Attempt,
		message.Partition,
		message.Offset,
	)

	video, err := r.repo.GetByID(ctx, job.VideoID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			log.Printf(
				`{"worker_id":"%s","video_id":"%s","event":"processing_skipped_not_found"}`,
				r.workerID,
				job.VideoID,
			)
			return true, nil
		}
		return false, err
	}

	if video.Status == models.StatusReady {
		log.Printf(
			`{"worker_id":"%s","video_id":"%s","event":"processing_skipped_already_ready"}`,
			r.workerID,
			job.VideoID,
		)
		return true, nil
	}

	video, err = r.repo.TryClaimForProcessing(ctx, job.VideoID, r.workerID, claimStaleAfter)
	if err != nil {
		if errors.Is(err, repository.ErrNotClaimable) {
			log.Printf(
				`{"worker_id":"%s","video_id":"%s","event":"processing_deferred_in_progress"}`,
				r.workerID,
				job.VideoID,
			)
			return false, err
		}
		return false, err
	}

	if video.Status == models.StatusReady {
		log.Printf(
			`{"worker_id":"%s","video_id":"%s","event":"processing_skipped_already_ready"}`,
			r.workerID,
			job.VideoID,
		)
		return true, nil
	}

	claimCtx, stopRefreshing := context.WithCancel(ctx)
	defer stopRefreshing()
	go r.refreshClaimLoop(claimCtx, job.VideoID)

	log.Printf(
		`{"worker_id":"%s","video_id":"%s","event":"processing_started","filename":%q}`,
		r.workerID,
		video.ID,
		video.Filename,
	)

	if err := r.processor.Process(ctx, video); err != nil {
		log.Printf(
			`{"worker_id":"%s","video_id":"%s","event":"processing_failed","error":%q}`,
			r.workerID,
			video.ID,
			err.Error(),
		)
		return true, err
	}

	log.Printf(
		`{"worker_id":"%s","video_id":"%s","event":"processing_completed"}`,
		r.workerID,
		video.ID,
	)
	return true, nil
}

func (r *KafkaRunner) refreshClaimLoop(ctx context.Context, videoID uuid.UUID) {
	ticker := time.NewTicker(claimRefreshEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.repo.RefreshClaim(ctx, videoID, r.workerID); err != nil {
				log.Printf(
					`{"worker_id":"%s","video_id":"%s","event":"claim_refresh_failed","error":%q}`,
					r.workerID,
					videoID,
					err.Error(),
				)
			}
		}
	}
}
