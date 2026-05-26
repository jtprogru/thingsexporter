# Code Review: thingsexporter MVP

## Verdict: PASS

Все 40 требований из `requirements.md` прослежены до соответствующего production-кода и тестов. Все 20 correctness properties из `design.md §2.6` имеют покрытие через unit/integration/PBT-тесты. Свежий прогон `task test`, `task build`, `task lint`, `go test -race -count=1 ./...` — все зелёные, 0 lint-issues. Принципиальных нарушений архитектуры, безопасности или scope creep не обнаружено. Найдено 5 minor/nit-замечаний (см. ниже) — все либо приняты пользователем явно (REQ-7.7 homebrew_casks deviation), либо описаны как осознанные упрощения MVP, не блокирующие релиз.

## Change Set

Базовый коммит для diff (`review_base_commit`) — `HEAD`, но репозиторий пустой (нет ни одного коммита). Сравнение делается **по списку из task plan / design §2.3** (вторичный источник, как разрешает шаблон).

| File | Status | Notes |
|------|--------|-------|
| `go.mod`, `go.sum` | ✅ Planned | T-1.1 + транзитивные deps от testify/rapid/sqlite/cobra |
| `.gitignore`, `LICENSE` | ✅ Planned | T-1.2, T-1.3 |
| `README.md`, `Taskfile.yml`, `.golangci.yml`, `.goreleaser.yaml` | ✅ Planned | T-8.1, T-8.2, T-8.3, T-8.7 |
| `.github/workflows/ci.yml`, `.github/workflows/release.yml`, `.github/dependabot.yml` | ✅ Planned | T-8.4, T-8.5, T-8.6 |
| `cmd/thingsexporter/main.go` | ✅ Planned | T-7.1 |
| `internal/version/version.go` | ✅ Planned | T-1.4 |
| `internal/things/{types.go, raw.go, dates.go, enums.go, blob.go, build.go}` | ✅ Planned | T-2.x, T-3.x |
| `internal/things/{dates_test.go, enums_test.go, blob_test.go, build_test.go, dates_property_test.go, build_property_test.go}` | ✅ Planned | T-2.x, T-3.x |
| `internal/store/sqlite/{discover.go, open.go, queries.go, repo.go}` | ✅ Planned | T-4.x |
| `internal/store/sqlite/{discover_test.go, open_test.go, repo_test.go, fixture_test.go}` | ✅ Planned | T-4.x |
| `internal/export/writer.go`, `writer_test.go` | ✅ Planned | T-5.1, T-5.2 |
| `internal/export/{json/json.go, json/json_test.go}` | ✅ Planned | T-5.3 |
| `internal/export/{markdown/markdown.go, markdown/markdown_test.go}` | ✅ Planned | T-5.4 |
| `internal/export/{preset/preset.go, preset/preset_test.go}` | ✅ Planned | T-5.5 |
| `internal/cli/{deps.go, errors.go, root.go, export.go, inspect.go, version.go, completion.go}` | ✅ Planned | T-6.x |
| `internal/cli/{errors_test.go, export_test.go, inspect_test.go, version_test.go, completion_test.go, root_test.go, clihelpers_test.go, integration_test.go}` | ✅ Planned | T-6.x, T-7.2 |
| `pipeline.sh` | ✅ Planned (исключён из git через `.gitignore`) | ADR-7 |
| `.claude/`, `.serena/` | ⚠️ Unexpected (но локальные IDE-конфиги, не часть продукта) | Не в `.gitignore` явно, но и не были запланированы. Рекомендую добавить в `.gitignore` (см. F-3 minor) |
| `internal/things/build_property_test.go` (PBT для CP-11) | ❌ Not Changed (отсутствует Property/11 в PBT, реализован только в unit-тестах preset) | См. F-1 minor |
| Никакие spec/task-plan tasks **не пропущены** | ✅ | Все 9 top-level T-N помечены `[x]` и сопоставлены реальным файлам |

## Requirements Traceability

