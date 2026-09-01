package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

const progressTTL = 24 * time.Hour

type ProgressStore struct {
	client *goredis.Client
}

func NewProgressStore(client *goredis.Client) *ProgressStore {
	return &ProgressStore{client: client}
}

func progressKey(videoID uuid.UUID) string {
	return fmt.Sprintf("streamforge:video:%s:progress", videoID.String())
}

func (s *ProgressStore) Set(ctx context.Context, videoID uuid.UUID, percent int) error {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return s.client.Set(ctx, progressKey(videoID), percent, progressTTL).Err()
}

func (s *ProgressStore) Get(ctx context.Context, videoID uuid.UUID) (int, bool, error) {
	val, err := s.client.Get(ctx, progressKey(videoID)).Int()
	if err == goredis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return val, true, nil
}

func (s *ProgressStore) Delete(ctx context.Context, videoID uuid.UUID) error {
	return s.client.Del(ctx, progressKey(videoID)).Err()
}
