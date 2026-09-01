package processor

import (
	"context"

	"github.com/google/uuid"
)

type ProgressReporter interface {
	Set(ctx context.Context, videoID uuid.UUID, percent int) error
	Delete(ctx context.Context, videoID uuid.UUID) error
}

type noopProgressReporter struct{}

func (noopProgressReporter) Set(context.Context, uuid.UUID, int) error  { return nil }
func (noopProgressReporter) Delete(context.Context, uuid.UUID) error     { return nil }

func NoopProgressReporter() ProgressReporter {
	return noopProgressReporter{}
}
