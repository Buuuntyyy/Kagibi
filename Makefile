# Kagibi Makefile
# Makefile for building, testing, and deploying Kagibi

.PHONY: help build push deploy clean test all

# Configuration
REGISTRY ?= docker.io
NAMESPACE ?= kagibi
VERSION ?= latest
KUBECONFIG ?= ~/.kube/config

# Colors
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(GREEN)Kagibi Makefile$(NC)"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

all: build push deploy ## Build, push, and deploy everything

# Docker build targets
build: build-backend build-frontend build-website ## Build all Docker images

build-backend: ## Build backend Docker image
	@echo "$(YELLOW)Building backend...$(NC)"
	docker build -t $(REGISTRY)/$(NAMESPACE)/backend:$(VERSION) \
	             -t $(REGISTRY)/$(NAMESPACE)/backend:latest \
	             -f backend/Dockerfile ./backend
	@echo "$(GREEN)✅ Backend built$(NC)"

build-frontend: ## Build frontend Docker image
	@echo "$(YELLOW)Building frontend...$(NC)"
	docker build -t $(REGISTRY)/$(NAMESPACE)/frontend:$(VERSION) \
	             -t $(REGISTRY)/$(NAMESPACE)/frontend:latest \
	             -f frontend/Dockerfile ./frontend
	@echo "$(GREEN)✅ Frontend built$(NC)"

build-website: ## Build website Docker image
	@echo "$(YELLOW)Building website...$(NC)"
	docker build -t $(REGISTRY)/$(NAMESPACE)/website:$(VERSION) \
	             -t $(REGISTRY)/$(NAMESPACE)/website:latest \
	             -f k8s/website/Dockerfile ./website
	@echo "$(GREEN)✅ Website built$(NC)"

# Docker push targets
push: push-backend push-frontend push-website ## Push all Docker images

push-backend: ## Push backend Docker image
	@echo "$(YELLOW)Pushing backend...$(NC)"
	docker push $(REGISTRY)/$(NAMESPACE)/backend:$(VERSION)
	docker push $(REGISTRY)/$(NAMESPACE)/backend:latest
	@echo "$(GREEN)✅ Backend pushed$(NC)"

push-frontend: ## Push frontend Docker image
	@echo "$(YELLOW)Pushing frontend...$(NC)"
	docker push $(REGISTRY)/$(NAMESPACE)/frontend:$(VERSION)
	docker push $(REGISTRY)/$(NAMESPACE)/frontend:latest
	@echo "$(GREEN)✅ Frontend pushed$(NC)"

push-website: ## Push website Docker image
	@echo "$(YELLOW)Pushing website...$(NC)"
	docker push $(REGISTRY)/$(NAMESPACE)/website:$(VERSION)
	docker push $(REGISTRY)/$(NAMESPACE)/website:latest
	@echo "$(GREEN)✅ Website pushed$(NC)"

# Kubernetes deployment targets
deploy: ## Deploy to Kubernetes
	@echo "$(YELLOW)Deploying to Kubernetes...$(NC)"
	cd k8s && bash deploy.sh
	@echo "$(GREEN)✅ Deployment complete$(NC)"

deploy-kustomize: ## Deploy using Kustomize
	@echo "$(YELLOW)Deploying with Kustomize...$(NC)"
	kubectl apply -k k8s/
	@echo "$(GREEN)✅ Deployment complete$(NC)"

# Status and monitoring
status: ## Show deployment status
	@echo "$(YELLOW)Checking deployment status...$(NC)"
	kubectl get pods -n kagibi
	kubectl get svc -n kagibi
	kubectl get ingress -n kagibi

logs-backend: ## Show backend logs
	kubectl logs -f deployment/backend -n kagibi

logs-frontend: ## Show frontend logs
	kubectl logs -f deployment/frontend -n kagibi

logs-website: ## Show website logs
	kubectl logs -f deployment/website -n kagibi

# Scale targets
scale-backend: ## Scale backend (usage: make scale-backend REPLICAS=5)
	kubectl scale deployment/backend --replicas=$(REPLICAS) -n kagibi

