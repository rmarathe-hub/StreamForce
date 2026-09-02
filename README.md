# StreamForge

Distributed video-processing and streaming platform (portfolio project).

Upload videos via the API, queue transcoding jobs in Kafka, process them in a separate worker, and play adaptive HLS in the browser.

## Stack

- **Frontend:** Next.js 15, TypeScript, Tailwind CSS, hls.js
- **API:** Go, chi router, pgx, Kafka producer
- **Worker:** Go, Kafka consumer, FFmpeg, ffprobe
- **Database:** PostgreSQL 16
- **Messaging:** Apache Kafka (KRaft, local Docker)
- **Cache:** Redis 7 (live transcoding progress)
- **Storage:** Local filesystem (`storage/uploads/`, `storage/hls/`)

## Prerequisites

- Go 1.22+
- Node.js 20+
- Docker Desktop
- FFmpeg (`brew install ffmpeg`)

## Quick start

### 1. Start infrastructure

```bash
docker compose up -d
```

Starts PostgreSQL (`15433`), Kafka (`29092`), and Redis (`6379`).

### 2. Start the API

```bash
cd services/api
go run ./cmd/api
```

The API runs at `http://localhost:8081`, applies migrations, and ensures the `video.jobs` Kafka topic exists.

### 3. Start the worker

In a second terminal:

```bash
export PATH="/opt/homebrew/bin:$PATH"
cd services/worker
go run ./cmd/worker
```

The worker consumes from Kafka topic `video.jobs` and runs FFmpeg.

### 4. Start the frontend

In a third terminal:

```bash
cd frontend
npm run dev
```

Open `http://localhost:3000`, upload an MP4, and watch status move `QUEUED` → `PROCESSING` → `READY` with a live progress bar during transcoding.

## Architecture

```
Frontend
   ↓
Go API  →  save upload + QUEUED in Postgres
        →  publish video.jobs to Kafka
   ↓
Kafka topic: video.jobs
   ↓
Go Worker  →  consume job  →  FFmpeg  →  HLS files
           →  publish progress % to Redis
   ↓
PostgreSQL + Redis + shared storage
   ↑
Go API serves /media/* and reads progress from Redis
```

**Phase 4 demo:** stop the worker, upload videos (they stay `QUEUED` in Kafka), then start the worker and watch them process.

**Phase 5 demo:** run multiple workers in the same consumer group (`streamforge-workers`). Kafka distributes partitions across workers so different videos transcode in parallel (up to 3 workers with the default 3-partition topic).

## Multiple workers (Phase 5)

Start up to **3 workers** in separate terminals (one consumer per Kafka partition):

```bash
export PATH="/opt/homebrew/bin:$PATH"

# Terminal A
make worker-1

# Terminal B
make worker-2

# Terminal C
make worker-3
```

Each worker gets a distinct `WORKER_ID` (`worker-1`, `worker-2`, …). Upload several videos at once and check worker logs — you should see different `worker_id` values handling different jobs.

Jobs are claimed atomically in Postgres (`claimed_by`, `claimed_at`) so two workers never process the same video. If a worker dies mid-transcode, the claim expires after 10 minutes and another worker can reclaim the job.

## Live progress (Phase 6)

While a video is `PROCESSING`, the worker publishes transcoding progress (0–100%) to Redis. The API enriches `GET /api/videos/{id}` with `progress_percent`, and the frontend shows a live progress bar.

Redis key format: `streamforge:video:{video_id}:progress` (TTL 24h, deleted when processing completes).

## Real-time updates (Phase 7)

Status and progress changes are published to Redis pub/sub (`streamforge:video:events`). The API forwards them over WebSocket:

```
ws://localhost:8081/api/ws/videos/{id}
```

The video detail page connects on load and receives live updates (no polling). Workers publish events on claim, progress, and completion; the API publishes on upload (`QUEUED`).

## Portfolio UI (Phase 8)

- **Thumbnails** — FFmpeg extracts a poster frame during transcoding (`storage/thumbnails/{id}.jpg`)
- **Video cards** — grid layout with thumbnails, status badges, and duration on the library page
- **System stats** — `GET /api/stats` returns video counts by status and active workers; shown on the home page

## Processing flow

```
Upload → QUEUED → (Kafka) → PROCESSING → FFmpeg/ffprobe → READY
```

## Kafka message (`video.jobs`)

```json
{
  "eventId": "uuid",
  "videoId": "uuid",
  "sourcePath": "uploads/....mp4",
  "attempt": 1
}
```

## API endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/api/stats` | System stats (counts by status, active workers) |
| `GET` | `/api/videos` | List all videos |
| `POST` | `/api/videos` | Upload video |
| `GET` | `/api/videos/{id}` | Get video by ID (includes `progress_percent` when processing) |
| `GET` | `/api/ws/videos/{id}` | WebSocket stream of status/progress updates |
| `GET` | `/media/*` | Serve uploads and HLS output |

## Environment variables

### Shared infrastructure

| Variable | Default | Description |
|----------|---------|-------------|
| `KAFKA_BROKERS` | `localhost:29092` | Comma-separated broker list |
| `KAFKA_TOPIC` | `video.jobs` | Video processing topic |
| `REDIS_URL` | `redis://localhost:6379/0` | Redis connection URL |

### API (`services/api`)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8081` | HTTP port |
| `DATABASE_URL` | postgres on `15433` | PostgreSQL connection string |
| `STORAGE_PATH` | `../../storage` | Local video storage root |
| `MIGRATIONS_PATH` | `../../migrations` | SQL migrations directory |
| `MAX_UPLOAD_MB` | `500` | Max upload size in MB |

### Worker (`services/worker`)

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | same as API | PostgreSQL connection string |
| `STORAGE_PATH` | `../../storage` | Must match API storage path |
| `FFMPEG_PATH` | `ffmpeg` | FFmpeg binary |
| `FFPROBE_PATH` | `ffprobe` | ffprobe binary |
| `WORKER_ID` | hostname | Log identifier |
| `KAFKA_CONSUMER_GROUP` | `streamforge-workers` | Kafka consumer group |

### Frontend (`frontend`)

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8081` | StreamForge API base URL |

## Project structure

```
streamforge/
├── frontend/
├── shared/             # models, repository, processor, kafka, redis
├── services/
│   ├── api/            # HTTP API + Kafka producer
│   └── worker/         # Kafka consumer + FFmpeg
├── migrations/
├── storage/
└── docker-compose.yml
```

## Next steps

- Docker Compose for full app stack, Kubernetes, k6 benchmarks