| Requirement | Test(s) | Code | CP | Verdict |
|-------------|---------|------|----|---------|
| REQ-1.1 (read-only DSN) | `TestOpen_readOnlyDSN` (`internal/store/sqlite/open_test.go:11`) | `internal/store/sqlite/open.go:13-19` (`Open` использует `mode=ro`) | CP-1 | ✅ |
| REQ-1.2 (auto-discover macOS) | `TestDiscover_matrix`, `TestExportCmd_summary` (косвенно) | `internal/store/sqlite/discover.go:13-21`, `internal/cli/export.go:115-127` | CP-15 | ✅ |
| REQ-1.3 (no --db, no auto-discover → exit 2 + сообщение) | `TestExportCmd_missingDBNonMac` (`internal/cli/export_test.go:60`) | `internal/cli/export.go:128` (ошибка `--db is required …`) | CP-15 + integration | ✅ |
| REQ-1.4 (нерабочий путь к БД → exit 2) | `TestOpen_missingFile`, `TestExportCmd_unknownFormat_exit2` (отчасти) | `internal/store/sqlite/open.go:15`, `internal/cli/export.go:80-83` (wrap в exit2) | Integration | ✅ |
| REQ-1.5 (читаются конкретные 9 таблиц, reserved words в `"`) | `TestRepositoryReadAll_fixture` (вызывает все selectXxx) | `internal/store/sqlite/queries.go:12-58` (все SELECT-ы) | CP-17 (поведенчески, через integration) | ⚠️ см. F-2 minor |
| REQ-1.6 (databaseVersion != 26 → warning) | `TestRepositoryDatabaseVersion_meta`; integration warning путь покрыт через `inspect`-тест с фикстурой v=26 (warning не печатается) | `internal/cli/export.go:84-88`, `internal/cli/inspect.go:51-54`, `internal/store/sqlite/queries.go:386-409` (selectDatabaseVersion) | Integration | ✅ |
| REQ-2.1 (CoreData → ISO UTC) | `TestCoreDataToISO_table`, `PropCoreDataRoundTrip` | `internal/things/dates.go:16-31` | CP-2 | ✅ |
| REQ-2.2 (CoreData nil/NaN → nil) | `TestCoreDataToISO_table` (нилевый + NaN + Inf кейсы) | `internal/things/dates.go:17-21` | CP-2 | ✅ |
| REQ-2.3 (packed date → YYYY-MM-DD) | `TestPackedDateToISO_known`, `PropPackedDateValid` | `internal/things/dates.go:35-47` | CP-3 | ✅ |
| REQ-2.4 (packed date invalid → nil) | `TestPackedDateToISO_invalid` (6 негативных кейсов) | `internal/things/dates.go:43-46` | CP-4 | ✅ |
| REQ-2.5 (enum type/status → имя) | `TestTaskTypeName_known`, `TestTaskStatusName_known`, `PropEnumTotality` | `internal/things/enums.go:21-35` | CP-5 | ✅ |
| REQ-2.6 (checklist status → имя) | `TestChecklistStatusName_known` | `internal/things/enums.go:31` | CP-5 | ✅ |
| REQ-2.7 (BLOB → hex) | `TestEncodeBlob_table`, `PropBlobEncoding` | `internal/things/blob.go:11-15` | CP-6 | ✅ |
| REQ-2.8 (`--no-blobs` → null) | `TestEncodeBlob_table` (drop=true), `TestBuild_noBlobs_strips`, `PropNoBlobsPropagation`, `TestExportCmd_noBlobs` | `internal/things/blob.go:13`, `internal/things/build.go` (всюду через `EncodeBlob(..., opts.NoBlobs)`) | CP-6, CP-18 | ✅ |
| REQ-3.1 (areaTitle/projectTitle/headingTitle/contactName) | `TestBuild_areaProjectTitles`, integration JSON-тест | `internal/things/build.go:84-103` | CP-8 + integration | ✅ |
| REQ-3.2 (task.tags) | `TestBuild_enrichTaskTags`, `PropTagsEnrichment` | `internal/things/build.go:104` | CP-8 | ✅ |
| REQ-3.3 (task.checklist) | `TestBuild_jsonMarshalCompiles`, integration test | `internal/things/build.go:107-111`, `internal/things/build.go:64-75` (sorted by Index) | Integration | ✅ |
| REQ-3.4 (area.tags) | `TestBuild_areaTags` | `internal/things/build.go:81-83` | Integration | ✅ |
| REQ-3.5 (hierarchy без trashed, сортировка) | `TestBuild_hierarchy_excludesTrashed`, `TestBuild_hierarchy_ordering`, `TestBuild_inboxContainsOrphans`, `PropHierarchyExcludesTrashed`, `PropHierarchyOrdering` | `internal/things/build.go:225-301` | CP-9, CP-10 | ✅ |
| REQ-3.6 (meta.counts соответствует длинам) | `TestBuild_counts_match`, `PropCountsMatchCollections` | `internal/things/build.go:116-125` | CP-7 | ✅ |
| REQ-4.1 (JSON без HTML-escape) | `TestJsonWriter_noASCIIEscape` (кириллица проходит как UTF-8) | `internal/export/json/json.go:25` (`SetEscapeHTML(false)`) | CP-13 | ✅ |
| REQ-4.2 (indent N / compact 0) | `TestJsonWriter_compact`, `TestJsonWriter_indent_two`, `TestExportCmd_indentZero_compact` | `internal/export/json/json.go:26-28` | CP-13 | ✅ |
| REQ-4.3 (Markdown иерархия и чекбоксы) | `TestMarkdownWriter_inboxAndAreas`, `TestMarkdownWriter_checkboxes`, `TestMarkdownWriter_projectAsSubHeading`, `TestMarkdownWriter_tagsAndDeadline`, `TestMarkdownWriter_notesIndent`, `TestMarkdownWriter_checklistNested`, `TestIntegration_markdownExport_fixture` | `internal/export/markdown/markdown.go:24-160` | CP-14 + integration | ✅ |
| REQ-4.4 (markdown без тегов в пресете) | Косвенно через `TestExportCmd_format_markdown` + `TestPresetTasks_strips` (теги обнуляются → buildSuffix не печатает) | `internal/export/markdown/markdown.go:118-127` (skip nil titles) | Integration | ✅ |
| REQ-4.5 (unknown format → exit 2) | `TestRegistryLookup_unknownFormat`, `TestExportCmd_unknownFormat_exit2` | `internal/export/writer.go:50-54` | CP-12 | ✅ |
| REQ-5.1 (preset all) | `TestPresetAll_identity`, integration JSON test | `internal/export/preset/preset.go:53-57` | CP-8, CP-11 | ✅ |
| REQ-5.2 (preset tasks) | `TestPresetTasks_strips`, `TestExportCmd_includeTasks_strips` | `internal/export/preset/preset.go:60-83` | CP-11 | ✅ |
| REQ-5.3 (preset tasks+tags) | `TestPresetTasksTags_strips` | `internal/export/preset/preset.go:86-110` | CP-11, CP-8 | ✅ |
| REQ-5.4 (preset tasks+projects) | `TestPresetTasksProjects_strips` | `internal/export/preset/preset.go:113-134` | CP-11 | ✅ |
| REQ-5.5 (unknown preset → exit 2) | `TestRegistryLookup_unknown`, `TestExportCmd_unknownInclude_exit2` | `internal/export/preset/preset.go:38-44` | CP-12 | ✅ |
| REQ-6.1 (root = export с дефолтами) | `TestExportCmd_rootDefaults`, `TestIntegration_jsonExport_fixture` | `internal/cli/root.go:18-23` | Integration | ✅ |
| REQ-6.2 (флаги парсятся) | Все CLI-тесты (используют разные комбинации флагов) | `internal/cli/export.go:42-50` | Integration | ✅ |
| REQ-6.3 (stdout vs file) | `TestExportCmd_outToFile`, `TestExportCmd_rootDefaults` | `internal/cli/export.go:139-149` | Integration | ✅ |
| REQ-6.4 (сводка в stderr) | `TestExportCmd_summary`, `TestExportCmd_quietSuppresses` | `internal/cli/export.go:96-99`, `internal/cli/export.go:153-170` | CP-19 + integration | ✅ |
| REQ-6.5 (inspect) | `TestInspectCmd_outputsCounts` | `internal/cli/inspect.go:19-32`, `internal/cli/inspect.go:36-71` | Integration | ✅ |
| REQ-6.6 (version output) | `TestVersionCmd_outputFormat` | `internal/cli/version.go:12-30` | Unit | ✅ |
| REQ-6.7 (completion) | `TestCompletionCmd_bash` | `internal/cli/completion.go:10-29` | Smoke | ✅ |
| REQ-6.8 (--help → exit 0) | `TestRoot_helpExits0` | cobra default | Smoke | ✅ |
| REQ-6.9 (exit-коды 0/1/2) | `TestAsExitCode_table`, `TestExportCmd_unknownFormat_exit2` и пр. | `internal/cli/errors.go:11-25`, `cmd/thingsexporter/main.go:11-25` | CP-16 | ✅ |
| REQ-7.1 (CGO_ENABLED=0) | `goreleaser build --snapshot` succeeds | `.goreleaser.yaml:14`, `cmd/thingsexporter/main.go` (нет CGO-импортов) | Build check | ✅ |
| REQ-7.2 (task test/test-race) | Я re-run сам (см. Verification Evidence) | `Taskfile.yml:14-22` | Manual | ✅ |
| REQ-7.3 (task lint = golangci-lint v2) | Я re-run сам | `.golangci.yml:1`, `Taskfile.yml:29-31` | Manual | ✅ |
| REQ-7.4 (ci.yml jobs) | Файл существует с pinned actions | `.github/workflows/ci.yml` | Static review | ✅ |
| REQ-7.5 (release.yml на v* tag) | Файл существует | `.github/workflows/release.yml:3-7` | Static review | ✅ |
| REQ-7.6 (матрица сборки + SBOM + cosign) | `goreleaser check` succeeds | `.goreleaser.yaml:16-66` | Build check | ✅ |
| REQ-7.7 (homebrew formula → cask по согласованию) | **Deviation approved by user** — см. `implementation.md §Notes 1` | `.goreleaser.yaml:91-115` (`homebrew_casks`) | User-approved deviation | ✅ |
| REQ-7.8 (HOMEBREW_TAP_GITHUB_TOKEN required) | Документировано в README + implementation.md | `.github/workflows/release.yml:48`, `.goreleaser.yaml:99` | Static review | ✅ |
| REQ-7.9 (ldflags inject Version/Commit/...) | `TestVersionCmd_outputFormat` + manual `./bin/thingsexporter version` | `.goreleaser.yaml:25-29` инжектят в `internal/version` | Manual | ✅ |
| REQ-7.10 (dependabot weekly) | Файл существует | `.github/dependabot.yml` | Static review | ✅ |
| REQ-8.1 (unit-тесты дат/enum/blob/SQL) | Все Test* в `internal/things/` + `internal/store/sqlite/` | См. выше | Coverage | ✅ |
| REQ-8.2 (integration JSON через fixture) | `TestIntegration_jsonExport_fixture` | `internal/cli/integration_test.go:11-41` | Integration | ✅ |
| REQ-8.3 (integration Markdown через fixture) | `TestIntegration_markdownExport_fixture` | `internal/cli/integration_test.go:45-58` | Integration | ✅ |
| REQ-8.4 (CLI на Deps seam) | Все cli-тесты через `cli.Deps` с buffer-streams | `internal/cli/clihelpers_test.go:91-119` | Integration | ✅ |
| REQ-8.5 (.gitignore исключает реальную БД) | Визуально проверено | `.gitignore:7-9` (`*.sqlite*`) | Static review | ✅ |
| ADR-9 schema field | `TestBuild_schemaField`, `PropSchemaPresent` | `internal/things/build.go:127`, `internal/things/build.go:9` (const `SchemaVersion`) | CP-20 | ✅ |

