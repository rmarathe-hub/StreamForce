package processor

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/rmarathe-hub/StreamForce/shared/repository"
)

type Runner struct {
	repo         *repository.VideoRepository
	processor    *Processor
	pollInterval time.Duration
	workerID     string
}

func NewRunner(
	repo *repository.VideoRepository,
	processor *Processor,
	workerID string,
	pollInterval time.Duration,
) *Runner {
	return &Runner{
		repo:         repo,
		processor:    processor,
		pollInterval: pollInterval,
		workerID:     workerID,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	log.Printf(`{"worker_id":"%s","event":"worker_started"}`, r.workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf(`{"worker_id":"%s","event":"worker_stopped"}`, r.workerID)
			return nil
		default:
		}

		processed, err := r.processNext(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			log.Printf(`{"worker_id":"%s","event":"processing_error","error":%q}`, r.workerID, err.Error())
		}

		if processed {
			continue
		}

		select {
		case <-ctx.Done():
			log.Printf(`{"worker_id":"%s","event":"worker_stopped"}`, r.workerID)
			return nil
		case <-time.After(r.pollInterval):
		}
	}
}

func (r *Runner) processNext(ctx context.Context) (bool, error) {
	video, err := r.repo.ClaimNextUploaded(ctx)
	if errors.Is(err, repository.ErrNoJob) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

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
