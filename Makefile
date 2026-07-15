APP_NAME = golf-api
DOCKER_IMAGE = $(APP_NAME)
DOCKER_CONTAINER = $(APP_NAME)-container

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

# --- Dev mode using docker-compose (hot reload if using Air) ---
dev:
    docker compose up --build

# --- Tear down dev environment ---
dev-down:
    docker compose down

# --- Rebuild everything cleanly ---
rebuild: stop clean build run

.PHONY: test test-unit test-integration test-all

test-unit:
	go test ./...

test-integration:
	go test -tags=integration ./...

test-all: test-unit test-integration
