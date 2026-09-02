package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rmarathe-hub/StreamForce/services/api/internal/config"
	"github.com/rmarathe-hub/StreamForce/services/api/internal/storage"
	"github.com/rmarathe-hub/StreamForce/services/api/internal/ws"
	"github.com/rmarathe-hub/StreamForce/shared/models"
	"github.com/rmarathe-hub/StreamForce/shared/redis"
	"github.com/rmarathe-hub/StreamForce/shared/repository"
)

type JobPublisher interface {
	PublishVideoJob(ctx context.Context, videoID uuid.UUID, sourcePath string, attempt int) error
}

type ProgressReader interface {
	Get(ctx context.Context, videoID uuid.UUID) (int, bool, error)
}

type VideoEventPublisher interface {
	Publish(ctx context.Context, event redis.VideoEvent) error
}

type Handler struct {
	repo      *repository.VideoRepository
	storage   *storage.LocalStorage
	publisher JobPublisher
	progress  ProgressReader
	events    VideoEventPublisher
	hub       *ws.Hub
	cfg       config.Config
}

func New(
	repo *repository.VideoRepository,
	store *storage.LocalStorage,
	publisher JobPublisher,
	progress ProgressReader,
	events VideoEventPublisher,
	hub *ws.Hub,
	cfg config.Config,
) *Handler {
	return &Handler{
		repo:      repo,
		storage:   store,
		publisher: publisher,
		progress:  progress,
		events:    events,
		hub:       hub,
		cfg:       cfg,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ListVideos(w http.ResponseWriter, r *http.Request) {
	videos, err := h.repo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list videos")
		return
	}
	writeJSON(w, http.StatusOK, videos)
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.GetStats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (h *Handler) GetVideo(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid video id")
		return
	}

	video, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "video not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get video")
		return
	}

	if video.Status == models.StatusProcessing && h.progress != nil {
		if percent, found, err := h.progress.Get(r.Context(), id); err == nil {
			if found {
				video.ProgressPercent = &percent
			} else {
				zero := 0
				video.ProgressPercent = &zero
			}
		}
	}

	writeJSON(w, http.StatusOK, video)
}

func (h *Handler) CreateVideo(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(h.cfg.MaxUploadMB << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	filename := filepath.Base(header.Filename)
	if filename == "" || filename == "." {
		writeError(w, http.StatusBadRequest, "invalid filename")
		return
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".mp4" && ext != ".mov" && ext != ".webm" && ext != ".mkv" {
		writeError(w, http.StatusBadRequest, "unsupported file type; upload an MP4, MOV, WEBM, or MKV video")
		return
	}

	limited := io.LimitReader(file, (h.cfg.MaxUploadMB<<20)+1)
	sourcePath, err := h.storage.SaveUpload(filename, limited)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save video")
		return
	}

	video, err := h.repo.Create(r.Context(), models.Video{
		Filename:   filename,
		Status:     models.StatusQueued,
		SourcePath: sourcePath,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create video record")
		return
	}

	if err := h.publisher.PublishVideoJob(r.Context(), video.ID, video.SourcePath, 1); err != nil {
		_ = h.repo.MarkFailed(r.Context(), video.ID, "failed to publish kafka job: "+err.Error())
		writeError(w, http.StatusInternalServerError, "failed to queue video for processing")
		return
	}

	if h.events != nil {
		_ = h.events.Publish(r.Context(), redis.VideoEventFromModel(video, nil))
	}

	writeJSON(w, http.StatusCreated, video)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, fmt.Sprintf("route not found: %s", r.URL.Path))
}
