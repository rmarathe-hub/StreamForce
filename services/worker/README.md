# Worker service

Consumes `video.jobs` from Kafka and runs FFmpeg transcoding in a separate process from the API.

```bash
export PATH="/opt/homebrew/bin:$PATH"
cd services/worker
go run ./cmd/worker
```

## Multiple workers

Run several workers in the same consumer group to process videos in parallel:

```bash
# From repo root
make worker-1   # terminal 1
make worker-2   # terminal 2
make worker-3   # terminal 3
```

Each worker uses a distinct `WORKER_ID` and joins consumer group `streamforge-workers` by default. Kafka assigns one partition per worker (up to 3 with the default topic).

Stop all workers to let uploads queue in Kafka, then restart to drain the backlog.
