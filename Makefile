BINARY := hashcards
.PHONY: frontend-deps
frontend-deps:
	cd frontend && pnpm install

.PHONY: frontend-build
frontend-build:
	cd frontend && pnpm run build


kill-ports:
	@lsof -ti:3000 | xargs -r kill -9 2>/dev/null || true
	@lsof -ti:3001 | xargs -r kill -9 2>/dev/null || true

server: kill-ports
	cd frontend && pnpm watch &
	air &


init: build kill-ports
	#./hashcards migrate up --dir=pb_data
	./$(BINARY) superuser upsert admin@mail.internal password --dir=pb_data



.PHONY: backend
backend:
	./$(BINARY) serve --config=config.toml

frontend-dev:
	cd frontend && pnpm dev


# ----------
.PHONY: frontend
frontend: kill-ports
	@echo "Starting backend with air..."
	@air & \
	BACKEND_PID=$$!; \
	echo "Starting frontend dev server..."; \
	cd frontend && pnpm run dev; \
	echo "Stopping backend (PID: $$BACKEND_PID)..."; \
	kill $$BACKEND_PID 2>/dev/null || true


test:
	go test ./...


# -----------


.PHONY: build-container
build-container:
	docker compose -f compose.yaml up --build --force-recreate


# ─────────────────────────────────────────
#  Docker / deploy
# ─────────────────────────────────────────
.PHONY: build-image
build-image: ## Build Docker image
	docker build -t registry.internal/go-hashcards:latest .

# ─────────────────────────────────────────
.PHONY: build-image-no-cache
build-image-no-cache: ## Build Docker image
	docker build --no-cache -t registry.internal/go-hashcards:latest .

.PHONY: push-image
push-image: ## Push Docker image
	docker push registry.internal/go-hashcards:latest

.PHONY: deploy
deploy: build-image push-image ## (*) Deploy stack via Komodo
	docker exec -it komodo km x -y destroy-stack hashcards
	docker exec -it komodo km x -y pull-stack   hashcards
	docker exec -it komodo km x -y deploy-stack hashcards

