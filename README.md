# StreamForge

Distributed video-processing and streaming platform (portfolio project).

**Phase 1** delivers a working upload flow: Next.js frontend → Go API → PostgreSQL → local file storage.

## Stack

- **Frontend:** Next.js 15, TypeScript, Tailwind CSS
- **API:** Go, chi router, pgx
- **Database:** PostgreSQL 16
- **Storage:** Local filesystem (`storage/uploads/`)

## Prerequisites

- Go 1.22+
- Node.js 20+
- Docker Desktop

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

Open `http://localhost:3000`.

> **Note:** Defaults use port `8081` for the API and `15433` for PostgreSQL to avoid common local conflicts (e.g. other Postgres or Spark on `8080`). Override with `PORT`, `DATABASE_URL`, and `NEXT_PUBLIC_API_URL` if needed.

## API endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/api/videos` | List all videos |
| `POST` | `/api/videos` | Upload video (`multipart/form-data`, field: `file`) |
| `GET` | `/api/videos/{id}` | Get video by ID |

## Environment variables

### API (`services/api`)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8081` | HTTP port |
| `DATABASE_URL` | `postgres://streamforge:streamforge@localhost:15433/streamforge?sslmode=disable` | PostgreSQL connection string |
| `STORAGE_PATH` | `../../storage` | Local video storage root |
| `MIGRATIONS_PATH` | `../../migrations` | SQL migrations directory |
| `MAX_UPLOAD_MB` | `500` | Max upload size in MB |

### Frontend (`frontend`)

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8081` | StreamForge API base URL |

## Project structure

```
streamforge/
├── frontend/           # Next.js app
├── services/
│   └── api/            # Go REST API
├── migrations/         # PostgreSQL migrations
├── storage/            # Local uploads (gitignored)
├── sample-videos/
├── docs/
├── infra/
├── load-tests/
└── docker-compose.yml
```

## Phase 1 complete when

- Drag an MP4 into the browser
- File is stored on disk and recorded in PostgreSQL
- Refresh `/videos` and the upload still appears

## Next phases

- **Phase 2:** FFmpeg + HLS transcoding and playback
- **Phase 3+:** Separate worker, Kafka, Redis, WebSockets, Kubernetes, k6
