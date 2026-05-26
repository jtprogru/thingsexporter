# thingsexporter MVP — Requirements

**Status:** Draft
**Author:** Mikhail Savin (через AI-ассистента в режиме spec-driven-dev)
**Date:** 2026-05-25

## Overview

`thingsexporter` — CLI-утилита на Go (модуль `github.com/jtprogru/thingsexporter`, бинарь `thingsexporter`), которая читает локальную SQLite-БД macOS-приложения Things 3 (`main.sqlite`) в режиме read-only и выгружает её содержимое в JSON или Markdown с настраиваемым составом контента. Утилита семантически воспроизводит поведение референсного Python-скрипта `/Users/jtprogru/Work/tmp/things3db/export.py` (даты Core Data → ISO 8601 UTC, packed dates → `YYYY-MM-DD`, enum-коды → имена, BLOB → hex, иерархия Areas → Projects → Tasks), добавляет к нему параметризацию форматов и пресетов состава и распространяется через GitHub Releases + Homebrew formula в `jtprogru/homebrew-tap`. Инфраструктура (Taskfile, golangci v2, goreleaser v2 с SBOM/cosign keyless, CI/release GitHub Actions, dependabot) повторяет соседний проект `jtprogru/todushka`. TUI намеренно отсутствует.

## Glossary

| Term | Definition | Code Artifact |
|------|------------|---------------|
| `ExportData` | Корневая структура выгрузки: `meta`, коллекции (`areas`/`tags`/`tasks`/`checklistItems`/`contacts`/`tombstones`), `links` (m2m-пары), `hierarchy` (срез Areas → Projects → Tasks). Содержимое определяется текущим пресетом `--include`. | `internal/things` |
| `IncludePreset` | Именованный пресет состава выгрузки: `all`, `tasks`, `tasks+tags`, `tasks+projects`. Определяет, какие коллекции и какие поля попадут в `ExportData`. | `internal/export/preset.go` |
| `Format` | Сериализатор `ExportData` в байты: `json` (default) или `markdown`. Реализует общий интерфейс `Writer{ Write(io.Writer, ExportData) error }`. | `internal/export` |
| `CoreDataTimestamp` | Вещественное число секунд от `2001-01-01 00:00:00 UTC`, в Things 3 хранится в полях `creationDate`, `userModificationDate`, `stopDate`, `lastReminderInteractionDate`, `repeaterMigrationDate`, `usedDate`, `deletionDate`. Конвертируется в ISO 8601 в UTC. | `internal/things/dates.go` |
| `PackedDate` | 32-битное целое с битовой раскладкой `(year<<16) \| (month<<12) \| (day<<7)`. В Things 3 хранится в полях `startDate`, `deadline`, `deadlineSuppressionDate`. Конвертируется в строку `YYYY-MM-DD`. | `internal/things/dates.go` |
| `DefaultMacOSDBPath` | Стандартный путь к `main.sqlite` для App Store-версии Things 3 на macOS: `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite`. | `internal/store/sqlite/discover.go` |
| `BlobValue` | Сериализация BLOB-поля: объект `{"__blob_hex__": "<hex>"}` (по умолчанию) или `null` (если BLOB пуст или включён флаг `--no-blobs`). | `internal/things/types.go` |
| `Trashed` | Признак нахождения сущности в корзине Things 3 (`TMTask.trashed = 1`). В MVP такие задачи попадают в коллекцию `tasks`, но исключаются из `hierarchy` — поведение совпадает с Python-скриптом. | `internal/things/types.go` |
| `HomebrewTap` | Внешний git-репозиторий `github.com/jtprogru/homebrew-tap`, в который GoReleaser публикует обновлённую formula при каждом релизе по тегу `v*`. Требует токена `HOMEBREW_TAP_GITHUB_TOKEN` (Classic PAT с правами `repo`) в secrets текущего репо. | `.goreleaser.yaml` |

## User Stories

