.PHONY: help build run dev test clean deps migrate-up migrate-down docker-build docker-run

# Variables
APP_NAME=rest-api-notes
BINARY_DIR=bin
CMD_DIR=cmd/server
MIGRATIONS_DIR=migrations

build: ## Build the application
	go build -o $(BINARY_DIR)/$(APP_NAME) $(CMD_DIR)/server.go

run: build ## Build and run the application
	./$(BINARY_DIR)/$(APP_NAME)

dev: ## Run in development mode with hot reload
	air