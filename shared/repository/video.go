package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rmarathe-hub/StreamForce/shared/models"
)

var (
	ErrNotFound = fmt.Errorf("video not found")
	ErrNoJob    = errors.New("no uploaded videos to process")
)

type VideoRepository struct {
	pool *pgxpool.Pool
}

func NewVideoRepository(pool *pgxpool.Pool) *VideoRepository {
	return &VideoRepository{pool: pool}
}

const videoColumns = `
	id, filename, status, source_path, hls_path, codec,
	duration, width, height, created_at, updated_at, error_message
`

func (r *VideoRepository) Create(ctx context.Context, video models.Video) (models.Video, error) {
	const query = `
		INSERT INTO videos (filename, status, source_path)
		VALUES ($1, $2, $3)
		RETURNING ` + videoColumns

	row := r.pool.QueryRow(ctx, query, video.Filename, video.Status, video.SourcePath)
	return scanVideo(row)
}

func (r *VideoRepository) List(ctx context.Context) ([]models.Video, error) {
	const query = `
		SELECT ` + videoColumns + `
		FROM videos
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list videos: %w", err)
	}
	defer rows.Close()

	var videos []models.Video
	for rows.Next() {
		video, err := scanVideo(rows)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}

	if videos == nil {
		videos = []models.Video{}
	}

	return videos, rows.Err()
}

func (r *VideoRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Video, error) {
	const query = `
		SELECT ` + videoColumns + `
		FROM videos
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	video, err := scanVideo(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Video{}, ErrNotFound
		}
		return models.Video{}, err
	}

	return video, nil
}

func (r *VideoRepository) ClaimNextUploaded(ctx context.Context) (models.Video, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.Video{}, fmt.Errorf("begin claim transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const selectQuery = `
		SELECT ` + videoColumns + `
		FROM videos
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`

	row := tx.QueryRow(ctx, selectQuery, models.StatusUploaded)
	video, err := scanVideo(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Video{}, ErrNoJob
		}
		return models.Video{}, fmt.Errorf("claim uploaded video: %w", err)
	}

	const updateQuery = `
		UPDATE videos
		SET status = $2, updated_at = NOW(), error_message = NULL
		WHERE id = $1
	`
	if _, err := tx.Exec(ctx, updateQuery, video.ID, models.StatusProcessing); err != nil {
		return models.Video{}, fmt.Errorf("mark claimed video processing: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Video{}, fmt.Errorf("commit claim transaction: %w", err)
	}

	video.Status = models.StatusProcessing
	return video, nil
}

func (r *VideoRepository) MarkProcessing(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE videos
		SET status = $2, updated_at = NOW(), error_message = NULL
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id, models.StatusProcessing)
	return err
}

func (r *VideoRepository) ResetInterruptedProcessing(ctx context.Context) (int64, error) {
	const query = `
		UPDATE videos
		SET status = $1, updated_at = NOW()
		WHERE status = $2
	`
	tag, err := r.pool.Exec(ctx, query, models.StatusQueued, models.StatusProcessing)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *VideoRepository) MarkReady(
	ctx context.Context,
	id uuid.UUID,
	hlsPath string,
	duration float64,
	width, height int,
	codec string,
) error {
	const query = `
		UPDATE videos
		SET status = $2,
		    hls_path = $3,
		    duration = $4,
		    width = $5,
		    height = $6,
		    codec = $7,
		    updated_at = NOW(),
		    error_message = NULL
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id, models.StatusReady, hlsPath, duration, width, height, codec)
	return err
}

func (r *VideoRepository) MarkFailed(ctx context.Context, id uuid.UUID, message string) error {
	const query = `
		UPDATE videos
		SET status = $2, error_message = $3, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id, models.StatusFailed, message)
	return err
}

type scannable interface {
	Scan(dest ...any) error
}

func scanVideo(row scannable) (models.Video, error) {
	var video models.Video
	err := row.Scan(
		&video.ID,
		&video.Filename,
		&video.Status,
		&video.SourcePath,
		&video.HLSPath,
		&video.Codec,
		&video.Duration,
		&video.Width,
		&video.Height,
		&video.CreatedAt,
		&video.UpdatedAt,
		&video.ErrorMessage,
	)
	if err != nil {
		return models.Video{}, err
	}
	return video, nil
}
