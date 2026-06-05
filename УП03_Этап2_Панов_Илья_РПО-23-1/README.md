# УП.03 · Этап 2. Проведение инсталляции ПО компьютерных систем

Проект **LandMark (Wanderlog)** — мобильное приложение (Flutter) + backend
(10 Go-микросервисов в Docker). На этом этапе настроены единые команды запуска,
проверки и форматирования.

- **Репозиторий:** https://github.com/Ilpaka/LandMark_App
- **Короткий отчёт:** [`docs/01_Короткий_отчет_по_этапу_2.docx`](docs/01_Короткий_отчет_по_этапу_2.docx)
- **Файлы в репозитории:** Makefile, `scripts/*.bat`, `docker-compose.yml`,
  `.dockerignore`, `.env.example`, `.editorconfig`, `.golangci.yml`, `INSTALL.md`

## Как запустить (кратко)

```bash
make setup        # установка зависимостей (Flutter + Go)
make docker-up    # поднять backend в Docker
make run-app      # запустить мобильный клиент
make check        # go vet + flutter analyze
make format       # gofmt + dart format
```

На Windows — те же действия через `scripts\*.bat`.

## Скриншоты проверки

### 1. Клонирование проекта в новую папку
![Чистый клон](screenshots/01_clean_clone.png)

### 2. Установка зависимостей (setup)
![setup](screenshots/02_setup_bat_success.png)

### 3. Команды Makefile / проверка `make check`
![make check](screenshots/03_make_help_or_make_check.png)

### 4. Успешная сборка Docker-образов
![docker build](screenshots/04_docker_build_success.png)

### 5. Запущенные сервисы (`docker compose up --build`)
![docker compose up](screenshots/05_docker_compose_up.png)

### 6. Программа открыта в браузере (Grafana)
![Программа в браузере](screenshots/06_grafana_opened_in_browser.png)

### 7. Линтер: «No issues found»
![Линтер](screenshots/07_linter_success.png)

### 8. Форматтер (`make format`)
![Форматтер](screenshots/08_formatter_success.png)

### 9. Коммит и push в ветку
![git commit](screenshots/09_git_commit.png)