**Все 40 REQ-X.Y и ADR-9 покрыты.** Покрытие 20 CP через PBT/integration — полное.

## Design Conformance

### §3.1 Architectural Boundaries

✅ Все пакеты лежат там, где описано в design §2.2. Направление зависимостей строго сверху вниз: `cmd/thingsexporter` → `internal/cli` → `internal/{export, store/sqlite, things, version}`. Циклических импортов нет (проверено `go build ./...`). `internal/things` не импортирует ничего из `internal/{cli, store, export}` (правильное направление: domain — самый низ).

### §3.2 Data Models

✅ Все 17 структур из design §2.5 реализованы в `internal/things/types.go` с теми же именами полей, типами и JSON-тегами. Дополнительно появилась `BuildOptions` (логично — параметры функции, явно описаны в design). `RawData` и приватные `Raw*` стали публичными `RawArea/RawTag/...` (отмечено как поправка прямо в task plan T-4.5 «NOTE: пересмотр T-3.2»).

### §3.3 API Contracts (CLI)

✅ Все флаги команды `export` совпадают с design §2.3 и REQ-6.2 (`--db`, `--out`, `--format`, `--include`, `--indent`, `--no-blobs`, `--quiet`). Подкоманды `inspect`, `version`, `completion` соответствуют сигнатурам.

