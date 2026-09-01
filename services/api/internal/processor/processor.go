package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/rmarathe-hub/StreamForce/services/api/internal/config"
	"github.com/rmarathe-hub/StreamForce/services/api/internal/models"
	"github.com/rmarathe-hub/StreamForce/services/api/internal/repository"
)

type Processor struct {
	repo    *repository.VideoRepository
	cfg     config.Config
	baseDir string
}

func New(repo *repository.VideoRepository, cfg config.Config) *Processor {
	return &Processor{
		repo:    repo,
		cfg:     cfg,
		baseDir: cfg.StoragePath,
	}
}

func (p *Processor) Enqueue(videoID uuid.UUID) {
	go func() {
		if err := p.Process(context.Background(), videoID); err != nil {
			log.Printf("video %s processing failed: %v", videoID, err)
		}
	}()
}

func (p *Processor) RecoverPending(ctx context.Context) error {
	videos, err := p.repo.ListByStatus(ctx, models.StatusUploaded)
	if err != nil {
		return err
	}

	for _, video := range videos {
		log.Printf("recovering pending video %s", video.ID)
		p.Enqueue(video.ID)
	}

	return nil
}

func (p *Processor) Process(ctx context.Context, videoID uuid.UUID) error {
	video, err := p.repo.GetByID(ctx, videoID)
	if err != nil {
		return err
	}

	if video.Status != models.StatusUploaded && video.Status != models.StatusFailed {
		return nil
	}

	if err := p.repo.MarkProcessing(ctx, videoID); err != nil {
		return fmt.Errorf("mark processing: %w", err)
	}

	sourcePath := filepath.Join(p.baseDir, video.SourcePath)
	meta, err := probeVideo(ctx, p.cfg.FFprobePath, sourcePath)
	if err != nil {
		_ = p.repo.MarkFailed(ctx, videoID, err.Error())
		return err
	}

	outputDir := filepath.Join(p.baseDir, "hls", videoID.String())
	if err := os.RemoveAll(outputDir); err != nil {
		_ = p.repo.MarkFailed(ctx, videoID, "failed to prepare output directory")
		return err
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		_ = p.repo.MarkFailed(ctx, videoID, "failed to create output directory")
		return err
	}

	heights := targetHeights(meta.Height)
	variants := make([]variantInfo, 0, len(heights))

	for _, height := range heights {
		label := fmt.Sprintf("%dp", height)
		variantDir := filepath.Join(outputDir, label)
		if err := os.MkdirAll(variantDir, 0o755); err != nil {
			_ = p.repo.MarkFailed(ctx, videoID, "failed to create variant directory")
			return err
		}

		if err := transcodeVariant(ctx, p.cfg.FFmpegPath, sourcePath, variantDir, height); err != nil {
			_ = p.repo.MarkFailed(ctx, videoID, err.Error())
			return err
		}

		variants = append(variants, variantInfo{
			Label:     label,
			Width:     scaledWidth(meta.Width, meta.Height, height),
			Height:    height,
			Bandwidth: bandwidthForHeight(height),
		})
	}

	masterPath := filepath.Join(outputDir, "master.m3u8")
	if err := writeMasterPlaylist(masterPath, variants); err != nil {
		_ = p.repo.MarkFailed(ctx, videoID, "failed to write master playlist")
		return err
	}

	relativeHLSPath := filepath.ToSlash(filepath.Join("hls", videoID.String(), "master.m3u8"))
	if err := p.repo.MarkReady(ctx, videoID, relativeHLSPath, meta.Duration, meta.Width, meta.Height, meta.Codec); err != nil {
		return fmt.Errorf("mark ready: %w", err)
	}

	log.Printf("video %s ready at %s", videoID, relativeHLSPath)
	return nil
}

type probeResult struct {
	Duration float64
	Width    int
	Height   int
	Codec    string
}

type ffprobeOutput struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
}

func probeVideo(ctx context.Context, ffprobePath, inputPath string) (probeResult, error) {
	cmd := exec.CommandContext(ctx, ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		inputPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return probeResult{}, fmt.Errorf("ffprobe failed: %w", err)
	}

	var parsed ffprobeOutput
	if err := json.Unmarshal(output, &parsed); err != nil {
		return probeResult{}, fmt.Errorf("parse ffprobe output: %w", err)
	}

	var result probeResult
	for _, stream := range parsed.Streams {
		if stream.CodecType == "video" && stream.Width > 0 && stream.Height > 0 {
			result.Width = stream.Width
			result.Height = stream.Height
			result.Codec = stream.CodecName
			break
		}
	}

	if result.Height == 0 {
		return probeResult{}, fmt.Errorf("no video stream found")
	}

	if parsed.Format.Duration != "" {
		duration, err := strconv.ParseFloat(parsed.Format.Duration, 64)
		if err == nil {
			result.Duration = duration
		}
	}

	return result, nil
}

func targetHeights(sourceHeight int) []int {
	candidates := []int{1080, 720, 480}
	var heights []int
	for _, height := range candidates {
		if sourceHeight >= height {
			heights = append(heights, height)
		}
	}
	if len(heights) == 0 {
		heights = []int{sourceHeight}
	}
	return heights
}

type variantInfo struct {
	Label      string
	Width      int
	Height     int
	Bandwidth  int
}

func scaledWidth(sourceWidth, sourceHeight, targetHeight int) int {
	if sourceHeight == 0 {
		return sourceWidth
	}
	width := int(float64(sourceWidth) * float64(targetHeight) / float64(sourceHeight))
	if width%2 != 0 {
		width--
	}
	if width < 2 {
		width = 2
	}
	return width
}

func bandwidthForHeight(height int) int {
	switch {
	case height >= 1080:
		return 5_000_000
	case height >= 720:
		return 2_800_000
	default:
		return 1_200_000
	}
}

func transcodeVariant(ctx context.Context, ffmpegPath, inputPath, outputDir string, height int) error {
	segmentPattern := filepath.Join(outputDir, "segment_%03d.ts")
	playlistPath := filepath.Join(outputDir, "index.m3u8")

	args := []string{
		"-y",
		"-i", inputPath,
		"-vf", fmt.Sprintf("scale=-2:%d", height),
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		"-ac", "2",
		"-hls_time", "4",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", segmentPattern,
		playlistPath,
	}

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed for %dp: %w", height, err)
	}

	return nil
}

func writeMasterPlaylist(path string, variants []variantInfo) error {
	var builder strings.Builder
	builder.WriteString("#EXTM3U\n")
	builder.WriteString("#EXT-X-VERSION:3\n")

	for _, variant := range variants {
		builder.WriteString(fmt.Sprintf(
			"#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d\n",
			variant.Bandwidth,
			variant.Width,
			variant.Height,
		))
		builder.WriteString(fmt.Sprintf("%s/index.m3u8\n", variant.Label))
	}

	return os.WriteFile(path, []byte(builder.String()), 0o644)
}
