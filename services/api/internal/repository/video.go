package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rmarathe-hub/StreamForce/services/api/internal/models"
)

type VideoRepository struct {
	pool *pgxpool.Pool
}

func NewVideoRepository(pool *pgxpool.Pool) *VideoRepository {
	return &VideoRepository{pool: pool}
}

func (r *VideoRepository) Create(ctx context.Context, video models.Video) (models.Video, error) {
	const query = `
		INSERT INTO videos (filename, status, source_path)
		VALUES ($1, $2, $3)
		RETURNING id, filename, status, source_path, duration, width, height, created_at, updated_at, error_message
	`

	row := r.pool.QueryRow(ctx, query, video.Filename, video.Status, video.SourcePath)
	return scanVideo(row)
}

func (r *VideoRepository) List(ctx context.Context) ([]models.Video, error) {
	const query = `
		SELECT id, filename, status, source_path, duration, width, height, created_at, updated_at, error_message
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
		SELECT id, filename, status, source_path, duration, width, height, created_at, updated_at, error_message
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

var ErrNotFound = fmt.Errorf("video not found")

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
