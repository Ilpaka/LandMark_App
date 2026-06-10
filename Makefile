# LandMark (Wanderlog) — единая карта команд проекта (УП.03, этап 2).
# Кроссплатформенные команды запуска, проверки и форматирования.
# Для Windows аналоги лежат в scripts/*.bat.

SHELL := /bin/bash
FLUTTER ?= flutter
APP_DIR := landmark_app
BACKEND := backend
GO_SERVICES := api-gateway auth-service favorites-service journal-service media-service moderation-service notifications-service places-service profile-service trips-service

.DEFAULT_GOAL := help
.PHONY: help setup run run-app check format test lint-strict \
        build release-check docker-build docker-up docker-down logs ps clean

help: ## Показать список команд
	@echo 'LandMark — команды проекта:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	  | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

setup: ## Установить зависимости (Flutter + Go)
	cd $(APP_DIR) && $(FLUTTER) pub get
	@for s in $(GO_SERVICES); do echo ">> go mod download: $$s"; \
	  (cd $(BACKEND)/$$s && go mod download) || exit 1; done

run: ## Запустить весь backend локально (Docker, foreground)
	docker compose up --build

run-app: ## Запустить мобильный клиент Flutter
	cd $(APP_DIR) && $(FLUTTER) run

check: ## Проверка качества: gofmt + go vet + flutter analyze
	@echo '>> gofmt (проверка форматирования Go)'
	@bad="$$(for s in $(GO_SERVICES); do gofmt -l $(BACKEND)/$$s; done)"; \
	  if [ -n "$$bad" ]; then echo 'Не отформатировано:'; echo "$$bad"; exit 1; fi
	@for s in $(GO_SERVICES); do echo ">> go vet: $$s"; \
	  (cd $(BACKEND)/$$s && go vet ./...) || exit 1; done
	cd $(APP_DIR) && $(FLUTTER) analyze

format: ## Отформатировать код (gofmt + dart format)
	@for s in $(GO_SERVICES); do echo ">> gofmt -w: $$s"; \
	  (cd $(BACKEND)/$$s && gofmt -w .); done
	cd $(APP_DIR) && dart format lib test

test: ## Тесты (Go -short + Flutter)
	@for s in $(GO_SERVICES); do echo ">> go test: $$s"; \
	  (cd $(BACKEND)/$$s && go test ./... -short) || exit 1; done
	cd $(APP_DIR) && $(FLUTTER) test

lint-strict: ## Глубокий линтер Go (golangci-lint, если установлен)
	@if command -v golangci-lint >/dev/null 2>&1; then \
	  for s in $(GO_SERVICES); do echo ">> golangci-lint: $$s"; \
	    (cd $(BACKEND)/$$s && golangci-lint run ./...) || exit 1; done; \
	else echo 'golangci-lint не установлен — пропускаю (опционально).'; fi

build: ## Собрать release-бинарники Go-сервисов в dist/
	@mkdir -p dist
	@for s in $(GO_SERVICES); do echo ">> go build: $$s"; \
	  (cd $(BACKEND)/$$s && CGO_ENABLED=0 go build -o ../../dist/$$s ./cmd/server) || exit 1; done
	@echo 'Бинарники собраны в dist/'

release-check: check test build ## Финальная проверка перед релизом (этап 6)
	@echo 'Проект готов к релизу.'

docker-build: ## Собрать docker-образы backend
	docker compose build

docker-up: ## Поднять backend в фоне (detached)
	docker compose up --build -d

docker-down: ## Остановить и удалить контейнеры
	docker compose down

logs: ## Логи контейнеров (follow)
	docker compose logs -f

ps: ## Статус контейнеров
	docker compose ps

clean: ## Очистить сборочный мусор Flutter
	cd $(APP_DIR) && $(FLUTTER) clean
