.PHONY: help dev build deploy clean backup

# Colors for output
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
RESET  := $(shell tput -Txterm sgr0)

help: ## Show this help message
	@echo '$(GREEN)Available targets:$(RESET)'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}'

install-tools: ## Install development tools
	@echo "$(GREEN)Installing development tools...$(RESET)"
	go install github.com/air-verse/air@v1.62.0
	@echo "$(GREEN)Tools installed!$(RESET)"

dev-%: ## Run specific app in development mode (e.g., make dev-grocery-list)
	@if [ ! -d "apps/$*" ]; then \
		echo "$(YELLOW)App 'apps/$*' not found$(RESET)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Starting $* in development mode...$(RESET)"
	@cd apps/$* && air || go run main.go

build: ## Build all applications
	@echo "$(GREEN)Building all applications...$(RESET)"
	@docker-compose -f deploy/docker-compose.yml build

up: ## Start all services locally
	@echo "$(GREEN)Starting all services...$(RESET)"
	@docker-compose -f deploy/docker-compose.yml up -d

down: ## Stop all services
	@echo "$(GREEN)Stopping all services...$(RESET)"
	@docker-compose -f deploy/docker-compose.yml down

logs: ## Show logs from all services
	@docker-compose -f deploy/docker-compose.yml logs -f

deploy: ## Deploy to production server
	@echo "$(GREEN)Deploying to production...$(RESET)"
	@./scripts/deploy.sh

backup: ## Backup all SQLite databases
	@echo "$(GREEN)Backing up databases...$(RESET)"
	@./scripts/backup.sh

clean: ## Clean build artifacts and temporary files
	@echo "$(GREEN)Cleaning...$(RESET)"
	@find . -type f -name "*.db" -not -path "*/backups/*" -delete
	@find . -type f -name "*.log" -delete
	@find . -type d -name "tmp" -exec rm -rf {} + 2>/dev/null || true
	@docker-compose -f deploy/docker-compose.yml down -v
	@echo "$(GREEN)Clean complete!$(RESET)"

new-app: ## Create a new app (usage: make new-app NAME=my-app PORT=3005)
	@if [ -z "$(NAME)" ] || [ -z "$(PORT)" ]; then \
		echo "$(YELLOW)Usage: make new-app NAME=my-app PORT=3005$(RESET)"; \
		exit 1; \
	fi
	@./scripts/new-app.sh $(NAME) $(PORT)

test: ## Run tests for all apps
	@echo "$(GREEN)Running tests...$(RESET)"
	@for app in apps/*; do \
		if [ -d "$$app" ]; then \
			echo "$(YELLOW)Testing $$(basename $$app)...$(RESET)"; \
			cd $$app && go test ./... || exit 1; \
			cd ../..; \
		fi \
	done
	@echo "$(GREEN)All tests passed!$(RESET)"