- Как **пользователь Things 3**, я хочу одной командой выгрузить всю свою БД в JSON, чтобы передавать её во внешние инструменты (бэкап, отчёты, аналитика) без зависимости от Python.
- Как **пользователь Things 3**, я хочу получить Markdown-снимок текущих задач в иерархии Areas → Projects → Tasks, чтобы хранить его в персональной wiki (Obsidian/Logseq).
- Как **скриптовик/автоматизатор**, я хочу запускать `thingsexporter` из shell-пайпов (выгрузка в stdout, человекочитаемые ошибки в stderr, осмысленные exit-коды), чтобы интегрировать его с cron/launchd.
- Как **maintainer**, я хочу `brew install jtprogru/tap/thingsexporter` без сборки из исходников, чтобы получить подписанный, проверенный SBOM-ом бинарь.
- Как **разработчик `thingsexporter`**, я хочу, чтобы CI на каждом push в main прогонял `go vet`, `govulncheck`, `go test -race` и `goreleaser check`, чтобы регрессии и уязвимости отлавливались до релиза.

## Requirements

### Группа 1 — Чтение БД

**REQ-1.1** WHEN пользователь запускает `thingsexporter` и передаёт `--db <path>`, the system SHALL открывать SQLite-файл строго в read-only режиме (DSN `file:<path>?mode=ro`) и НЕ выполнять никаких write-операций над файлом, его WAL/SHM-журналами или содержащим каталогом.

**REQ-1.2** WHEN пользователь запускает `thingsexporter` БЕЗ `--db`, текущая ОС — macOS и файл существует по пути `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite`, the system SHALL использовать этот путь как источник БД и выводить в stderr строку `using DB: <resolved path>` (если не задан `--quiet`).

**REQ-1.3** WHEN пользователь запускает `thingsexporter` БЕЗ `--db` и автоматическое определение пути не дало результата (ОС не macOS, либо файл по `DefaultMacOSDBPath` отсутствует), the system SHALL завершаться с кодом 2, выводить в stderr сообщение вида `error: --db is required (no Things 3 database found at default path)` и НЕ создавать выходной файл.

**REQ-1.4** WHEN указанный (или авто-определённый) путь к БД не существует, недоступен на чтение или не является валидным SQLite-файлом, the system SHALL завершаться с кодом 2 и выводить в stderr оригинальную ошибку драйвера, префиксированную `error: `, и НЕ создавать выходной файл.

**REQ-1.5** WHEN БД успешно открыта, the system SHALL читать таблицы `TMArea`, `TMTag`, `TMTask`, `TMChecklistItem`, `TMContact`, `TMTombstone`, `Meta`, `TMTaskTag(tasks, tags)`, `TMAreaTag(areas, tags)` полностью, использовать в SQL двойные кавычки вокруг идентификаторов-резервированных слов (`"index"`, `"type"`, `"status"`, `"start"`) и не выполнять никаких других запросов.

**REQ-1.6** WHEN значение `databaseVersion` в таблице `Meta` отличается от поддерживаемого MVP списка (на момент релиза — `[26]`), the system SHALL выводить в stderr предупреждение `warning: unsupported Things 3 databaseVersion=<n>, output may be incomplete`, но продолжать выгрузку.

### Группа 2 — Конвертация типов

**REQ-2.1** WHEN поле содержит `CoreDataTimestamp`, the system SHALL конвертировать его в строку ISO 8601 в UTC (например, `2024-10-28T07:48:41.024293+00:00`) для каждого из полей: `creationDate`, `userModificationDate`, `stopDate`, `lastReminderInteractionDate`, `repeaterMigrationDate` (для задач и чек-листов), `usedDate` (для тегов), `deletionDate` (для тумбстонов).

**REQ-2.2** WHEN поле `CoreDataTimestamp` равно `NULL` или не парсится как число, the system SHALL подставлять `null` в соответствующее поле JSON-вывода.

**REQ-2.3** WHEN поле содержит `PackedDate` (`startDate`, `deadline`, `deadlineSuppressionDate`), the system SHALL декодировать его в строку `YYYY-MM-DD` и сохранять параллельно в дополнительных полях `startDateISO`, `deadlineISO`, `deadlineSuppressionDateISO` рядом с оригинальным числовым значением.

**REQ-2.4** WHEN `PackedDate` равен `NULL`, `0`, не парсится или декодированные `year`/`month`/`day` выходят за валидные диапазоны (`1970 ≤ year ≤ 2100`, `1 ≤ month ≤ 12`, `1 ≤ day ≤ 31`), the system SHALL подставлять `null` в соответствующее `*ISO`-поле.