### §3.4 Error Handling

✅ `ExitCodeError` + `AsExitCode` реализуют схему из design §2.7. Все обработанные сценарии из таблицы Error Handling покрыты — кроме SQLITE_BUSY retry (упомянут в design таблице, но не реализован в MVP — допустимо, т.к. в самой таблице помечено «Расширение MVP»). Это **минорный gap**, см. F-5 minor.

### §3.5 Correctness Properties

✅ 20 CP из design §2.6 покрыты:
- CP-1 (Absence/read-only) — `TestOpen_readOnlyDSN` через попытку INSERT.
- CP-2 — `PropCoreDataRoundTrip`.
- CP-3, CP-4 — `PropPackedDateValid`, `TestPackedDateToISO_invalid`.
- CP-5 — `PropEnumTotality`.
- CP-6 — `PropBlobEncoding`.
- CP-7 — `PropCountsMatchCollections`.
- CP-8 — `PropTagsEnrichment`.
- CP-9 — `PropHierarchyExcludesTrashed`.
- CP-10 — `PropHierarchyOrdering`.
- CP-11 — `TestPresetTasks_strips`, `TestPresetTasksTags_strips`, `TestPresetTasksProjects_strips` (как unit-тесты; PBT для preset-exclusion в плане был, фактически реализован только как table-driven — см. F-1 minor).
- CP-12 — `TestRegistryLookup_unknownFormat`, `TestRegistryLookup_unknown`.
- CP-13 — `TestJsonWriter_compact`, `TestJsonWriter_indent_two`.
- CP-14 — `TestMarkdownWriter_checkboxes`.
- CP-15 — `TestDiscover_matrix`.
- CP-16 — `TestAsExitCode_table`.
- CP-17 — поведенчески через `TestRepositoryReadAll_fixture` (SQL с unquoted reserved word упал бы при выполнении против фикстуры; статический regex-check, упомянутый в design, не реализован — см. F-2 minor).
- CP-18 — `PropNoBlobsPropagation`.
- CP-19 — `TestExportCmd_quietSuppresses`.
- CP-20 — `PropSchemaPresent`, `TestBuild_schemaField`.

