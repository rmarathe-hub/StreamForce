# StreamForge

Distributed video-processing and streaming platform (portfolio project).

Upload videos via the API, transcode them into adaptive HLS in a separate worker process, and play streams in the browser.

## Stack

- **Frontend:** Next.js 15, TypeScript, Tailwind CSS, hls.js
- **API:** Go, chi router, pgx
- **Worker:** Go, FFmpeg, ffprobe
- **Database:** PostgreSQL 16
- **Storage:** Local filesystem (`storage/uploads/`, `storage/hls/`)

## Prerequisites

- Go 1.22+
- Node.js 20+
- Docker Desktop
- FFmpeg (`brew install ffmpeg`)

## Quick start

### 1. Start PostgreSQL

```bash
docker compose up -d
```

### 2. Start the API

```bash
cd services/api
go run ./cmd/api
```

The API runs at `http://localhost:8081` and applies migrations on startup.

### 3. Start the worker

In a second terminal:

```bash
export PATH="/opt/homebrew/bin:$PATH"
cd services/worker
go run ./cmd/worker
```

The worker polls PostgreSQL for `UPLOADED` videos and runs FFmpeg independently of the API.

### 4. Start the frontend

In a third terminal:

```bash
cd frontend
npm run dev
```

Open `http://localhost:3000`, upload an MP4, and open the video detail page to watch the HLS stream once status is `READY`.

> **Note:** Defaults use port `8081` for the API and `15433` for PostgreSQL to avoid common local conflicts. Override with `PORT`, `DATABASE_URL`, and `NEXT_PUBLIC_API_URL` if needed.

## Architecture (Phase 3)

```
Frontend
   ↓
Go API  →  saves upload + marks UPLOADED
   ↓
PostgreSQL  ←→  Go Worker  →  FFmpeg  →  HLS files
   ↑
Go API serves /media/*
```

The API and worker are separate processes sharing PostgreSQL and the `storage/` directory. If the worker is stopped, uploads still succeed and remain `UPLOADED` until the worker restarts.

## Processing flow

```
Upload → UPLOADED → (worker claims job) → PROCESSING → FFmpeg/ffprobe → READY
```

HLS output:

```
storage/hls/{videoId}/
├── 1080p/index.m3u8 + segments
├── 720p/index.m3u8 + segments
├── 480p/index.m3u8 + segments
└── master.m3u8
```

## API endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/api/videos` | List all videos |
| `POST` | `/api/videos` | Upload video (`multipart/form-data`, field: `file`) |
| `GET` | `/api/videos/{id}` | Get video by ID |
| `GET` | `/media/*` | Serve uploaded files and HLS output |

## Environment variables

### API (`services/api`)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8081` | HTTP port |
| `DATABASE_URL` | `postgres://streamforge:streamforge@localhost:15433/streamforge?sslmode=disable` | PostgreSQL connection string |
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
| `WORKER_POLL_SECONDS` | `2` | Idle poll interval |

### Frontend (`frontend`)

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8081` | StreamForge API base URL |

## Project structure

```
streamforge/
├── frontend/
├── shared/             # models, repository, processor, database
├── services/
│   ├── api/            # HTTP API only
│   └── worker/         # FFmpeg worker
├── migrations/
├── storage/
└── docker-compose.yml
```

## Next steps

- Kafka job queue (replace DB polling)
- Redis progress, WebSockets
- Docker Compose, Kubernetes, k6 benchmarks
