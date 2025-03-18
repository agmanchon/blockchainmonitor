# Makefile

# Define the path to the setup directory where docker-compose.yml is located
SETUP_DIR := ./setup

# Define the docker-compose command
DOCKER_COMPOSE := docker-compose -f $(SETUP_DIR)/docker-compose.yml
GO_RUN_CMD = go run

CUR_DIR := $(CURDIR)
MAIN_FILE := $(CUR_DIR)/main.go
CONFIG_PATH := $(CUR_DIR)/setup

DOCKER_COMPOSE := docker-compose -f $(CONFIG_PATH)/docker-compose.yml

run:
	@echo "Starting blockchain monitor"
	$(GO_RUN_CMD) $(MAIN_FILE) start --configPath $(CONFIG_PATH)
# Define the target for setting up the services using docker-compose
provision:
	@echo "Starting services with Docker Compose..."
	$(DOCKER_COMPOSE) up -d

# Add a command to stop the services (optional)
deprovision:
	@echo "Stopping services with Docker Compose..."
	$(DOCKER_COMPOSE) down -v

# Add a command to view logs (optional)
logs:
	@echo "Viewing logs with Docker Compose..."
	$(DOCKER_COMPOSE) logs -f