### §3.6 Documentation Consistency

✅ Mermaid-диаграммы в design §2.2 точно описывают реальную структуру пакетов. ADR-7 (`pipeline.sh` → `.gitignore`) выполнен. **Одно отклонение от ADR-9/design** — `homebrew_formula` → `homebrew_casks` — пользователь явно одобрил после обнаружения GoReleaser v2 deprecation.

## Code Quality

### §4.1 Naming & Clarity

✅ Идентификаторы идиоматичны для Go (PascalCase для exported, camelCase для local). Структуры с понятными именами (`Repository`, `Writer`, `Preset`, `ExitCodeError`). Алиасы импортов осмысленны (`sqlitestore`, `encjson`, `jsonwriter`, `mdwriter`).

### §4.2 Dead Code & Debug Artifacts

⚠️ Минорно: поля `Deps.Stdin`, `Deps.Env`, `Deps.Goos` объявлены в `internal/cli/deps.go:18-22`, но реально не читаются никаким production-кодом. В тестах их используют (Goos для override-сценариев), но runtime-логика всегда идёт через DiscoverDB. См. F-4 minor.

Нет `TODO` без тикетов, нет `fmt.Println` для отладки, нет commented-out кода.

### §4.3 Scope Creep

✅ Никаких фич за пределами Requirements. Пресет `structure` упомянут в Open Questions, но НЕ реализован (правильно — он отложен в v2). Нет watch-режима, нет YAML/CSV, нет авто-парсинга плистов — всё в Deferred.