**REQ-2.5** WHEN задача содержит поля-коды `type` или `status`, the system SHALL добавлять в JSON-вывод производные поля `typeName` (`{0: "todo", 1: "project", 2: "heading"}`) и `statusName` (`{0: "open", 2: "canceled", 3: "completed"}`); при неизвестном коде значение поля — `null`.

**REQ-2.6** WHEN чек-лист-айтем содержит поле-код `status`, the system SHALL добавлять в JSON-вывод производное поле `statusName` (`{0: "open", 3: "completed"}`); при неизвестном коде — `null`.

**REQ-2.7** WHEN поле является BLOB и флаг `--no-blobs` НЕ задан, the system SHALL сериализовать его как объект `{"__blob_hex__": "<lowercase-hex>"}`, если содержимое непустое, либо как `null`, если BLOB пуст.

**REQ-2.8** WHEN флаг `--no-blobs` задан, the system SHALL подставлять `null` в каждое BLOB-поле выходной структуры независимо от его исходного содержимого.

### Группа 3 — Обогащение задач и сборка `ExportData`

**REQ-3.1** WHEN формируется коллекция `tasks` (любой пресет, где задачи присутствуют), the system SHALL добавлять каждой задаче поля `areaTitle`, `projectTitle`, `headingTitle`, `contactName`, разрешая UUID-ссылки `area`/`project`/`heading`/`contact` через таблицы `TMArea`/`TMTask`/`TMContact`; при несуществующей ссылке — `null`.

**REQ-3.2** WHEN формируется коллекция `tasks` для пресета, включающего теги, the system SHALL добавлять каждой задаче поле `tags` — массив объектов `{uuid, title}`, построенный из `TMTaskTag`; при пустом списке — `[]`.

**REQ-3.3** WHEN формируется коллекция `tasks` для пресета `all`, the system SHALL добавлять каждой задаче поле `checklist` — массив чек-лист-айтемов, относящихся к этой задаче, отсортированный по `index` (ASC, `null` в конце); при отсутствии чек-листа — `[]`.

**REQ-3.4** WHEN формируется коллекция `areas` для пресета `all`, the system SHALL добавлять каждой области поле `tags` — массив объектов `{uuid, title}`, построенный из `TMAreaTag`; при пустом списке — `[]`.

**REQ-3.5** WHEN формируется поле `hierarchy` (только для пресета `all`), the system SHALL строить структуру `{areas: [{uuid, title, visible, index, tags, items: [{uuid, title, typeName, statusName}]}], inbox_or_orphan_tasks: [{uuid, title, typeName, statusName}]}` со следующими правилами:
- внешние области сортируются по `index` ASC (`null` в конце);
- `items[]` области содержит только задачи и проекты, у которых `trashed=0`, `project=NULL`, `heading=NULL` и `area=<uuid области>`, отсортированные по `index` ASC (`null` в конце);
- `inbox_or_orphan_tasks` содержит задачи и проекты с `trashed=0`, `project=NULL`, `heading=NULL`, `area=NULL`, отсортированные по `index` ASC.

**REQ-3.6** WHEN формируется корневой объект `ExportData`, the system SHALL включать поле `meta` со структурой `{source: <db path>, exportedAt: <ISO 8601 UTC текущий момент>, counts: {areas, tags, tasks, checklistItems, contacts, tombstones, taskTagLinks, areaTagLinks}, db_meta_rows: [{key, value}]}`, где `counts` отражает фактические длины соответствующих коллекций ПОСЛЕ применения пресета, а `db_meta_rows` — построчная копия таблицы `Meta`.

### Группа 4 — Форматы вывода

**REQ-4.1** WHEN `--format json` (или формат не задан, дефолт), the system SHALL сериализовать `ExportData` как UTF-8 JSON с отступом по умолчанию 2 пробела и БЕЗ ASCII-escape (поддерживая кириллицу и эмодзи в исходных данных).

**REQ-4.2** WHEN указан `--indent <N>`, the system SHALL использовать `N` пробелов как отступ JSON; при `N=0` сериализатор SHALL производить компактный JSON без переносов строк и отступов.

