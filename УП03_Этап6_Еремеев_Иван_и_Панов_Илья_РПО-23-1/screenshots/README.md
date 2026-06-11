# Скриншоты этапа 6

Все изображения сгенерированы из реальных выводов команд и данных GitHub API
(сырые логи лежат рядом в `../reports/`). Живые страницы для проверки:

- Issue: https://github.com/Ilpaka/LandMark_App/issues/7
- Pull Request: https://github.com/Ilpaka/LandMark_App/pull/8
- Actions: https://github.com/Ilpaka/LandMark_App/actions
- Релизы/теги: https://github.com/Ilpaka/LandMark_App/releases

| Файл | Что видно |
|---|---|
| 01_github_issue_created.png | Issue #7: описание, шаги, ожидаемый/фактический результат |
| 02_branch_created.png | Ветка feat/sixth_stage и коммиты этапа |
| 03_bug_reproduced.png | curl: text/plain «404 page not found»; падающие тесты-репродукции |
| 04_logs_diagnostics.png | Лог старта сервиса (маршруты gin), вывод о причине |
| 05_fix_commit.png | Коммит фикса: 19 файлов (9 сервисов + 18 тестов) |
| 06_tests_passed_locally.png | gofmt + go vet + go test по всем 10 модулям — зелёные |
| 07_github_actions_success.png | Успешный workflow run «CI» на PR этапа |
| 08_pull_request.png | PR #8: описание, чек-лист проверки, Closes #7 |
| 09_changelog_release_notes.png | CHANGELOG.md и RELEASE_NOTES.md (версия 0.3.1) |
| 10_release_tag_or_archive.png | Тег v0.3.1, релизный архив workflow «Release» |
