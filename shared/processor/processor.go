package processor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/rmarathe-hub/StreamForce/shared/models"
	"github.com/rmarathe-hub/StreamForce/shared/repository"
)

const (
	progressSetupEnd     = 5
	progressTranscodeEnd = 95
	progressFinalize     = 98
)

type Processor struct {
	repo     *repository.VideoRepository
	cfg      Config
	baseDir  string
	progress ProgressReporter
}

func New(repo *repository.VideoRepository, cfg Config, progress ProgressReporter) *Processor {
	if progress == nil {
		progress = NoopProgressReporter()
	}
	return &Processor{
		repo:     repo,
		cfg:      cfg,
		baseDir:  cfg.StoragePath,
		progress: progress,
	}
}

func (p *Processor) Process(ctx context.Context, video models.Video) error {
	if video.Status == models.StatusReady {
		return nil
	}

	if video.Status != models.StatusProcessing {
		return fmt.Errorf("video %s is not processing", video.ID)
	}

	tracker := newProgressTracker(ctx, p.progress, video.ID)
	tracker.report(0)

	sourcePath := filepath.Join(p.baseDir, video.SourcePath)
	meta, err := probeVideo(ctx, p.cfg.FFprobePath, sourcePath)
	if err != nil {
		tracker.clear()
		_ = p.repo.MarkFailed(ctx, video.ID, err.Error())
		return err
	}

	tracker.report(progressSetupEnd)

	thumbnailPath, err := p.generateThumbnail(ctx, sourcePath, video.ID, meta.Duration)
	if err != nil {
		tracker.clear()
		_ = p.repo.MarkFailed(ctx, video.ID, err.Error())
		return err
	}

	outputDir := filepath.Join(p.baseDir, "hls", video.ID.String())
	if err := os.RemoveAll(outputDir); err != nil {
		tracker.clear()
		_ = p.repo.MarkFailed(ctx, video.ID, "failed to prepare output directory")
		return err
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		tracker.clear()
		_ = p.repo.MarkFailed(ctx, video.ID, "failed to create output directory")
		return err
	}

	heights := targetHeights(meta.Height)
	variants := make([]variantInfo, 0, len(heights))

	for index, height := range heights {
		label := fmt.Sprintf("%dp", height)
		variantDir := filepath.Join(outputDir, label)
		if err := os.MkdirAll(variantDir, 0o755); err != nil {
			tracker.clear()
			_ = p.repo.MarkFailed(ctx, video.ID, "failed to create variant directory")
			return err
		}

		variantIndex := index
		if err := transcodeVariant(
			ctx,
			p.cfg.FFmpegPath,
			sourcePath,
			variantDir,
			height,
			meta.Duration,
			func(variantPercent int) {
				tracker.report(overallProgress(variantIndex, len(heights), variantPercent))
			},
		); err != nil {
			tracker.clear()
			_ = p.repo.MarkFailed(ctx, video.ID, err.Error())
			return err
		}

		variants = append(variants, variantInfo{
			Label:     label,
			Width:     scaledWidth(meta.Width, meta.Height, height),
			Height:    height,
			Bandwidth: bandwidthForHeight(height),
		})
	}

	tracker.report(progressFinalize)

	masterPath := filepath.Join(outputDir, "master.m3u8")
	if err := writeMasterPlaylist(masterPath, variants); err != nil {
		tracker.clear()
		_ = p.repo.MarkFailed(ctx, video.ID, "failed to write master playlist")
		return err
	}

	relativeHLSPath := filepath.ToSlash(filepath.Join("hls", video.ID.String(), "master.m3u8"))
	if err := p.repo.MarkReady(
		ctx,
		video.ID,
		relativeHLSPath,
		thumbnailPath,
		meta.Duration,
		meta.Width,
		meta.Height,
		meta.Codec,
	); err != nil {
		tracker.clear()
		return fmt.Errorf("mark ready: %w", err)
	}

	tracker.report(100)
	tracker.clear()
	return nil
}

type progressTracker struct {
	ctx      context.Context
	reporter ProgressReporter
	videoID  uuid.UUID
	lastPct  int
}

func newProgressTracker(ctx context.Context, reporter ProgressReporter, videoID uuid.UUID) *progressTracker {
	return &progressTracker{ctx: ctx, reporter: reporter, videoID: videoID, lastPct: -1}
}

func (t *progressTracker) report(percent int) {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	if percent <= t.lastPct && percent < 100 {
		return
	}
	t.lastPct = percent
	_ = t.reporter.Set(t.ctx, t.videoID, percent)
}

func (t *progressTracker) clear() {
	_ = t.reporter.Delete(t.ctx, t.videoID)
}

func overallProgress(variantIndex, variantCount, variantPercent int) int {
	if variantCount <= 0 {
		return progressSetupEnd
	}

	transcodeRange := progressTranscodeEnd - progressSetupEnd
	weight := float64(transcodeRange) / float64(variantCount)
	base := progressSetupEnd + int(float64(variantIndex)*weight)
	add := int(float64(variantPercent) * weight / 100)
	total := base + add
	if total > progressTranscodeEnd {
		return progressTranscodeEnd
	}
	return total
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
	Label     string
	Width     int
	Height    int
	Bandwidth int
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

func transcodeVariant(
	ctx context.Context,
	ffmpegPath, inputPath, outputDir string,
	height int,
	duration float64,
	onProgress func(int),
) error {
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
	}

	if onProgress != nil {
		args = append(args, "-progress", "pipe:1", "-nostats")
	}

	args = append(args, playlistPath)

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	cmd.Stderr = os.Stderr

	if onProgress == nil || duration <= 0 {
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("ffmpeg failed for %dp: %w", height, err)
		}
		return nil
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg stdout pipe for %dp: %w", height, err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg start for %dp: %w", height, err)
	}

	scanner := bufio.NewScanner(stdout)
	lastReported := -1
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "out_time_ms=") {
			continue
		}

		value := strings.TrimPrefix(line, "out_time_ms=")
		if value == "N/A" {
			continue
		}

		micros, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			continue
		}

		percent := int(float64(micros) / (duration * 1_000_000) * 100)
		if percent > 100 {
			percent = 100
		}
		if percent > lastReported {
			lastReported = percent
			onProgress(percent)
		}
	}

	if err := scanner.Err(); err != nil {
		_ = cmd.Process.Kill()
		return fmt.Errorf("read ffmpeg progress for %dp: %w", height, err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg failed for %dp: %w", height, err)
	}

	onProgress(100)
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

func (p *Processor) generateThumbnail(
	ctx context.Context,
	sourcePath string,
	videoID uuid.UUID,
	duration float64,
) (string, error) {
	thumbDir := filepath.Join(p.baseDir, "thumbnails")
	if err := os.MkdirAll(thumbDir, 0o755); err != nil {
		return "", fmt.Errorf("create thumbnail directory: %w", err)
	}

	outputPath := filepath.Join(thumbDir, videoID.String()+".jpg")
	seek := 1.0
	if duration > 0 && duration < seek {
		seek = duration / 2
	}

	args := []string{
		"-y",
		"-ss", fmt.Sprintf("%.2f", seek),
		"-i", sourcePath,
		"-frames:v", "1",
		"-vf", "scale=640:-2",
		"-q:v", "3",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, p.cfg.FFmpegPath, args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("generate thumbnail: %w", err)
	}

	return filepath.ToSlash(filepath.Join("thumbnails", videoID.String()+".jpg")), nil
}
