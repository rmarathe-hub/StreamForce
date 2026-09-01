package processor

import (
	"context"
	"errors"
	"log"

	"github.com/rmarathe-hub/StreamForce/shared/kafka"
	"github.com/rmarathe-hub/StreamForce/shared/models"
	"github.com/rmarathe-hub/StreamForce/shared/repository"
	segmentio "github.com/segmentio/kafka-go"
)

type KafkaRunner struct {
	consumer  *kafka.Consumer
	repo      *repository.VideoRepository
	processor *Processor
	workerID  string
}

func NewKafkaRunner(
	consumer *kafka.Consumer,
	repo *repository.VideoRepository,
	processor *Processor,
	workerID string,
) *KafkaRunner {
	return &KafkaRunner{
		consumer:  consumer,
		repo:      repo,
		processor: processor,
		workerID:  workerID,
	}
}

func (r *KafkaRunner) Run(ctx context.Context) error {
	log.Printf(`{"worker_id":"%s","event":"kafka_worker_started"}`, r.workerID)

	for {
		message, err := r.consumer.Fetch(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Printf(`{"worker_id":"%s","event":"kafka_worker_stopped"}`, r.workerID)
				return nil
			}
			return err
		}

		if err := r.handleMessage(ctx, message); err != nil {
			log.Printf(
				`{"worker_id":"%s","event":"kafka_message_failed","error":%q}`,
				r.workerID,
				err.Error(),
			)
		}

		if err := r.consumer.Commit(ctx, message); err != nil {
			return err
		}
	}
}

func (r *KafkaRunner) handleMessage(ctx context.Context, message segmentio.Message) error {
	job, err := kafka.ParseVideoJobMessage(message.Value)
	if err != nil {
		return err
	}

	log.Printf(
		`{"worker_id":"%s","video_id":"%s","event":"kafka_message_received","event_id":"%s","attempt":%d}`,
		r.workerID,
		job.VideoID,
		job.EventID,
		job.Attempt,
	)

	video, err := r.repo.GetByID(ctx, job.VideoID)
	if err != nil {
		return err
	}

	if video.Status == models.StatusReady {
		log.Printf(
			`{"worker_id":"%s","video_id":"%s","event":"processing_skipped_already_ready"}`,
			r.workerID,
			job.VideoID,
		)
		return nil
	}

	if video.Status == models.StatusProcessing {
		log.Printf(
			`{"worker_id":"%s","video_id":"%s","event":"processing_skipped_in_progress"}`,
			r.workerID,
			job.VideoID,
		)
		return nil
	}

	if err := r.repo.MarkProcessing(ctx, job.VideoID); err != nil {
		return err
	}

	video.Status = models.StatusProcessing

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
		return err
	}

	log.Printf(
		`{"worker_id":"%s","video_id":"%s","event":"processing_completed"}`,
		r.workerID,
		video.ID,
	)
	return nil
}
