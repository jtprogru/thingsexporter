# Exploration: thingsexporter MVP

## Intent

Реализовать на Go CLI-утилиту `thingsexporter`, которая читает локальную базу приложения Things 3 (`main.sqlite`) и выгружает данные в файл (или stdout) в одном из выбранных форматов и с настраиваемым составом контента. По функционалу базового JSON-экспорта должна полностью повторять референсный Python-скрипт `/Users/jtprogru/Work/tmp/things3db/export.py`, плюс добавлять:

1. **Параметризацию формата** (JSON по умолчанию; Markdown как первая дополнительная цель; архитектура должна позволять добавлять YAML/CSV без переделок).
2. **Параметризацию состава** (только задачи; задачи+теги; задачи+проекты+области; «всё и сразу»; и логичные комбинации).
3. **Инфраструктуру как у `todushka`**: `Taskfile.yml`, `.golangci.yml`, `.goreleaser.yaml` с публикацией Homebrew-каска в `homebrew-tap`, GitHub Actions CI + release с govulncheck/cosign/syft, Dependabot.

Триггер — пользователь хочет переписать одноразовый Python-скрипт в распространяемый бинарь, который не требует Python и устанавливается одной командой через `brew install jtprogru/tap/thingsexporter`. TUI явно не нужен.

## Investigation

Проектная документация (`.spec/`) ещё не существует — репозиторий пустой. Контекст собран из двух соседних проектов.

### Референсный Python-скрипт `/Users/jtprogru/Work/tmp/things3db/export.py`

Логика, которую обязательно нужно воспроизвести (цитирую файлы и строки):

- **Открытие БД read-only**: `sqlite3.connect(f"file:{path}?mode=ro", uri=True)` (`export.py:77-78`). Это критично — БД Things 3 живая и может быть открыта самим приложением; писать туда нельзя.
- **Таблицы, которые читаются полностью**: `TMArea`, `TMTag`, `TMTask`, `TMChecklistItem`, `TMContact`, `TMTombstone`, `Meta` (`export.py:80-88`).
- **Join-таблицы**: `TMTaskTag(tasks, tags)`, `TMAreaTag(areas, tags)` — читаются как пары UUID (`export.py:85-86`).
- **Конвертация дат**:
  - Core Data timestamp (`REAL` — секунды от `2001-01-01 UTC`) → ISO 8601 в UTC. Функция `core_data_to_iso` (`export.py:26-34`). Применяется к: `creationDate`, `userModificationDate`, `stopDate`, `lastReminderInteractionDate`, `repeaterMigrationDate`, `usedDate`, `deletionDate`.
  - Packed date (`INTEGER` — Things 3 хранит `startDate`/`deadline`/`deadlineSuppressionDate` как `(year<<16)|(month<<12)|(day<<7)`). Функция `packed_date_to_iso` (`export.py:37-57`). Производится валидация диапазонов: `1970 ≤ year ≤ 2100`, `1 ≤ month ≤ 12`, `1 ≤ day ≤ 31`.
- **Перевод enum-кодов в имена**: `TASK_TYPE = {0: todo, 1: project, 2: heading}`, `TASK_STATUS = {0: open, 2: canceled, 3: completed}`, `CHECKLIST_STATUS = {0: open, 3: completed}` (`export.py:21-23`).
- **BLOB-поля** (`cachedTags`, `experimental`, `recurrenceRule`, `repeater`, `Meta.value`, `definition`) сериализуются как `{"__blob_hex__": "<hex>"}` либо `null` (`export.py:67-69`). Содержимое (XML-плисты, бинарный формат повторов) **не парсится** — это вне MVP.
- **Обогащение задач**: каждой задаче добавляются `areaTitle`, `projectTitle`, `headingTitle`, `contactName`, развёрнутый список тегов (`tags: [{uuid, title}]`) и собственный чек-лист, отсортированный по `index` (`export.py:134-157`).
- **Иерархический срез `hierarchy`**: `areas[].items[]` (только не-trashed корневые задачи/проекты этой области) и `inbox_or_orphan_tasks[]` (`export.py:167-211`).
- **Корневой объект**: `{meta, areas, tags, tasks, checklistItems, contacts, tombstones, links, hierarchy}` (`export.py:213-240`). `meta.counts` содержит точные счётчики каждой коллекции — это естественный test oracle.

