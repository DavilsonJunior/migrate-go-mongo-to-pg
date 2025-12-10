# Makefile para o projeto Go Migration

.PHONY: help build clean test run-seed run-simple run-goroutines run-stream run-stream-goroutines run-break-memory docker-up docker-down

# Configurações
BINARY_DIR=bin
ENV_FILE=.env

# Cores para output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

help: ## Mostra esta ajuda
	@echo "$(BLUE)Go Migration - PostgreSQL ↔ MongoDB$(NC)"
	@echo ""
	@echo "$(YELLOW)Comandos disponíveis:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Configura o ambiente inicial
	@echo "$(BLUE)Configurando ambiente...$(NC)"
	@mkdir -p $(BINARY_DIR)
	@if [ ! -f $(ENV_FILE) ]; then cp .env.example $(ENV_FILE); echo "$(GREEN)Arquivo .env criado. Configure suas variáveis!$(NC)"; fi
	@go mod tidy
	@echo "$(GREEN)Setup completo!$(NC)"

build: setup ## Compila todos os executáveis
	@echo "$(BLUE)Compilando todos os executáveis...$(NC)"
	@go build -o $(BINARY_DIR)/seed_mongo ./seed_mongo
	@go build -o $(BINARY_DIR)/migrate_simple ./migrate_simple
	@go build -o $(BINARY_DIR)/migrate_goroutines ./migrate_goroutines_only
	@go build -o $(BINARY_DIR)/migrate_stream ./migrate_stream_only
	@go build -o $(BINARY_DIR)/migrate_stream_goroutines ./migrate_stream_goroutines
	@go build -o $(BINARY_DIR)/break_memory ./break_memory
	@echo "$(GREEN)Compilação concluída! Executáveis em $(BINARY_DIR)/$(NC)"

clean: ## Remove binários compilados
	@echo "$(YELLOW)Limpando binários...$(NC)"
	@rm -rf $(BINARY_DIR)
	@rm -f *_bin
	@echo "$(GREEN)Limpeza concluída!$(NC)"

test: ## Executa testes
	@echo "$(BLUE)Executando testes...$(NC)"
	@go test ./...

docker-up: ## Sobe os containers dos bancos
	@echo "$(BLUE)Subindo containers dos bancos...$(NC)"
	@docker-compose up -d
	@echo "$(GREEN)Containers iniciados!$(NC)"
	@echo "$(YELLOW)PostgreSQL: localhost:5440$(NC)"
	@echo "$(YELLOW)MongoDB: localhost:27017$(NC)"

docker-down: ## Para os containers dos bancos
	@echo "$(YELLOW)Parando containers...$(NC)"
	@docker-compose down
	@echo "$(GREEN)Containers parados!$(NC)"

docker-logs: ## Mostra logs dos containers
	@docker-compose logs -f

run-seed: ## Executa o seeder do MongoDB
	@echo "$(BLUE)Populando MongoDB com dados de teste...$(NC)"
	@go run ./seed_mongo

run-simple: ## Executa migração simples
	@echo "$(BLUE)Executando migração simples...$(NC)"
	@go run ./migrate_simple

run-goroutines: ## Executa migração com goroutines
	@echo "$(BLUE)Executando migração com goroutines...$(NC)"
	@go run ./migrate_goroutines_only

run-stream: ## Executa migração com streaming
	@echo "$(BLUE)Executando migração com streaming...$(NC)"
	@go run ./migrate_stream_only

run-stream-goroutines: ## Executa migração otimizada (streaming + goroutines)
	@echo "$(BLUE)Executando migração otimizada...$(NC)"
	@go run ./migrate_stream_goroutines

run-break-memory: ## Executa teste de limite de memória
	@echo "$(RED)⚠️  ATENÇÃO: Este teste pode consumir muita memória!$(NC)"
	@echo "$(YELLOW)Pressione Ctrl+C para interromper se necessário$(NC)"
	@sleep 3
	@go run ./break_memory

benchmark: docker-up run-seed ## Executa benchmark de todas as estratégias
	@echo "$(BLUE)Executando benchmark completo...$(NC)"
	@echo "\n$(YELLOW)=== Migração Simples ===== $(NC)"
	@time make run-simple
	@echo "\n$(YELLOW)=== Migração com Goroutines ===== $(NC)"
	@time make run-goroutines
	@echo "\n$(YELLOW)=== Migração com Streaming ===== $(NC)"
	@time make run-stream
	@echo "\n$(YELLOW)=== Migração Otimizada ===== $(NC)"
	@time make run-stream-goroutines
	@echo "\n$(GREEN)Benchmark completo!$(NC)"

dev-setup: ## Setup completo para desenvolvimento
	@echo "$(BLUE)Configurando ambiente de desenvolvimento...$(NC)"
	@make setup
	@make docker-up
	@sleep 5
	@make run-seed
	@echo "$(GREEN)Ambiente pronto para desenvolvimento!$(NC)"
	@echo "$(YELLOW)Teste com: make run-stream-goroutines$(NC)"

check-env: ## Verifica se as variáveis de ambiente estão configuradas
	@echo "$(BLUE)Verificando configuração...$(NC)"
	@if [ ! -f $(ENV_FILE) ]; then echo "$(RED)❌ Arquivo .env não encontrado!$(NC)"; exit 1; fi
	@echo "$(GREEN)✅ Arquivo .env encontrado$(NC)"
	@go run -c "migration-go/internal/config" || echo "$(YELLOW)⚠️  Execute 'make docker-up' antes de rodar as migrações$(NC)"

install-deps: ## Instala dependências Go
	@echo "$(BLUE)Instalando dependências...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)Dependências instaladas!$(NC)"

format: ## Formata o código Go
	@echo "$(BLUE)Formatando código...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)Código formatado!$(NC)"

lint: ## Executa linter no código
	@echo "$(BLUE)Executando linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint não encontrado. Instalando...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

all: clean build docker-up run-seed ## Executa setup completo