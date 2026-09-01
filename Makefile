.PHONY: db-up db-down api worker frontend dev

db-up:
	docker compose up -d

db-down:
	docker compose down

api:
	cd services/api && go run ./cmd/api

worker:
	cd services/worker && go run ./cmd/worker

frontend:
	cd frontend && npm run dev

dev:
	@echo "Start infrastructure: make db-up"
	@echo "Start API:            make api"
	@echo "Start worker:         make worker"
	@echo "Start frontend:       make frontend"