### Схема `main.sqlite` (выгружена через `sqlite3 .schema`)

Подтверждены типы и наличие колонок. Существенные детали для Go-моделей:

- `TMTask` — 39 колонок, в том числе зарезервированные слова `index`, `type`, `status` (нужно квотировать в SQL и в Go-структурах использовать другие имена с тегами `db:"index"`).
- BLOB-колонки: `TMTask.cachedTags`, `TMTask.rt1_recurrenceRule`, `TMTask.experimental`, `TMTask.repeater`, `TMArea.cachedTags`, `TMArea.experimental`, `TMSettings.experimental`, `TMTag.experimental`, `TMChecklistItem.experimental`, `Meta.value` (TEXT с XML-плистом), `TMMetaItem.value`, `BSSyncronyMetadata.value`, `TMSmartList.definition`.
- Индексы по `TMTask(project)`, `TMTask(area)`, `TMTask(heading)`, `TMTaskTag(tasks)`, `TMAreaTag(areas)`, `TMChecklistItem(task)`, `TMTombstone(deletedObjectUUID)` — выгрузка идёт без `WHERE`, индексы не критичны, но порядок ввода это не определяет, поэтому в Go-коде нужно сортировать там, где требуется детерминизм.
- Лишние таблицы (которые Python-скрипт **не читает** и MVP тоже не читает): `TMSettings`, `TMSmartList`, `TMMetaItem`, `BSSyncronyMetadata`, `ThingsTouch_ExtensionCommandStore_*`, `sqlite_sequence`. Из них имеет смысл по запросу выгрузить только `TMSettings` и `TMSmartList`; остальные — служебные.

### Реальный размер тестовой БД

`sqlite3 ... "SELECT COUNT(*) FROM TMTask"` → 613 задач, 4 области, 3 тега, 22 чек-лист-айтема, 55 связей task↔tag (т.е. ~700 строк суммарно). На таком объёме скорость не проблема — главное корректность.

### Инфраструктура `todushka` (`/Users/jtprogru/Work/github/jtprogru/todushka`)

Что нужно скопировать в `thingsexporter` практически 1-в-1:

- **`.goreleaser.yaml`** (`todushka/.goreleaser.yaml:1-115`):
  - `version: 2`, `CGO_ENABLED=0`, цели `linux/darwin × amd64/arm64`, `flags: -trimpath`.
  - `ldflags` инжектят `Version/Commit/Date/BuiltBy` в пакет `internal/version` (нужно создать симметричный пакет).
  - `archives: tar.gz` с шаблонным именем.
  - `checksum` SHA-256, `sboms` через syft, `signs` через cosign keyless (Sigstore OIDC).
  - `snapshot.version_template: "{{ incpatch .Version }}-next"`.
  - `changelog` группирует по conventional commits (`feat`, `fix`).
  - **`homebrew_casks`** публикует в `jtprogru/homebrew-tap@main` с токеном `HOMEBREW_TAP_GITHUB_TOKEN`. Включает `post-install` hook, снимающий quarantine на macOS.
- **`.github/workflows/ci.yml`** (`todushka/.github/workflows/ci.yml:1-58`):
  - Job `test`: `go vet`, `govulncheck`, `go test -race -coverprofile=cover.out ./...`.
  - Job `goreleaser-check`: `goreleaser check` + `goreleaser build --snapshot --clean --single-target`.
  - Actions запиннены по SHA (best practice, ничего не меняем).
- **`.github/workflows/release.yml`** (`todushka/.github/workflows/release.yml:1-50`):
  - Триггер на тег `v*`, права `contents: write` + `id-token: write` (для Sigstore).
  - `govulncheck` → `go test ./...` → cosign installer → syft installer → `goreleaser release --clean`.
  - Передаёт `GITHUB_TOKEN` и `HOMEBREW_TAP_GITHUB_TOKEN`.
