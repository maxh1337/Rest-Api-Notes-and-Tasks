#!/bin/bash

APP_NAME="."

mkdir -p $APP_NAME/{bin,cmd/server,docker,docs,internal/{config,delivery/http/{handlers,middleware,routes},domain/{entities,repositories,services},infrastructure/{database},usecases},migrations,pkg,scripts,tmp}



echo "package config" > $APP_NAME/internal/config/config.go
echo "package handlers" > $APP_NAME/internal/delivery/http/handlers/health_handler.go
echo "package middleware" > $APP_NAME/internal/delivery/http/middleware/logging.go
echo "package routes" > $APP_NAME/internal/delivery/http/routes/router.go
echo "package entities" > $APP_NAME/internal/domain/entities/user.go
echo "package repositories" > $APP_NAME/internal/domain/repositories/user_repository.go
echo "package services" > $APP_NAME/internal/domain/services/user_service.go
echo "package usecases" > $APP_NAME/internal/usecases/user_usecase.go
echo "package database" > $APP_NAME/internal/infrastructure/database/postgres.go

# Пустые вспомогательные файлы
touch $APP_NAME/{Makefile,README.md,go.mod,go.sum,project-structure.txt}
touch $APP_NAME/init.sh
touch $APP_NAME/docs/api.md
touch $APP_NAME/migrations/.gitkeep
touch $APP_NAME/scripts/.gitkeep
touch $APP_NAME/tmp/.gitkeep
touch $APP_NAME/bin/.gitkeep
touch $APP_NAME/pkg/.gitkeep
touch $APP_NAME/docker/Dockerfile
touch $APP_NAME/docker/docker-compose.yml

echo "✅ Чистая структура '$APP_NAME' создана без кода — только packages."
