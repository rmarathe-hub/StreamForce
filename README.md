# StreamForge

Distributed video-processing and streaming platform (portfolio project).

Upload videos, transcode them into adaptive HLS with FFmpeg, and play streams in the browser.

## Stack

- **Frontend:** Next.js 15, TypeScript, Tailwind CSS, hls.js
- **API:** Go, chi router, pgx
- **Processing:** FFmpeg, ffprobe
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

### 3. Start the frontend

```bash
cd frontend
npm install
npm run dev
```

Open `http://localhost:3000`, upload an MP4, and open the video detail page to watch the HLS stream once status is `READY`.

> **Note:** Defaults use port `8081` for the API and `15433` for PostgreSQL to avoid common local conflicts. Override with `PORT`, `DATABASE_URL`, and `NEXT_PUBLIC_API_URL` if needed.

## Processing flow

```
Upload → UPLOADED → PROCESSING → FFmpeg/ffprobe → READY
```

HLS output:

```
storage/hls/{videoId}/
├── 1080p/index.m3u8 + segments
├── 720p/index.m3u8 + segments
├── 480p/index.m3u8 + segments
└── master.m3u8
```

Renditions are generated without upscaling:

- 1080p source → 1080p / 720p / 480p
- 720p source → 720p / 480p
- 480p source → 480p

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
| `FFMPEG_PATH` | `ffmpeg` | FFmpeg binary |
| `FFPROBE_PATH` | `ffprobe` | ffprobe binary |

### Frontend (`frontend`)

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8081` | StreamForge API base URL |

## Project structure

```
streamforge/
├── frontend/           # Next.js app
├── services/
│   └── api/            # Go REST API + transcoding
├── migrations/         # PostgreSQL migrations
├── storage/            # Local uploads + HLS (gitignored)
├── sample-videos/
├── docs/
├── infra/
├── load-tests/
└── docker-compose.yml
```

## Next steps

- Separate API and worker service
- Kafka job queue, Redis progress, WebSockets
- Docker Compose, Kubernetes, k6 benchmarks
