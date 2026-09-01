# Worker service

Consumes `video.jobs` from Kafka and runs FFmpeg transcoding in a separate process from the API.

```bash
export PATH="/opt/homebrew/bin:$PATH"
cd services/worker
go run ./cmd/worker
```

The worker joins consumer group `streamforge-workers` by default. Stop it to let uploads queue in Kafka, then restart to drain the backlog.
