.PHONY: db-up db-down api frontend dev

db-up:
	docker compose up -d

db-down:
	docker compose down

api:
	cd services/api && go run ./cmd/api

frontend:
	cd frontend && npm run dev

dev:
	@echo "Start PostgreSQL: make db-up"
	@echo "Start API:        make api"
	@echo "Start frontend:   make frontend"
