# GNUmakefile - Terrafactor Build & Deployment

# Configure shell path
SHELL := /bin/bash

# Application Configuration
BINARY_NAME := terrafactor
APP_NAME := terrafactor

# Versioning
VERSION := $(shell cat API_VERSION 2>/dev/null || echo "0.0.2")
DOCKER_TAG := $(subst +,-,$(VERSION))

# Build Configuration
SRC_DIR := .
BUILD_DIR := ./build
ARTIFACTS_DIR := ./artifacts
GO_MAIN := main.go

# Exclude specific directories and/or file patterns
EXCLUDE_DIR := ./tests
EXCLUDE_PATTERN := *.back.go

# Find command adjusted to exclude the specified directories and patterns
SOURCES := $(shell find $(SRC_DIR) -name '*.go' ! -path "$(EXCLUDE_DIR)/*" ! -name "$(EXCLUDE_PATTERN)")

# Container Registry Configuration (override via environment)
# Format: <REGISTRY>/<PROJECT>/<REPO>/<IMAGE>:<TAG>
REGISTRY ?= us-west1-docker.pkg.dev
REGISTRY_PROJECT ?= beh-dev-auto
REGISTRY_PROJECT_PROD ?= beh-prod-cent-infra
REGISTRY_REPO ?= central-infrastructure-registry
IMAGE_ORG ?= cloudputation
IMAGE_NAME := $(APP_NAME)
FULL_IMAGE := $(REGISTRY)/$(REGISTRY_PROJECT)/$(REGISTRY_REPO)/$(IMAGE_NAME):$(DOCKER_TAG)

# Fallback for Docker Hub / simple registries (when not using GCP Artifact Registry)
SIMPLE_IMAGE := $(IMAGE_ORG)/$(IMAGE_NAME):$(DOCKER_TAG)

# GKE Configuration (for gke-deploy target)
GKE_PROJECT ?= $(REGISTRY_PROJECT)
GKE_REGION ?= us-west1
GKE_CLUSTER ?= default

# Build flags
CGO_ENABLED := 0
GOOS := linux
GOARCH := amd64

# Local development container
LOCAL_CONTAINER_NAME := $(APP_NAME)
LOCAL_PORT ?= 8080

# Phony targets
.PHONY: all dev build clean docker-build docker-push docker-local docker-dev docker-prod gke-deploy gke-deploy-local gke-deploy-dev gke-deploy-prod local-deploy local-restart help

# Default target
all: build docker-build docker-push
	@echo "✓ Complete build pipeline finished"

# Local dev build via helpers/build.sh
dev:
	@bash helpers/build.sh

# Build the Go binary
build: $(SOURCES)
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@echo "Downloading Go dependencies..."
	@GO111MODULE=on go mod tidy
	@GO111MODULE=on go mod download
	@echo "Compiling binary..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		GO111MODULE=on go build -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)
	@echo "✓ Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Build the Docker image
docker-build: build
	@echo "Building Docker image..."
	@echo "  Image: $(IMAGE_NAME):$(DOCKER_TAG)"
	@docker build \
		--platform linux/amd64 \
		--build-arg PRODUCT_VERSION=$(VERSION) \
		-t $(IMAGE_NAME):$(DOCKER_TAG) \
		.
	@echo "✓ Docker image built: $(IMAGE_NAME):$(DOCKER_TAG)"

# Push the Docker image to the registry
docker-push: docker-build
	@echo "Pushing to container registry..."
	@echo "  Registry: $(REGISTRY)"
	@echo "  Full image: $(FULL_IMAGE)"
	@if [[ "$(REGISTRY)" == *"pkg.dev"* ]]; then \
		echo "Configuring registry authentication (GCP Artifact Registry)..."; \
		gcloud auth configure-docker $(REGISTRY) --quiet; \
		echo "Tagging image for registry..."; \
		docker tag $(IMAGE_NAME):$(DOCKER_TAG) $(FULL_IMAGE); \
		echo "Pushing image..."; \
		docker push $(FULL_IMAGE); \
		echo "✓ Image pushed: $(FULL_IMAGE)"; \
	else \
		echo "Tagging image for registry..."; \
		docker tag $(IMAGE_NAME):$(DOCKER_TAG) $(SIMPLE_IMAGE); \
		echo "Pushing image..."; \
		docker push $(SIMPLE_IMAGE); \
		echo "✓ Image pushed: $(SIMPLE_IMAGE)"; \
	fi