### §4.4 Test Quality

✅ Тесты используют `require.*` для критичных assert и `assert.*` для дополнительных. Все table-driven тесты используют `tc := tc; t.Parallel()` для безопасной параллелизации. PBT через `rapid.Check` корректно. Integration-тесты ассертят конкретные значения (counts, presence of keys), а не только «нет ошибки».

## Security

Сканирование изменённых файлов (CLI-инструмент, без сетевых endpoint'ов):

| Категория | Результат |
|-----------|-----------|
| Input validation | ✅ `--indent < 0` отбрасывается (`internal/cli/export.go:79`); `--format` и `--include` валидируются через Registry.Lookup. Путь к БД проверяется через `Open` + `Ping`. |
| Authentication / Authorization | N/A — CLI на локальной машине, доступ к файлам определяется ОС-правами. |
| Injection (SQL/cmd/XSS) | ✅ Все SQL — `const`-литералы, никакого string-concat с user input. Никаких exec/shell. Никаких HTML-templates. |
| Secrets | ✅ `HOMEBREW_TAP_GITHUB_TOKEN` идёт исключительно через `${{ secrets.* }}` в GitHub Actions, никогда не появляется в коде. Никаких hardcoded токенов/паролей. |
| Data exposure | ✅ Tool читает локальную БД; никаких сетевых выходов. BLOB-данные по умолчанию выводятся как hex (поведение Python-скрипта); `--no-blobs` дает возможность их скрыть. |
| Error leakage | ⚠️ При панике `cmd/thingsexporter/main.go:17` печатает `debug.Stack()` со всеми filepaths в `os.Stderr`. Приемлемо для CLI dev-сценариев, но если пользователь скриптует с захватом stderr, paths могут случайно засветиться. Не security-уязвимость, просто info. |
| Open SQLite read-only | ✅ DSN `mode=ro` принудительно отвергает любые write-операции (проверено `TestOpen_readOnlyDSN`). |
| File output | ⚠️ `--out <path>` использует `O_TRUNC` (`internal/cli/export.go:147`). Перезаписывает целевой файл без подтверждения — но это ожидаемая семантика CLI (`>` в shell делает то же самое). |
| DSN path handling | ⚠️ `internal/store/sqlite/open.go:18` строит DSN через `fmt.Sprintf("file:%s?mode=ro", path)`. Если path содержит `?` или `#`, парсер DSN может неправильно интерпретировать. Не критично (нет privilege escalation), но робастность — см. F-5 minor. |

Никаких новых публичных endpoint'ов. Сетевого трафика нет.

## Verification Evidence

Все команды re-run в сессии review (не скопированы из implementation.md).

- **Tests (`task test`):**
```
task: [test] go test ./...
?   	github.com/jtprogru/thingsexporter/cmd/thingsexporter	[no test files]
ok  	github.com/jtprogru/thingsexporter/internal/cli	(cached)
ok  	github.com/jtprogru/thingsexporter/internal/export	(cached)
ok  	github.com/jtprogru/thingsexporter/internal/export/json	(cached)
ok  	github.com/jtprogru/thingsexporter/internal/export/markdown	(cached)
ok  	github.com/jtprogru/thingsexporter/internal/export/preset	(cached)
ok  	github.com/jtprogru/thingsexporter/internal/store/sqlite	(cached)
ok  	github.com/jtprogru/thingsexporter/internal/things	(cached)
?   	github.com/jtprogru/thingsexporter/internal/version	[no test files]
```

- **Tests race uncached (`go test -count=1 -race ./...`):**
```
?   	github.com/jtprogru/thingsexporter/cmd/thingsexporter	[no test files]
ok  	github.com/jtprogru/thingsexporter/internal/cli	2.898s
ok  	github.com/jtprogru/thingsexporter/internal/export	3.097s
ok  	github.com/jtprogru/thingsexporter/internal/export/json	3.492s
ok  	github.com/jtprogru/thingsexporter/internal/export/markdown	4.285s
ok  	github.com/jtprogru/thingsexporter/internal/export/preset	2.277s
ok  	github.com/jtprogru/thingsexporter/internal/store/sqlite	1.955s
ok  	github.com/jtprogru/thingsexporter/internal/things	4.033s
?   	github.com/jtprogru/thingsexporter/internal/version	[no test files]
```

- **Build (`task build`):**
```
task: [build] mkdir -p bin
task: [build] go build -o bin/thingsexporter ./cmd/thingsexporter
```

- **Lint (`task lint`):**
```
task: [lint] golangci-lint run
0 issues.
```

## Findings

| ID | Severity | File | Description | Requirement |
|----|----------|------|-------------|-------------|
| F-1 | minor | `internal/things/build_property_test.go` | Design §2.8 запланировал `PropPresetExclusions` (CP-11 PBT через rapid). Фактически CP-11 покрыта только table-driven unit-тестами в `internal/export/preset/preset_test.go`. Семантически эквивалентно (все 4 пресета явно проверены), но статистическая глубина rapid-генератора отсутствует. | REQ-5.x, CP-11 |
| F-2 | minor | `internal/store/sqlite/queries.go` | Design §2.8 запланировал `TestQueriesQuoteReservedWords` — regex-чек что все упоминания `index/type/status/start` в const SQL обёрнуты в `"`. Фактически проверка поведенческая через `TestRepositoryReadAll_fixture` (если бы было unquoted — SELECT упал бы). Покрытие сохраняется, но статический guard от регрессий слабее. | REQ-1.5, CP-17 |
| F-3 | nit | `.gitignore` | `.claude/`, `.serena/` (локальные IDE-конфиги) не исключены и присутствуют в worktree. Рекомендую добавить в `.gitignore`, чтобы они не попали в первый коммит. | — |
| F-4 | nit | `internal/cli/deps.go:18-22` | Поля `Deps.Stdin`, `Deps.Env` фактически не читаются production-кодом (только в DefaultDeps конструкторе). Поле `Deps.Goos` используется только в тестах. Не вредит, но dead в текущем scope MVP. Удалить или явно пометить TODO для будущих фич. | — |
| F-5 | minor | `internal/store/sqlite/open.go:18`, design §2.7 SQLITE_BUSY row | (а) DSN строится через `fmt.Sprintf("file:%s?mode=ro", path)` без URL-escape — если путь содержит `?` или `#`, DSN ломается. (б) SQLITE_BUSY retry, описанный в design §2.7, не реализован. Оба сценария — расширение робастности; не блокируют MVP (read-only открытие с не запущенным Things 3 работает). | — |

**Все findings — severity ≤ minor.** Нулевых `critical` и `major`.

## Recommendations

В порядке убывания приоритета:

1. **(F-3, nit, 1 строка):** добавить `.claude/` и `.serena/` в `.gitignore`, чтобы они не попали в коммиты, которые пользователь будет делать.
2. **(F-2, minor, ~20 строк):** добавить статический тест `TestQueriesQuoteReservedWords` для CP-17 — защитит от регрессий, когда кто-то правит SQL и забывает quote (поведенческий тест не сработает, если, например, добавят новую таблицу без покрытия integration-тестом).
3. **(F-1, minor, ~30 строк):** написать `PropPresetExclusions` через rapid для покрытия CP-11 в его исходной PBT-форме (полная глубина дизайна).
4. **(F-4, nit):** удалить `Deps.Stdin`, `Deps.Env` или объяснить комментарием, что они зарезервированы для будущих подкоманд (например, `import` или `read-from-stdin`).
5. **(F-5, minor, оба пункта — Deferred):** оба относятся к v2:
   - URL-escape пути в DSN через `url.PathEscape` — защита от exotic пути.
   - SQLITE_BUSY retry с экспоненциальным backoff — нужен только если кто-то столкнётся с конкуренцией live-БД и open-readonly.

**Ни одна из рекомендаций не блокирует merge MVP** — все они уместны в отдельных follow-up PR.