**REQ-4.3** WHEN `--format markdown`, the system SHALL сериализовать `ExportData` как Markdown по правилам:
- секция `# Inbox` в начале содержит `inbox_or_orphan_tasks` как `- [ ] title`/`- [x] title`/`- [- ] title`-список;
- секция `# Areas` далее: для каждой области заголовок `## <title>`, под ней список её корневых проектов и задач (как из `hierarchy.areas[].items`);
- каждый проект разворачивается в `### <title>` со своим списком задач из этого проекта (`TMTask.project = <uuid>`);
- маркер чекбокса: `[ ]` для open, `[x]` для completed, `[-]` для canceled;
- метаданные задачи рендерятся inline после title через два пробела: теги как `#tag` (только пресеты с тегами), deadline как `⏰ YYYY-MM-DD`;
- заметки (`notes`) задачи рендерятся как индентированный блок (4 пробела) под задачей, если не пусты;
- чек-лист (если пресет `all`) рендерится как вложенный список под задачей с такими же чекбоксами.

**REQ-4.4** WHEN формат — `markdown`, а пресет не содержит коллекций тегов (`tasks`, `tasks+projects`), the system SHALL пропускать inline-теги в выводе и НЕ добавлять для них пустые маркеры.

**REQ-4.5** WHEN `--format` имеет неизвестное значение, the system SHALL завершаться с кодом 2 и выводить в stderr `error: unknown format "<value>" (supported: json, markdown)`.

### Группа 5 — Пресеты состава (`--include`)

**REQ-5.1** WHEN `--include all` (или не задан, дефолт), the system SHALL включать в `ExportData` все коллекции (`areas`, `tags`, `tasks`, `checklistItems`, `contacts`, `tombstones`), полностью обогащённые задачи (REQ-3.1, 3.2, 3.3), обогащённые области (REQ-3.4), блок `links` (`taskTag`, `areaTag`) и блок `hierarchy` (REQ-3.5).

**REQ-5.2** WHEN `--include tasks`, the system SHALL включать в `ExportData` только коллекцию `tasks` с конвертацией дат и enum (Группа 2) БЕЗ полей `tags`, `checklist`, `areaTitle`, `projectTitle`, `headingTitle`, `contactName`, и НЕ включать остальные коллекции, `links` и `hierarchy`. Поле `meta` SHALL присутствовать; `counts` SHALL содержать только ключ `tasks`.

**REQ-5.3** WHEN `--include tasks+tags`, the system SHALL включать в `ExportData` коллекции `tasks` (с полем `tags` по REQ-3.2, без `checklist`, без resolve-полей) и `tags` (с полем `parentTitle` и конвертированным `usedDate`). `meta.counts` SHALL содержать ключи `tasks` и `tags`.

**REQ-5.4** WHEN `--include tasks+projects`, the system SHALL включать в `ExportData` коллекции `areas`, `tasks` (с полями `areaTitle`, `projectTitle`, `headingTitle` по REQ-3.1, без `tags` и без `checklist`). `meta.counts` SHALL содержать ключи `areas` и `tasks`.

**REQ-5.5** WHEN `--include` имеет неизвестное значение, the system SHALL завершаться с кодом 2 и выводить в stderr `error: unknown include preset "<value>" (supported: all, tasks, tasks+tags, tasks+projects)`.

### Группа 6 — CLI-поверхность и вывод

**REQ-6.1** WHEN пользователь запускает `thingsexporter` БЕЗ подкоманды, the system SHALL выполнять эквивалент `thingsexporter export` с дефолтными значениями (`--format json`, `--include all`, `--out -`, `--indent 2`).

**REQ-6.2** WHEN пользователь запускает `thingsexporter export [флаги]`, the system SHALL применять флаги: `--db <path>`, `--out <path|->` (default `-` = stdout), `--format <json|markdown>` (default `json`), `--include <preset>` (default `all`), `--indent <int>` (default `2`), `--no-blobs` (bool, default false), `--quiet` (bool, default false).

**REQ-6.3** WHEN `--out` равен `-` или не задан, the system SHALL писать сериализованный вывод в `os.Stdout`; WHEN `--out <path>` указывает файл, the system SHALL создавать или перезаписывать файл по этому пути с правами `0644` и завершаться с кодом 2 при ошибке записи.

**REQ-6.4** WHEN экспорт успешно завершён и `--quiet` НЕ задан, the system SHALL выводить в stderr краткий отчёт вида:
```
OK -> <output-path>
  areas: <N>
  tags: <N>
  tasks: <N>
  …
```
с теми ключами, что присутствуют в `meta.counts`. WHEN `--quiet` задан — отчёт SHALL подавляться.