# Deploy to GKE cluster - all environments (build once, docker build 3x)
gke-deploy: build
	@echo "Building and pushing all environments..."
	@$(MAKE) docker-local
	@$(MAKE) docker-dev
	@$(MAKE) docker-prod
	@echo "✓ All environments deployed"

# Individual env targets WITH full build (for standalone use)
gke-deploy-local: build docker-local
gke-deploy-dev: build docker-dev
gke-deploy-prod: build docker-prod

# Docker-only targets (no app rebuild - assumes prior build)
docker-local:
	@echo "Building Docker image for local..."
	@if [ -f helpers/gke-deploy.sh ]; then \
		REGISTRY=$(REGISTRY) \
			REGISTRY_PROJECT=$(REGISTRY_PROJECT) \
			REGISTRY_REPO=$(REGISTRY_REPO) \
			IMAGE_NAME=$(IMAGE_NAME) \
			VERSION=$(DOCKER_TAG)-local \
			GKE_PROJECT=$(GKE_PROJECT) \
			GKE_REGION=$(GKE_REGION) \
			bash helpers/gke-deploy.sh --env local; \
	else \
		echo "Warning: helpers/gke-deploy.sh not found, skipping GKE deployment"; \
		docker build --platform linux/amd64 --build-arg PRODUCT_VERSION=$(VERSION)-local -t $(IMAGE_NAME):$(DOCKER_TAG)-local .; \
	fi
	@echo "✓ Local docker build complete"

docker-dev:
	@echo "Building Docker image for dev..."
	@if [ -f helpers/gke-deploy.sh ]; then \
		REGISTRY=$(REGISTRY) \
			REGISTRY_PROJECT=$(REGISTRY_PROJECT) \
			REGISTRY_REPO=$(REGISTRY_REPO) \
			IMAGE_NAME=$(IMAGE_NAME) \
			VERSION=$(DOCKER_TAG)-dev \
			GKE_PROJECT=$(GKE_PROJECT) \
			GKE_REGION=$(GKE_REGION) \
			bash helpers/gke-deploy.sh --env dev; \
	else \
		echo "Warning: helpers/gke-deploy.sh not found, skipping GKE deployment"; \
		docker build --platform linux/amd64 --build-arg PRODUCT_VERSION=$(VERSION)-dev -t $(IMAGE_NAME):$(DOCKER_TAG)-dev .; \
	fi
	@echo "✓ Dev docker build complete"

docker-prod:
	@echo "Building Docker image for prod..."
	@if [ -f helpers/gke-deploy.sh ]; then \
		REGISTRY=$(REGISTRY) \
			REGISTRY_PROJECT=$(REGISTRY_PROJECT_PROD) \
			REGISTRY_REPO=$(REGISTRY_REPO) \
			IMAGE_NAME=$(IMAGE_NAME) \
			VERSION=$(DOCKER_TAG)-prod \
			GKE_PROJECT=$(REGISTRY_PROJECT_PROD) \
			GKE_REGION=$(GKE_REGION) \
			bash helpers/gke-deploy.sh --env prod; \
	else \
		echo "Warning: helpers/gke-deploy.sh not found, skipping GKE deployment"; \
		docker build --platform linux/amd64 --build-arg PRODUCT_VERSION=$(VERSION)-prod -t $(IMAGE_NAME):$(DOCKER_TAG)-prod .; \
	fi
	@echo "✓ Prod docker build complete"

# Build and run locally in Docker
local-deploy: docker-build
	@echo "Stopping existing container (if running)..."
	@docker rm -f $(LOCAL_CONTAINER_NAME) 2>/dev/null || true
	@echo "Starting local container..."
	@docker run -d \
		--name $(LOCAL_CONTAINER_NAME) \
		-p $(LOCAL_PORT):8080 \
		--add-host=host.docker.internal:host-gateway \
		$(IMAGE_NAME):$(DOCKER_TAG)
	@echo "✓ Container running: $(LOCAL_CONTAINER_NAME)"
	@echo "  → http://localhost:$(LOCAL_PORT)"

