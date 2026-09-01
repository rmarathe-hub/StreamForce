# Worker service

Runs FFmpeg transcoding in a separate process from the API.

The worker polls PostgreSQL for `UPLOADED` videos, claims them with `FOR UPDATE SKIP LOCKED`, and writes HLS output to shared storage.

```bash
export PATH="/opt/homebrew/bin:$PATH"
cd services/worker
go run ./cmd/worker
```

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | local postgres on `15433` | PostgreSQL connection string |
| `STORAGE_PATH` | `../../storage` | Shared storage root (must match API) |
| `FFMPEG_PATH` | `ffmpeg` | FFmpeg binary |
| `FFPROBE_PATH` | `ffprobe` | ffprobe binary |
| `WORKER_ID` | hostname | Identifier in structured logs |
| `WORKER_POLL_SECONDS` | `2` | Poll interval when idle |