scale-frontend: ## Scale frontend (usage: make scale-frontend REPLICAS=5)
	kubectl scale deployment/frontend --replicas=$(REPLICAS) -n kagibi

scale-website: ## Scale website (usage: make scale-website REPLICAS=3)
	kubectl scale deployment/website --replicas=$(REPLICAS) -n kagibi

# Update targets
update-backend: build-backend push-backend ## Build, push, and update backend
	kubectl rollout restart deployment/backend -n kagibi
	kubectl rollout status deployment/backend -n kagibi

update-frontend: build-frontend push-frontend ## Build, push, and update frontend
	kubectl rollout restart deployment/frontend -n kagibi
	kubectl rollout status deployment/frontend -n kagibi

update-website: build-website push-website ## Build, push, and update website
	kubectl rollout restart deployment/website -n kagibi
	kubectl rollout status deployment/website -n kagibi

# Rollback targets
rollback-backend: ## Rollback backend deployment
	kubectl rollout undo deployment/backend -n kagibi

rollback-frontend: ## Rollback frontend deployment
	kubectl rollout undo deployment/frontend -n kagibi

rollback-website: ## Rollback website deployment
	kubectl rollout undo deployment/website -n kagibi

# Testing targets
test-backend: ## Run backend tests
	cd backend && go test ./...

test-frontend: ## Run frontend tests
	cd frontend && npm test

test: test-backend test-frontend ## Run all tests

# Clean targets
clean: ## Delete Kubernetes resources
	@echo "$(RED)Deleting Kagibi from Kubernetes...$(NC)"
	kubectl delete namespace kagibi

clean-images: ## Remove local Docker images
	@echo "$(YELLOW)Removing local images...$(NC)"
	docker rmi $(REGISTRY)/$(NAMESPACE)/backend:$(VERSION) || true
	docker rmi $(REGISTRY)/$(NAMESPACE)/frontend:$(VERSION) || true
	docker rmi $(REGISTRY)/$(NAMESPACE)/website:$(VERSION) || true
	@echo "$(GREEN)✅ Images removed$(NC)"

# Development targets
dev-backend: ## Run backend in development mode
	cd backend && go run main.go

dev-frontend: ## Run frontend in development mode
	cd frontend && npm run dev

# Backup targets
backup-db: ## Create database backup
	kubectl create job --from=cronjob/postgres-backup manual-backup-$(shell date +%Y%m%d-%H%M%S) -n kagibi

list-backups: ## List all backups
	kubectl exec -it postgres-0 -n kagibi -- ls -lh /backups

# Monitoring targets
metrics: ## Show pod metrics
	kubectl top pods -n kagibi
	kubectl top nodes

events: ## Show recent events
	kubectl get events -n kagibi --sort-by='.lastTimestamp'

describe-backend: ## Describe backend deployment
	kubectl describe deployment/backend -n kagibi

describe-frontend: ## Describe frontend deployment
	kubectl describe deployment/frontend -n kagibi

describe-website: ## Describe website deployment
	kubectl describe deployment/website -n kagibi

# Port forwarding targets
port-forward-backend: ## Port forward backend to localhost:8080
	kubectl port-forward svc/backend-service 8080:8080 -n kagibi

port-forward-frontend: ## Port forward frontend to localhost:3000
	kubectl port-forward svc/frontend-service 3000:80 -n kagibi

port-forward-db: ## Port forward PostgreSQL to localhost:5432
	kubectl port-forward svc/postgres-service 5432:5432 -n kagibi

# Info targets
info: ## Show configuration info
	@echo "$(GREEN)Configuration:$(NC)"
	@echo "  Registry: $(REGISTRY)"
	@echo "  Namespace: $(NAMESPACE)"
	@echo "  Version: $(VERSION)"
	@echo "  Kubeconfig: $(KUBECONFIG)"
	@echo ""
	@echo "$(GREEN)Images:$(NC)"
	@echo "  - $(REGISTRY)/$(NAMESPACE)/backend:$(VERSION)"
	@echo "  - $(REGISTRY)/$(NAMESPACE)/frontend:$(VERSION)"
	@echo "  - $(REGISTRY)/$(NAMESPACE)/website:$(VERSION)"