**REQ-6.5** WHEN пользователь запускает `thingsexporter inspect [--db <path>]`, the system SHALL открывать БД (REQ-1.1—1.4), читать только счётчики (`SELECT COUNT(*)` по каждой таблице из REQ-1.5) и выводить их в stdout как JSON `{"path": "<path>", "databaseVersion": <n|null>, "counts": {...}}`. WHEN `--quiet` задан, дополнительные сообщения SHALL не выводиться.

**REQ-6.6** WHEN пользователь запускает `thingsexporter version`, the system SHALL выводить в stdout строку формата:
```
thingsexporter <Version>
  commit:    <Commit|->
  built:     <Date|->
  built by:  <BuiltBy|->
  go:        <runtime.Version()>
  platform:  <GOOS>/<GOARCH>
```
где `Version/Commit/Date/BuiltBy` — переменные пакета `internal/version`, инжектируемые `-ldflags` при сборке.

**REQ-6.7** WHEN пользователь запускает `thingsexporter completion <bash|zsh|fish|powershell>`, the system SHALL генерировать и выводить в stdout соответствующий shell-completion-скрипт (стандартный механизм cobra).

**REQ-6.8** WHEN пользователь запускает `thingsexporter --help` (или любую подкоманду с `--help`), the system SHALL выводить в stdout справку cobra и завершаться с кодом 0.

**REQ-6.9** WHEN команда успешно завершилась, the system SHALL возвращать exit-код `0`; WHEN возникает ошибка валидации флагов/значений или ошибка ввода-вывода — `2`; WHEN возникает паника или нерасклассифицированная ошибка — `1`.

### Группа 7 — Инфраструктура сборки, релизов и публикации

**REQ-7.1** WHEN запускается `task build` (или `go build`), the system SHALL собирать бинарь из `./cmd/thingsexporter` с `CGO_ENABLED=0` и БЕЗ зависимостей от C-библиотек; SQLite SHALL читаться через `modernc.org/sqlite`.

**REQ-7.2** WHEN запускается `task test`, the system SHALL выполнять `go test ./...`; WHEN запускается `task test-race` — `go test -race ./...`.

**REQ-7.3** WHEN запускается `task lint`, the system SHALL выполнять `golangci-lint run` с конфигурацией v2, эквивалентной `todushka/.golangci.yml` (линтеры `govet`, `staticcheck`, `errcheck`, `gosec`, `gocritic`, `revive`, `unused`, `ineffassign`; форматтеры `gofmt`, `goimports`; `gosec` исключает `G104`; в тестах отключены `gosec` и `errcheck`).

**REQ-7.4** WHEN на ветку `main` происходит push или открывается pull request, GitHub Actions workflow `ci` SHALL запускать на `ubuntu-latest`: `go vet ./...`, `govulncheck ./...`, `go test -race -coverprofile=cover.out ./...`, `goreleaser check`, `goreleaser build --snapshot --clean --single-target` (использовать pinned-by-SHA action versions, эквивалентные `todushka/.github/workflows/ci.yml`).

**REQ-7.5** WHEN в репозиторий пушится git-tag вида `v*`, GitHub Actions workflow `release` SHALL запускать на `ubuntu-latest`: `govulncheck`, `go test ./...`, установку `cosign` и `syft`, `goreleaser release --clean` с permissions `contents: write` и `id-token: write`; передавать `GITHUB_TOKEN` и `HOMEBREW_TAP_GITHUB_TOKEN` из repo secrets.

**REQ-7.6** WHEN GoReleaser выполняет `release`, the system SHALL собирать матрицу `linux/darwin × amd64/arm64`, упаковывать каждую сборку в `tar.gz` с именем `thingsexporter_<Version>_<Os>_<Arch>` (Os с заглавной буквы, `amd64`→`x86_64`, `386`→`i386`), генерировать `checksums.txt` (SHA-256), генерировать SBOM через syft (`*.sbom.json`), подписывать checksum через cosign keyless (`*.bundle`), публиковать GitHub Release с автоматическим changelog (Conventional Commits: `feat`/`fix`, исключая `docs:`/`test:`/`chore:`/merge-commits).