- **`.github/dependabot.yml`**: weekly `gomod` + `github-actions`, лимит 5 PR, префиксы коммитов `chore(deps)`/`chore(ci)`.
- **`Taskfile.yml`** (`todushka/Taskfile.yml:1-59`): таргеты `test`, `test-race`, `build`, `lint`, `fmt`, `tidy`, `run`, `cross-compile`. Переменные `BIN_DIR`, `CMD_PATH`.
- **`.golangci.yml`**: v2-формат, линтеры `govet/staticcheck/errcheck/gosec/gocritic/revive/unused/ineffassign`, форматтеры `gofmt/goimports`, `gosec` исключает `G104`, в тестах отключены `gosec` и `errcheck`.

### Структура проекта `todushka`

`cmd/<binary>/main.go` + `internal/{cli,version,…}`. CLI собирается на `github.com/spf13/cobra` через `cli.Deps`-структуру (явный test seam — DI streams/env). Версия печатается отдельной командой `version`, формат вывода человекочитаемый. Конвенция явно подходит и `thingsexporter`.

## Build Tooling

- **Orchestrator:** [Taskfile.dev](https://taskfile.dev) (повторяем `todushka`).
- **Test:** `task test` (под капотом `go test ./...`); `task test-race` для CI.
- **Build:** `task build` → `bin/thingsexporter`; `task cross-compile` для linux/darwin × amd64/arm64.
- **Lint:** `task lint` → `golangci-lint run` (v2 config).
- **Fmt:** `task fmt` → `go fmt ./...` + `goimports -w .`.
- **Generate:** не требуется (нет proto/моков/ORM-генераторов).
- **Source:** новый `Taskfile.yml` в корне репозитория, перенесённый из `todushka` с подстановкой бинарника.

## Options Considered

### Драйвер SQLite

#### Option A: `modernc.org/sqlite` (pure Go, без CGO) — **рекомендуется**
- Pluggable в `database/sql`, DSN `file:path?mode=ro` для read-only.
- Совместим с `CGO_ENABLED=0`, что критично для текущей `.goreleaser.yaml` (сейчас в `todushka` стоит `CGO_ENABLED=0`).
- Cross-compile тривиален: один и тот же go-код собирается под linux/darwin × amd64/arm64 без дополнительных тулчейнов.
- Зрелый, активно поддерживается, используется в большом числе production-проектов.
- Минусы: больше размер бинаря (~10–15 МБ), запуск чуть медленнее `mattn/go-sqlite3` на больших нагрузках. Для разовой выгрузки 600 строк — несущественно.

#### Option B: `mattn/go-sqlite3` (CGO-обёртка над системным SQLite)
- Быстрее в рантайме, бинарь меньше.
- Требует CGO → ломает текущий `CGO_ENABLED=0` в goreleaser, требует cross-toolchain (zig/osxcross) для macOS-сборок из linux-раннера CI. Это значительное усложнение CI ради 200 мс выигрыша на выгрузке.

#### Option C: `ncruces/go-sqlite3` (WASM-runtime)
- CGO-free, но через wazero WASM SQLite — больше cold-start overhead, экзотичнее экосистема.
- Зрелость инфраструктуры ниже `modernc.org/sqlite`.

**Вывод:** Option A. CGO-free — единственный путь, совместимый с принятой инфраструктурой релизов.

### Формат CLI и параметризация состава

#### Option A: Один корневой `thingsexporter` с флагами `--format` и `--include` — **рекомендуется**
- Команда: `thingsexporter --db main.sqlite --format json --include all --out things3.json`.
- `--format`: `json` (default), `markdown`. Архитектура — интерфейс `Writer{ Write(io.Writer, ExportData) error }`, регистрация через map[format]factory.
- `--include`: enum-набор пресетов состава:
  - `tasks` — только задачи (без связей);
  - `tasks+tags` — задачи с резолвом тегов;
  - `tasks+projects` — задачи с резолвом проектов/областей/заголовков;
  - `structure` — areas + projects + tags без тел задач (оглавление);
  - `all` (default) — полный экспорт как у Python-скрипта (включая `tombstones`, `contacts`, `checklistItems`, `links`, `hierarchy`).
- Плюсы: один бинарь, простая mental model, легко скриптовать.

#### Option B: Подкоманды `thingsexporter export json|markdown`
- Симметрично `kubectl get`/`docker compose`. Каждая подкоманда — свой набор флагов.
- Плюсы: каждый формат может иметь специфичные флаги (например, `markdown` — `--heading-level`, `--checkbox-style`); меньше «магических» enum-флагов.
- Минусы: больше boilerplate; пользователю сложнее переключать формат в shell-скриптах (нужно менять имя команды, а не значение флага).

#### Option C: Гибрид — корневая команда + подкоманды
- `thingsexporter` без аргументов = `thingsexporter export --format json --include all` (sane default).
- Подкоманды `export`, `inspect` (счётчики без записи), `version`, `completion` (shell completion от cobra).
- Плюсы: расширяемо без слома совместимости; место для будущих режимов (например, `thingsexporter diff` против предыдущей выгрузки).

**Вывод:** Option C. Корневой command с пресет-флагами + явные подкоманды `version`/`inspect`/`completion`. Pure flags-on-root оставим как default-поведение для совместимости с привычкой Python-скрипта.

### Markdown-схема

#### Option A: Иерархия Areas → Projects → Tasks с GFM-чекбоксами — **рекомендуется**
- `# Area title` → `## Project title` → `- [ ] Task title` (или `[x]` для completed, `[~]` для canceled — кастомизировать через флаг отдельной итерации).
- Метаданные задачи (теги, deadline) выводятся inline: `- [ ] buy milk @home  #P1 ⏰ 2026-06-01`.
- Notes уходят в индентированный блок под задачей.
- Чек-лист — вложенный список под задачей.
- Inbox/orphan задачи — секция `# Inbox` в начале.
- Это естественный человекочитаемый формат для просмотра в Obsidian/Bear/Logseq.

#### Option B: Plain markdown-таблица
- Колонки: `| title | status | area | project | tags | deadline |`. Хорошо для grep, плохо для чтения глазами и для больших notes.

#### Option C: YAML frontmatter + body per task (для Obsidian)
- По одному `.md` файлу на задачу — отдельный режим, выходит за MVP (требует `--out-dir` вместо `--out`).

**Вывод:** Option A для MVP; Option C — кандидат на v2.

## Constraints & Risks

- **Read-only открытие БД обязательно.** Things 3 пишет в `main.sqlite` непрерывно (WAL + SHM файлы рядом подтверждают активный режим). Любая попытка write-открытия приведёт к повреждению или блокировке. В Go: DSN `file:<path>?mode=ro&immutable=1&_pragma=query_only(1)` (синтаксис modernc.org/sqlite).
- **WAL/SHM файлы.** Если приложение Things 3 запущено, последние правки сидят в `main.sqlite-wal` и видны только при правильном read-only открытии файла с теми же режимами журналирования. `mode=ro` без `immutable=1` достаточно — SQLite сам подтянет WAL. `immutable=1` нужен, если БД на read-only ФС; для основного use-case не обязателен и может скрыть свежие правки. **MVP: `mode=ro` без `immutable`.** В документации честно предупреждаем, что можно запускать при работающем Things 3.
- **Bытащить «как у Python»**: bit-to-bit совпадение JSON не гарантировано (порядок ключей в map в Go-stdlib `encoding/json` сортирует по алфавиту; Python `json.dump` сохраняет insertion order). Семантическое совпадение возможно и его и проверяем (счётчики, наличие полей, корректность дат). Это нужно явно зафиксировать в Requirements.
- **Зарезервированные слова SQL** (`index`, `type`, `status`, `start`) — обязательны двойные кавычки в `SELECT`. modernc.org/sqlite это поддерживает; нужны table-driven тесты SQL-генерации, чтобы не словить regression.
- **Локали и таймзоны.** Все Core Data timestamps преобразуем в UTC ISO 8601 (как Python-скрипт). Packed dates — в `YYYY-MM-DD` без таймзоны (это «дата без времени» в Things 3).
- **BLOB-поля.** В MVP сериализуем как hex (повторение Python-поведения); парсинг плистов и `recurrenceRule` — Deferred / Needs spike.
- **Размер бинаря с `modernc.org/sqlite`.** Ожидаемо 10–15 МБ. Для Homebrew cask это норма; SBOM и подписи продолжают работать.
- **Cosign keyless + Sigstore.** Требует `id-token: write` permission и работает только из GitHub Actions (как в `todushka`). Локальный `goreleaser release` без cosign невозможен — но локально мы используем `goreleaser build --snapshot --single-target`, который не подписывает. Это уже учтено в инфраструктуре `todushka`, копируем без изменений.
- **Homebrew-tap токен.** В `release.yml` обязателен `HOMEBREW_TAP_GITHUB_TOKEN` (классический PAT с правами `repo` на `jtprogru/homebrew-tap`). Уже выдан для `todushka` — пользователь подтвердил, что добавит в repo secrets, когда дойдём до настройки CI. **Без него `goreleaser release` упадёт.** В Requirements зафиксируем этот dependency как явное предусловие первого релиза.
- **`homebrew_cask` vs `homebrew_formula`.** `todushka` использует `homebrew_casks` — этот блок предназначен для приложений с macOS-specific шагами (например, снятие quarantine). Для CLI-инструмента более идиоматичен `homebrew_formula` (работает и на linux через homebrew-on-linux). **Question to user:** оставить `homebrew_casks` для симметрии с `todushka` или перейти на `homebrew_formula` (рекомендуем formula для чистого CLI). Закрепим в Requirements.
- **Документация к плистам в `Meta.value`.** `databaseVersion` сейчас 26 (на тестовой БД). Things 3 регулярно обновляет схему — нужно проверять `databaseVersion` при экспорте и предупреждать, если значение неизвестное. Это страховка от молчаливой поломки на будущих версиях Things. **MVP:** читать `databaseVersion` и выводить как мета-предупреждение в `stderr`, если оно не в списке `[26]`. Хардкод-список — Deferred.
- **Стандартный `database/sql` vs прямой драйвер.** Использование `database/sql` даёт стандартизацию и помогает в тестах через `DATA-MOCK`; прямой driver минимально быстрее. Выбираем `database/sql` для совместимости с инструментами и читаемости.
- **Параллелизм.** На 700 строк он не нужен. MVP — синхронный последовательный read.
- **Тестовые данные.** В репозитории `thingsexporter` нельзя коммитить реальную `main.sqlite` пользователя (содержит личные задачи). Нужен генератор fixture-БД (Go-функция, создающая in-memory SQLite со схемой и набором представительных строк) либо anonymized snapshot. **Решение для MVP:** генератор fixture в `internal/testdata`, который создаёт SQLite-файл с минимальным DDL и контролируемым набором данных. Реальная пользовательская БД — только для ручного smoke-теста.

## Recommended Direction

**Стек и архитектура:**

- **Go**: версия — последняя совместимая с экосистемой `todushka` (`go.mod` `todushka` сейчас `go 1.26.3`; повторяем).
- **SQLite-драйвер**: `modernc.org/sqlite` через `database/sql`. DSN: `file:<path>?mode=ro`.
- **CLI**: `github.com/spf13/cobra` (как в `todushka`).
- **Структура проекта** (отражает `todushka`):
  ```
  cmd/thingsexporter/main.go         — точка входа
  internal/cli/                      — cobra commands + Deps test-seam
    root.go, export.go, inspect.go, version.go, completion.go, deps.go
  internal/version/version.go        — Version/Commit/Date/BuiltBy
  internal/things/                   — domain: модели Area/Tag/Task/Checklist/Contact/Tombstone
    types.go, dates.go (core_data + packed), enums.go
  internal/store/sqlite/             — чтение БД (database/sql + modernc)
    open.go, queries.go, repo.go, fixture.go (для тестов)
  internal/export/                   — Writer interface + регистрация форматов
    writer.go, json.go, markdown.go, preset.go (include presets)
  internal/testdata/                 — генератор fixture-БД
  ```
- **Формат JSON**: совместим с Python-выгрузкой по составу и значениям; порядок ключей не гарантируется (фиксируем в Requirements).
- **Markdown**: иерархия Areas → Projects → Tasks с GFM-чекбоксами; пресет полей определяется `--include`.
- **Параметризация**:
  - `--db <path>` (required, без default — намеренно: пути к БД Things 3 OS-зависимы и пользователь должен явно указать);
  - `--out <path>` (default: `-`, т.е. stdout);
  - `--format json|markdown` (default: `json`);
  - `--include all|tasks|tasks+tags|tasks+projects|structure` (default: `all`);
  - `--indent <int>` (default: `2`; `0` = компактный, только для JSON);
  - `--quiet` (не печатать сводку в stderr).
- **Подкоманды (cobra)**:
  - корневая (без подкоманды) = `export` с дефолтами для совместимости с Python-скриптом;
  - `export` — явная форма;
  - `inspect` — печатает только `meta.counts` в stdout без выгрузки данных (быстрый health-check БД);
  - `version` — версия + commit + дата + Go-toolchain (повторяем формат `todushka`);
  - `completion` — стандартный cobra-генератор shell-completion.

**Инфраструктура:**

- `Taskfile.yml`, `.golangci.yml`, `.goreleaser.yaml`, `.github/{workflows/ci.yml,workflows/release.yml,dependabot.yml}` — переносятся из `todushka` с минимальной подстановкой `todushka` → `thingsexporter` и адаптацией `homepage`/`description`.
- Решение по `homebrew_casks` vs `homebrew_formula` — **рекомендуем `homebrew_formula`** (формула, не каск): это CLI-инструмент без macOS-специфики, формула работает и на linuxbrew, и сборка проще (нет post-install hook со снятием quarantine).
- `LICENSE` — MIT (как `todushka`).
- `README.md` — минимальный (install / usage / examples), без TUI-секций.

**Что НЕ делаем в MVP:**

- TUI, watch-режим, импорт обратно в Things 3, синхронизация, парсинг плистов/recurrenceRule в человекочитаемый вид, отдельные форматы YAML/CSV/per-task-markdown.

## Scope Boundaries

- **Must-have (v1):**
  - CLI `thingsexporter` на cobra с командами `export` (default), `inspect`, `version`, `completion`.
  - Чтение `main.sqlite` через `modernc.org/sqlite` в режиме `mode=ro`.
  - Полное воспроизведение семантики Python-скрипта для JSON: все 7 коллекций + `links` + `hierarchy` + `meta.counts`.
  - Конвертация Core Data timestamps и packed dates 1-в-1 с Python-скриптом (включая `null` для невалидных).
  - Маппинг enum-кодов (`typeName`, `statusName`).
  - BLOB → hex (`{"__blob_hex__": "..."}`).
  - Markdown-формат: иерархический список Areas → Projects → Tasks с чекбоксами, тегами и дедлайнами.
  - Параметризация `--format` и `--include` (5 пресетов).
  - `--db`, `--out`, `--indent`, `--quiet`.
  - Структурированная сводка в stderr (как Python) при отсутствии `--quiet`.
  - `Taskfile.yml`, `.golangci.yml`, `.goreleaser.yaml`, CI + release workflows, dependabot.
  - Goreleaser публикует в Homebrew formula (или cask — финальное решение в Requirements).
  - SBOM (syft) + cosign keyless подписи (повторяем `todushka`).
  - `internal/version` с `Version/Commit/Date/BuiltBy`.
  - Unit-тесты: fixture-БД + golden JSON; конверсия дат; маппинг enum; SQL-генерация для зарезервированных слов; markdown-сериализация.
  - `README.md`, `LICENSE`.
- **Deferred (v2):**
  - Парсинг XML-плистов в `Meta.value` (databaseVersion и пр.) в человекочитаемые поля.
  - Парсинг `rt1_recurrenceRule` и `repeater` BLOB в человекочитаемые RRULE.
  - Форматы YAML, CSV, per-task-Markdown (Obsidian-style).
  - Фильтры: `--include-trashed`, `--status`, `--area`, `--tag`, `--since <date>`.
  - Авто-обнаружение пути к `main.sqlite` (`~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite` для macOS App-Store-версии).
  - `diff` против предыдущей выгрузки.
  - Watch-режим.
- **Needs spike:**
  - Корректное чтение БД, **открытой Things.app** (WAL-режим, lock-conflict-free). Если возникают `SQLITE_BUSY`, нужно тестировать на live-БД и, возможно, добавлять флаг `--snapshot` (копировать БД в `tmp` перед чтением). MVP опирается на `mode=ro` и описанные предположения.
  - Будущие миграции схемы Things 3 (`databaseVersion > 26`) — нужно построить regression-набор fixture-БД от разных версий, как только они появятся.

## Assumptions & Open Questions

**Assumptions** (требуют подтверждения, прежде чем перейти в Requirements):

- `[ASSUMPTION: Go-модуль будет называться "github.com/jtprogru/thingsexporter", бинарь — "thingsexporter".]`
- `[ASSUMPTION: Семантическое совпадение JSON-выгрузки с Python-скриптом считается достаточным; bit-to-bit совпадение ключей и форматирования не обязательно.]`
- `[ASSUMPTION: Для Homebrew используем "homebrew_formula" (а не "homebrew_casks", как в todushka) — это чистый CLI без macOS-специфики и должен работать на linuxbrew.]`
- `[ASSUMPTION: Тестовая fixture-БД генерируется в Go-коде (DDL + INSERT через database/sql); реальный "main.sqlite" пользователя в репозиторий не коммитится.]`
- `[ASSUMPTION: Целевые ОС/архитектуры повторяют todushka: linux/darwin × amd64/arm64. Windows и FreeBSD — вне scope.]`
- `[ASSUMPTION: Минимальная Go-версия — 1.26.3, как в todushka.]`
- `[ASSUMPTION: HOMEBREW_TAP_GITHUB_TOKEN тот же, что используется для todushka; пользователь добавит его в repo secrets thingsexporter перед первым релизом.]`
- `[ASSUMPTION: LICENSE — MIT, copyright owner — "Mikhail Savin <jtprogru@gmail.com>" (по аналогии с todushka и git user в этом репозитории).]`
- `[ASSUMPTION: Корневая команда без аргументов выполняет "export --format json --include all --out -" (stdout). Это сохраняет привычный UX Python-скрипта при запуске без флагов.]`
- `[ASSUMPTION: При --out пути, заканчивающемся на ".json"/".md", формат можно НЕ указывать явно (авто-детект); но если --format задан — он приоритетнее. Это удобство, не критично для MVP — могу убрать, если избыточно.]`

**Open Questions:**

1. Подтвердить имя Go-модуля и бинаря (`thingsexporter` vs альтернативы вроде `things3-export`, `things-dump`).
2. `homebrew_formula` или `homebrew_casks`? (рекомендую formula).
3. Нужно ли в MVP авто-определение пути к БД на macOS, или достаточно требовать явный `--db`? (по умолчанию — требовать явный путь, как и Python-скрипт).
4. Нужно ли уже на стадии MVP добавить флаг `--snapshot`, который копирует БД во временный файл перед чтением (защита от lock-conflict с запущенным Things.app)? (по умолчанию — нет, MVP полагается на `mode=ro`).
5. Семейство пресетов `--include`: пять достаточно или нужны другие комбинации? Конкретно «только теги», «только области», «только тумбстоны» — стоит ли выделить?
6. Markdown: какой стиль чекбокса использовать для `canceled`? (`[~]`, `[-]`, `[x]` со страйкаутом, отдельный emoji). Хардкод vs флаг.
7. Нужны ли в JSON-формате дополнительные опции (например, `--no-blobs` — выкинуть BLOB-поля полностью)?
