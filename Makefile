.PHONY: db-up db-down api worker worker-% workers frontend dev

db-up:
	docker compose up -d

db-down:
	docker compose down

api:
	cd services/api && go run ./cmd/api

worker:
	cd services/worker && go run ./cmd/worker

worker-%:
	WORKER_ID=worker-$* cd services/worker && go run ./cmd/worker

workers:
	@echo "Start multiple workers (one per terminal, up to 3 in parallel):"
	@echo "  export PATH=\"/opt/homebrew/bin:\$$PATH\""
	@echo "  make worker-1"
	@echo "  make worker-2"
	@echo "  make worker-3"

frontend:
	cd frontend && npm run dev

dev:
	@echo "Start infrastructure: make db-up"
	@echo "Start API:            make api"
	@echo "Start worker(s):      make worker   (or make worker-1, worker-2, worker-3)"
	@echo "Start frontend:       make frontend"