**REQ-7.7** WHEN GoReleaser выполняет `release`, the system SHALL обновлять Homebrew **formula** (не cask) в репозитории `github.com/jtprogru/homebrew-tap@main` с описанием `Export Things 3 SQLite database to JSON or Markdown`, homepage `https://github.com/jtprogru/thingsexporter`, license `MIT`, commit-author `goreleaserbot <bot@goreleaser.com>`, commit-message-template `chore(thingsexporter): bring formula to {{ .Tag }}`.

**REQ-7.8** WHEN `HOMEBREW_TAP_GITHUB_TOKEN` отсутствует в repo secrets во время `release`, GoReleaser SHALL завершаться с ненулевым кодом, оставлять GitHub Release созданным БЕЗ обновления formula, и логировать причину; restart релиза после добавления токена SHALL приводить к корректной публикации formula.

**REQ-7.9** WHEN ldflags инжектируют `Version/Commit/Date/BuiltBy` в `github.com/jtprogru/thingsexporter/internal/version`, the system SHALL подставлять для них `{{.Version}}`/`{{.Commit}}`/`{{.Date}}`/`"goreleaser"` соответственно; локальная сборка БЕЗ ldflags SHALL давать `Version="dev"`, остальные поля — пустые строки, выводимые как `-` в `version`-команде.

**REQ-7.10** WHEN Dependabot обновляет зависимости, the system SHALL открывать weekly PR-ы для `gomod` (директория `/`) и `github-actions` (директория `/`) с лимитом 5 одновременных PR на каждую экосистему и префиксами коммитов `chore(deps)` / `chore(ci)` соответственно.

### Группа 8 — Тестирование

**REQ-8.1** WHEN запускается `task test`, the system SHALL покрывать табличными unit-тестами: `core_data_to_iso` (вкл. границы и `null`), `packed_date_to_iso` (вкл. невалидные года/месяцы/дни), маппинг `typeName`/`statusName` (вкл. неизвестные коды), сериализацию BLOB (пустой/непустой/`--no-blobs`), генерацию SQL для каждой целевой таблицы (проверка двойных кавычек вокруг резервированных слов).

**REQ-8.2** WHEN запускается `task test`, the system SHALL содержать минимум один integration-тест, который генерирует fixture-SQLite-БД (через `database/sql` в файл `t.TempDir()` со схемой Things 3 и контролируемым набором строк), запускает полный экспорт с `--include all --format json` и проверяет: значения `meta.counts`, корректность дат, корректность `areaTitle`/`projectTitle`, наличие задачи в нужной секции `hierarchy`, исключение trashed-задач из `hierarchy`, наличие `links.taskTag`.

**REQ-8.3** WHEN запускается `task test`, the system SHALL содержать integration-тест Markdown-формата на той же fixture-БД, проверяющий: наличие секций `# Inbox` / `# Areas`, заголовков `##`/`###` для каждой области/проекта, корректность маркеров `[ ]`/`[x]`/`[-]`, наличие тегов как `#tag` и deadline как `⏰ YYYY-MM-DD`.

**REQ-8.4** WHEN запускается `task test`, the system SHALL содержать тесты CLI-поверхности на семе `cli.Deps` (по аналогии с todushka): передача custom stdout/stderr/stdin/env, корректность exit-кодов (0/2 для валидных и невалидных аргументов), формат отчёта в stderr.

**REQ-8.5** WHEN в репозиторий коммитятся файлы, реальная пользовательская БД (`main.sqlite`, `main.sqlite-wal`, `main.sqlite-shm`) SHALL быть в `.gitignore`; fixture-БД для тестов SHALL генерироваться кодом в runtime, а не храниться в git.

## Topological Order

```
Группа 1 (чтение БД) → Группа 2 (конвертация типов) → Группа 3 (обогащение и сборка ExportData) → Группа 4 (форматы вывода) → Группа 5 (пресеты состава) → Группа 6 (CLI-поверхность) → Группа 7 (инфраструктура) → Группа 8 (тестирование)
Причина: каждая следующая группа потребляет результат предыдущей. Чтение БД даёт сырые строки → конвертация делает их типизированными → обогащение собирает ExportData → форматы решают, как его сериализовать → пресеты определяют, какой именно ExportData собирать → CLI оркестрирует флаги и подкоманды → инфраструктура упаковывает всё в релизный артефакт → тесты валидируют каждый уровень.

REQ-7.5 → REQ-7.7 → REQ-7.8 (release workflow должен запустить goreleaser, который сначала формирует GitHub Release, а затем — обновляет formula).
REQ-8.2 → REQ-8.3 (markdown-тест переиспользует fixture-генератор из JSON-теста).
REQ-1.6 (предупреждение про databaseVersion) независимо.
REQ-7.10 (dependabot) независимо.
```