# Quick restart: rebuild container only (skip go build)
local-restart:
	@echo "Rebuilding Docker image (skipping go build)..."
	@docker build \
		--platform linux/amd64 \
		--build-arg PRODUCT_VERSION=$(VERSION) \
		-t $(IMAGE_NAME):$(DOCKER_TAG) \
		.
	@echo "Stopping existing container..."
	@docker rm -f $(LOCAL_CONTAINER_NAME) 2>/dev/null || true
	@echo "Starting local container..."
	@docker run -d \
		--name $(LOCAL_CONTAINER_NAME) \
		-p $(LOCAL_PORT):8080 \
		--add-host=host.docker.internal:host-gateway \
		$(IMAGE_NAME):$(DOCKER_TAG)
	@echo "✓ Container restarted: $(LOCAL_CONTAINER_NAME)"
	@echo "  → http://localhost:$(LOCAL_PORT)"

# Clean up build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(ARTIFACTS_DIR)
	@echo "✓ Clean complete"

# Show help
help:
	@echo "$(APP_NAME) Build & Deployment"
	@echo ""
	@echo "Variables (override with environment or make VAR=value):"
	@echo "  VERSION          Current: $(VERSION)"
	@echo "  REGISTRY         Current: $(REGISTRY)"
	@echo "  REGISTRY_PROJECT Current: $(REGISTRY_PROJECT)"
	@echo "  REGISTRY_REPO    Current: $(REGISTRY_REPO)"
	@echo "  IMAGE_ORG        Current: $(IMAGE_ORG)"
	@echo "  IMAGE_NAME       Current: $(IMAGE_NAME)"
	@echo "  FULL_IMAGE       Current: $(FULL_IMAGE)"
	@echo ""
	@echo "Targets:"
	@echo "  make dev              - Local dev build via helpers/build.sh (native arch)"
	@echo "  make build            - Build $(BINARY_NAME) binary ($(GOOS)/$(GOARCH))"
	@echo "  make docker-build     - Build Docker image"
	@echo "  make docker-push      - Push to container registry"
	@echo "  make all              - Full pipeline (build → docker-build → docker-push)"
	@echo ""
	@echo "  GKE Deploy (full pipeline: app + docker):"
	@echo "  make gke-deploy       - Deploy all envs to GKE (local, dev, prod)"
	@echo "  make gke-deploy-local - Deploy local env to GKE"
	@echo "  make gke-deploy-dev   - Deploy dev env to GKE"
	@echo "  make gke-deploy-prod  - Deploy prod env to GKE"
	@echo ""
	@echo "  Docker-only (assumes app already built):"
	@echo "  make docker-local     - Build & push local Docker image only"
	@echo "  make docker-dev       - Build & push dev Docker image only"
	@echo "  make docker-prod      - Build & push prod Docker image only"
	@echo ""
	@echo "  make local-deploy     - Full build and run locally in Docker"
	@echo "  make local-restart    - Quick restart (rebuild container only)"
	@echo "  make clean            - Remove build artifacts"
	@echo "  make help             - Show this help"
	@echo ""
	@echo "Registry Configuration:"
	@echo "  GCP Artifact Registry (default):"
	@echo "    REGISTRY=us-west1-docker.pkg.dev"
	@echo "    Uses REGISTRY_PROJECT and REGISTRY_REPO"
	@echo ""
	@echo "  Docker Hub:"
	@echo "    REGISTRY=docker.io IMAGE_ORG=myorg"
	@echo "    Format: docker.io/myorg/$(IMAGE_NAME):$(DOCKER_TAG)"
	@echo ""
	@echo "  AWS ECR:"
	@echo "    REGISTRY=123456789.dkr.ecr.us-east-1.amazonaws.com IMAGE_ORG=myrepo"
	@echo ""
	@echo "Examples:"
	@echo "  make build                                    	# Standard build"
	@echo "  make REGISTRY=docker.io IMAGE_ORG=myorg push  	# Push to Docker Hub"
	@echo "  make VERSION=1.2.3 all                        	# Override version"
	@echo "  make gke-deploy                               	# Deploy to GKE"
	@echo "  make LOCAL_PORT=3000 local-deploy             	# Run on different port"
