APP_NAME = golf-api
DOCKER_IMAGE = $(APP_NAME)
DOCKER_CONTAINER = $(APP_NAME)-container

.PHONY: build run stop clean logs dev dev-down dev-logs rebuild \
        run-local test test-unit test-integration test-all

# --- Local development (no Docker) ---
run-local:
	go run ./cmd/api

# --- Build the production Docker image ---
build:
	docker build -t $(DOCKER_IMAGE) .

# --- Run the container (production mode) ---
run:
	docker run --rm -p 8080:8080 --name $(DOCKER_CONTAINER) $(DOCKER_IMAGE)

# --- Stop the running container ---
stop:
	docker stop $(DOCKER_CONTAINER) || true

# --- Remove the container if it exists ---
clean:
	docker rm $(DOCKER_CONTAINER) || true

# --- View logs from the running container ---
logs:
	docker logs -f $(DOCKER_CONTAINER)

# --- Dev mode using docker-compose (foreground, streams logs) ---
dev:
	docker compose up --build

# --- Dev mode, detached ---
dev-detached:
	docker compose up --build -d

# --- Follow logs for the api service in detached dev mode ---
dev-logs:
	docker compose logs -f api

# --- Tear down dev environment ---
dev-down:
	docker compose down

# --- Rebuild everything cleanly ---
rebuild: stop clean build run

# --- Tests ---
test-unit:
	go test ./...

test-integration:
	go test -tags=integration ./...

test-all: test-unit test-integration