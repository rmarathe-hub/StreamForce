package handlers

import (
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
	"github.com/rmarathe-hub/StreamForce/shared/models"
	"github.com/rmarathe-hub/StreamForce/shared/repository"
)

type Handler struct {
	repo    *repository.VideoRepository
	storage *storage.LocalStorage
	cfg     config.Config
}

func New(repo *repository.VideoRepository, store *storage.LocalStorage, cfg config.Config) *Handler {
	return &Handler{repo: repo, storage: store, cfg: cfg}
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
		Status:     models.StatusUploaded,
		SourcePath: sourcePath,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create video record")
		return
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
