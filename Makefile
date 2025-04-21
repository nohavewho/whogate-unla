# Registry configurations
DOCKER_REGISTRY ?= docker.io
GHCR_REGISTRY ?= ghcr.io
ALI_REGISTRY ?= registry.cn-hangzhou.aliyuncs.com

# Project configurations
PROJECT_NAME ?= mcp-gateway
IMAGE_TAG ?= $(shell cat pkg/version/VERSION)

# Service configurations
SERVICES = apiserver mcp-gateway mock-user-svc

# Build flags
LDFLAGS = -X main.version=$(VERSION)

# Registry targets
.PHONY: docker ghcr ali

# Build all services
.PHONY: build
build:
	@for service in $(SERVICES); do \
		docker build -t $(PROJECT_NAME)-$$service:$(IMAGE_TAG) \
			-f deploy/docker/multi/$$service/Dockerfile .; \
		docker tag $(PROJECT_NAME)-$$service:$(IMAGE_TAG) $(PROJECT_NAME)-$$service:latest; \
	done
	docker build -t $(PROJECT_NAME)-allinone:$(IMAGE_TAG) \
		-f deploy/docker/allinone/Dockerfile .
	docker tag $(PROJECT_NAME)-allinone:$(IMAGE_TAG) $(PROJECT_NAME)-allinone:latest

# Build multi-container version
.PHONY: build-multi
build-multi:
	@for service in $(SERVICES); do \
		docker build -t $(PROJECT_NAME)-$$service:$(IMAGE_TAG) \
			-f deploy/docker/multi/$$service/Dockerfile .; \
		docker tag $(PROJECT_NAME)-$$service:$(IMAGE_TAG) $(PROJECT_NAME)-$$service:latest; \
	done

# Build all-in-one version
.PHONY: build-allinone
build-allinone:
	docker build -t $(PROJECT_NAME)-allinone:$(IMAGE_TAG) \
		-f deploy/docker/allinone/Dockerfile .
	docker tag $(PROJECT_NAME)-allinone:$(IMAGE_TAG) $(PROJECT_NAME)-allinone:latest

# Run multi-container version
.PHONY: run-multi
run-multi:
	docker-compose -f deploy/docker/multi/docker-compose.yml up

# Run all-in-one version
.PHONY: run-allinone
run-allinone:
	docker-compose -f deploy/docker/allinone/docker-compose.yml up

# Push to Docker Hub
docker: build
	@for service in $(SERVICES); do \
		docker tag $(PROJECT_NAME)-$$service:$(IMAGE_TAG) \
			$(DOCKER_REGISTRY)/$(PROJECT_NAME)/$$service:$(IMAGE_TAG); \
		docker push $(DOCKER_REGISTRY)/$(PROJECT_NAME)/$$service:$(IMAGE_TAG); \
	done
	docker tag $(PROJECT_NAME)-allinone:$(IMAGE_TAG) \
		$(DOCKER_REGISTRY)/$(PROJECT_NAME)/allinone:$(IMAGE_TAG)
	docker push $(DOCKER_REGISTRY)/$(PROJECT_NAME)/allinone:$(IMAGE_TAG)

# Push to GitHub Container Registry
ghcr: build
	@for service in $(SERVICES); do \
		docker tag $(PROJECT_NAME)-$$service:$(IMAGE_TAG) \
			$(GHCR_REGISTRY)/$(PROJECT_NAME)/$$service:$(IMAGE_TAG); \
		docker push $(GHCR_REGISTRY)/$(PROJECT_NAME)/$$service:$(IMAGE_TAG); \
	done
	docker tag $(PROJECT_NAME)-allinone:$(IMAGE_TAG) \
		$(GHCR_REGISTRY)/$(PROJECT_NAME)/allinone:$(IMAGE_TAG)
	docker push $(GHCR_REGISTRY)/$(PROJECT_NAME)/allinone:$(IMAGE_TAG)

# Push to Alibaba Cloud Container Registry
ali: build
	@for service in $(SERVICES); do \
		docker tag $(PROJECT_NAME)-$$service:$(IMAGE_TAG) \
			$(ALI_REGISTRY)/$(PROJECT_NAME)/$$service:$(IMAGE_TAG); \
		docker push $(ALI_REGISTRY)/$(PROJECT_NAME)/$$service:$(IMAGE_TAG); \
	done
	docker tag $(PROJECT_NAME)-allinone:$(IMAGE_TAG) \
		$(ALI_REGISTRY)/$(PROJECT_NAME)/allinone:$(IMAGE_TAG)
	docker push $(ALI_REGISTRY)/$(PROJECT_NAME)/allinone:$(IMAGE_TAG)

# Clean up local images
.PHONY: clean
clean:
	@for service in $(SERVICES); do \
		docker rmi $(PROJECT_NAME)-$$service:$(IMAGE_TAG) || true; \
		docker rmi $(DOCKER_REGISTRY)/$(PROJECT_NAME)/$$service:$(IMAGE_TAG) || true; \
		docker rmi $(GHCR_REGISTRY)/$(PROJECT_NAME)/$$service:$(IMAGE_TAG) || true; \
		docker rmi $(ALI_REGISTRY)/$(PROJECT_NAME)/$$service:$(IMAGE_TAG) || true; \
	done
	docker rmi $(PROJECT_NAME)-allinone:$(IMAGE_TAG) || true
	docker rmi $(DOCKER_REGISTRY)/$(PROJECT_NAME)/allinone:$(IMAGE_TAG) || true
	docker rmi $(GHCR_REGISTRY)/$(PROJECT_NAME)/allinone:$(IMAGE_TAG) || true
	docker rmi $(ALI_REGISTRY)/$(PROJECT_NAME)/allinone:$(IMAGE_TAG) || true 