## Conflict Priority

```
REQ-2.7 (BLOB как hex по умолчанию) vs REQ-2.8 (BLOB как null при --no-blobs).
Resolution: REQ-2.8 имеет приоритет, если флаг --no-blobs явно установлен; иначе действует REQ-2.7. Дефолт без флага — REQ-2.7, что соответствует поведению Python-скрипта.

REQ-1.2 (auto-discover) vs REQ-1.3 (ошибка без --db).
Resolution: REQ-1.2 проверяется первым. REQ-1.3 срабатывает только если REQ-1.2 не дала результата (ОС не macOS либо файл по DefaultMacOSDBPath отсутствует). Явный --db всегда отключает обе ветки.

REQ-6.1 (root = export с дефолтами) vs REQ-6.5/6.6/6.7/6.8 (подкоманды).
Resolution: cobra разрешает диспетчеризацию автоматически — наличие подкоманды или `--help` отключает REQ-6.1. Конфликт декларативный, разрешается фреймворком.
```

## Open Design Questions

| Question | Why It Matters | Impacted Requirements |
|----------|---------------|----------------------|
| Каков точный layout структур `ExportData`/`Task`/`Area`/`Tag`/`ChecklistItem` в Go (struct-теги, опциональные поля, `*string` vs `string` для nullable)? | Определяет API пакета `internal/things`, форму JSON-вывода и поведение `omitempty`. | REQ-2.x, REQ-3.x, REQ-4.1 |
| Как именно строить `meta.counts` для частичных пресетов: считать только включённые коллекции или всегда считать все, но в выводе оставлять только релевантные ключи? | Влияет на инвариант REQ-3.6 и на тесты REQ-8.2. | REQ-3.6, REQ-5.x |
| Использовать ли `database/sql` напрямую или поднимать тонкий wrapper (типа `sqlx`)? | Влияет на боли при чтении 39-колонной `TMTask` и тесты SQL-генерации. | REQ-1.5, REQ-8.1 |
| Какой stdlib-пакет JSON: `encoding/json` или `encoding/json/v2` (если доступен в go 1.26)? | Влияет на детерминизм порядка ключей и поддержку `indent=0`. | REQ-4.1, REQ-4.2 |
| Где именно располагать map-фабрику Format-ов и Include-пресетов — в одном пакете `internal/export` или разнесённых `internal/export/format`/`internal/export/preset`? | Архитектурное решение, влияет на тесты и точки расширения. | REQ-4.x, REQ-5.x |
| Какой именно тег чек-листа использовать в Markdown (вложенный список `  - [ ]` vs `- [ ]` под notes)? | Влияет на читабельность и Obsidian-совместимость. | REQ-4.3 |

Эти вопросы намеренно НЕ закрываются на стадии Requirements — это архитектурные решения для фазы Design.

## Verification Commands

| Action   | Command                                                | Source                                                     |
|----------|--------------------------------------------------------|------------------------------------------------------------|
| Test     | `task test`                                            | `Taskfile.yml` (повторяем `todushka/Taskfile.yml`)         |
| Test-race| `task test-race`                                       | `Taskfile.yml`                                             |
| Build    | `task build`                                           | `Taskfile.yml`                                             |
| Cross    | `task cross-compile`                                   | `Taskfile.yml`                                             |
| Lint     | `task lint` (`golangci-lint run`)                      | `Taskfile.yml` + `.golangci.yml`                           |
| Fmt      | `task fmt` (`go fmt ./... && goimports -w .`)          | `Taskfile.yml`                                             |
| Tidy     | `task tidy` (`go mod tidy`)                            | `Taskfile.yml`                                             |
| Release-check (local) | `goreleaser check && goreleaser build --snapshot --clean --single-target` | `.goreleaser.yaml` + CI workflow (`ci.yml`) |
| Vuln     | `govulncheck ./...`                                    | CI workflow (`ci.yml`, `release.yml`)                      |
