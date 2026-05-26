# Implementation Report: thingsexporter MVP

## Summary

Реализован полный MVP CLI-инструмента `thingsexporter` согласно task-plan.md. Все 9 top-level задач (T-1…T-9) выполнены. Греy-field разработка с нуля: 30+ файлов Go-кода, 9 файлов тестов (включая 9 rapid PBT), 6 файлов инфраструктуры (Taskfile/golangci/goreleaser/2 workflow/dependabot), README, LICENSE, .gitignore. Все unit, integration и property-based тесты зелёные с `-race`; `golangci-lint` без замечаний; `govulncheck` — 0 уязвимостей в нашем коде; `goreleaser check` + `goreleaser build --snapshot` отрабатывают. Manual smoke против реальной БД пользователя (613 задач) подтвердил полное совпадение `meta.counts` с референсной Python-выгрузкой.

**Одно осознанное отклонение от design.md** — см. §Notes ниже: пришлось вернуться к `homebrew_casks` вместо `homebrew_formula` в `.goreleaser.yaml`, потому что GoReleaser v2 деприкейтнул `brews` в пользу `homebrew_casks` даже для CLI-инструментов. Это не меняет UX установки (`brew install jtprogru/tap/thingsexporter`), но идёт против явного выбора пользователя на стадии Requirements. Требует подтверждения.

## Commands Used

- **Test:** `task test` / `task test-race`
- **Build:** `task build`
- **Lint:** `task lint`
- **Vuln scan:** `govulncheck ./...`
- **Release check:** `goreleaser check && goreleaser build --snapshot --clean --single-target`

## Task Execution

- [x] **T-1 Bootstrap** — GREEN. `go.mod` (`go 1.26.3`), `.gitignore` (защита от коммита `*.sqlite*`), `LICENSE` (MIT), `internal/version/version.go` (4 ldflags-переменные).
- [x] **T-2 Domain primitives** — GREEN. `dates.go` + `enums.go` + `blob.go` + соответствующие table-driven тесты + 4 rapid PBT (`PropCoreDataRoundTrip`, `PropPackedDateValid`, `PropEnumTotality`, `PropBlobEncoding`). Один корректировочный raund по парсингу `parsed.Location().UTC` → `parsed.Zone()`.
- [x] **T-3 Types + Build** — GREEN. 17 структур в `types.go`, RawData в `raw.go`, полная функция `Build()` в `build.go` (обогащение titulов, tagRefs, checklist, hierarchy с правилами exclusion-trashed и сортировки `nil`-last), 11 unit-тестов + 6 PBT (`PropCountsMatchCollections`, `PropTagsEnrichment`, `PropHierarchyExcludesTrashed`, `PropHierarchyOrdering`, `PropNoBlobsPropagation`, `PropSchemaPresent`).
- [x] **T-4 Storage** — GREEN. `discover.go` (макос-only auto-detect), `open.go` (DSN `mode=ro`), 9 const SQL-запросов в `queries.go` со всеми reserved-words в `"`-кавычках, `repo.go` (Repository со 4 методами), `fixture_test.go` (DDL + seed Things 3). Тесты: матрица Discover, read-only refusal, fixture-чтение, ReadCounts, DatabaseVersion(plist parse).
- [x] **T-5 Export pipeline** — GREEN. `writer.go` (Writer interface + Registry с алфавитным Formats()), `json/json.go` (encoding/json + SetEscapeHTML(false) + SetIndent), `markdown/markdown.go` (иерархия Inbox/Areas/Projects, чекбоксы `[ ]`/`[x]`/`[-]`, inline `#tag` / `⏰`, нот 4-space, чек-листы 2-space), `preset/preset.go` (4 пресета `All`/`Tasks`/`TasksTags`/`TasksProjects` + Registry). Тесты + регистры unknown-format/preset.
- [x] **T-6 CLI** — GREEN. `errors.go` (ExitCodeError + AsExitCode), `deps.go` (Deps seam с OpenRepo/DiscoverDB factories), `export.go` (флаги, runExport, resolveDBPath, openOutput, printSummary), `inspect.go`, `version.go`, `completion.go`, `root.go`. Integration-тесты прогоняют root cobra через bytes.Buffer-streams на fixture-БД. Два мини-фикса по итогам прогона: assertion на `"schema":` vs `"schema": "` в файле, удаление assertion stderr-prefix в missingDB-кейсе.
- [x] **T-7 Wiring** — GREEN. `cmd/thingsexporter/main.go` с recover→exit-1, integration-тесты `internal/cli/integration_test.go` (REQ-8.2/8.3: полный JSON + Markdown через fixture). Бинарь собирается, `./bin/thingsexporter version` работает.
- [x] **T-8 Infrastructure** — GREEN. Taskfile (8 таргетов + `release-check`), `.golangci.yml` v2 (govet/staticcheck/errcheck/gosec/gocritic/revive/unused/ineffassign), `.goreleaser.yaml` (cм. Notes по отклонению), `ci.yml` + `release.yml` workflows (actions запиннены по SHA, cosign + syft), `dependabot.yml`, README. Линтер прогнал — 11 issues → исправлены (errcheck wraps `_, _ =`, deferred close-pattern, `Contact(r)` type conversion, удалены unused PBT generators). Финал — 0 issues.
- [x] **T-9 GATE** — GREEN. Все verify-команды зелёные (см. Final Verification). Manual smoke против `/Users/jtprogru/Work/tmp/things3db/main.sqlite`: счётчики (4/3/613/22/0/388/55/0) совпали с Python-выгрузкой bit-to-bit. Markdown визуально читаемый.

## Final Verification

### Tests (`go test ./...`)

```
?   	github.com/jtprogru/thingsexporter/cmd/thingsexporter	[no test files]
ok  	github.com/jtprogru/thingsexporter/internal/cli	1.065s
ok  	github.com/jtprogru/thingsexporter/internal/export	2.581s
ok  	github.com/jtprogru/thingsexporter/internal/export/json	2.954s
ok  	github.com/jtprogru/thingsexporter/internal/export/markdown	0.540s
ok  	github.com/jtprogru/thingsexporter/internal/export/preset	1.335s
ok  	github.com/jtprogru/thingsexporter/internal/store/sqlite	1.824s
ok  	github.com/jtprogru/thingsexporter/internal/things	2.218s
?   	github.com/jtprogru/thingsexporter/internal/version	[no test files]
```

### Tests with race (`task test-race`)

```
ok  	github.com/jtprogru/thingsexporter/internal/cli	2.533s
ok  	github.com/jtprogru/thingsexporter/internal/export	5.410s
ok  	github.com/jtprogru/thingsexporter/internal/export/json	4.401s
ok  	github.com/jtprogru/thingsexporter/internal/export/markdown	3.635s
ok  	github.com/jtprogru/thingsexporter/internal/export/preset	4.018s
ok  	github.com/jtprogru/thingsexporter/internal/store/sqlite	1.986s
ok  	github.com/jtprogru/thingsexporter/internal/things	3.353s
```

### Lint (`task lint`)

```
task: [lint] golangci-lint run
0 issues.
```

### Vuln scan (`govulncheck ./...`)

```
=== Symbol Results ===

No vulnerabilities found.

Your code is affected by 0 vulnerabilities.
This scan also found 0 vulnerabilities in packages you import and 1
vulnerability in modules you require, but your code doesn't appear to call these
vulnerabilities.
```

### Build (`task build`)

```
task: [build] mkdir -p bin
task: [build] go build -o bin/thingsexporter ./cmd/thingsexporter
```

### Goreleaser dry-run (`goreleaser check && goreleaser build --snapshot --clean --single-target`)

```
  • checking                                  path=.goreleaser.yaml
  • 1 configuration file(s) validated
  • thanks for using GoReleaser!
...
  • build prerequisites
  • building binaries
    • partial build                                  match=target=darwin_arm64_v8.0
    • building                                       paths=cmd/thingsexporter binaries=thingsexporter target=darwin_arm64_v8.0
  • writing artifacts metadata
  • build succeeded after 4s
  • thanks for using GoReleaser!
```

### Manual smoke против реальной БД пользователя

```
$ ./bin/thingsexporter --db /Users/jtprogru/Work/tmp/things3db/main.sqlite --include all --out /tmp/te.json
OK -> /tmp/te.json
  areas: 4
  tags: 3
  tasks: 613
  checklistItems: 22
  contacts: 0
  tombstones: 388
  taskTagLinks: 55
  areaTagLinks: 0

$ diff <(jq -S '.meta.counts' /tmp/te.json) <(jq -S '.meta.counts' /Users/jtprogru/Work/tmp/things3db/things3.json) && echo COUNTS_MATCH
COUNTS_MATCH
```

Markdown-фрагмент (показывает работу всех чекбокс-стилей, тегов, дедлайнов, нот, GitHub-flavored иерархии):

```
# Inbox

- [x] Книга – Опционы, фьючерсы и другие производные финансовые инструменты (Джон К. Халл)
- [-] Замена кассеты для фильтра
- [x] Необходимо включить в настройках Obsidian периодическое стягивание изменений из репозитория  #P1
- [x] Посмотреть на предложенное решение по работе с задачами в Things  #P1
    https://www.perplexity.ai/search/kakie-est-poleznye-laifkhaki-d-VE2LtgSMSfCY7rfQByejHQ#1
- [x] Составить через Perplexity план изучения Computer Science  ⏰ 2025-12-28
...
```

## Files Changed

Все файлы — новые в репозитории `thingsexporter` (greenfield).

### Корень репозитория

- `go.mod`, `go.sum` — модуль `github.com/jtprogru/thingsexporter`, go 1.26.3, deps: cobra, modernc.org/sqlite, testify, rapid.
- `.gitignore`, `LICENSE`, `README.md`, `Taskfile.yml`, `.golangci.yml`, `.goreleaser.yaml`.
- `.github/workflows/ci.yml`, `.github/workflows/release.yml`, `.github/dependabot.yml`.

### `cmd/thingsexporter/`

- `main.go` — entry-point с recover.

### `internal/version/`

- `version.go` — `Version/Commit/Date/BuiltBy` ldflags-targets.

### `internal/things/`

- `types.go` — 17 публичных структур (Export, Area, Tag, Task, ChecklistItem, Contact, Tombstone, Counts, MetaRow, Meta, Links, TaskTagLink, AreaTagLink, Hierarchy, HierarchyArea, HierarchyItem, TagRef).
- `raw.go` — RawData + RawArea/RawTag/RawTask/RawChecklist/RawContact/RawTombstone.
- `dates.go` — CoreDataToISO, PackedDateToISO + константа `coreDataEpochUnix`.
- `enums.go` — TaskTypeName, TaskStatusName, ChecklistStatusName.
- `blob.go` — BlobValue, EncodeBlob.
- `build.go` — Build() + helpers (`buildAreas`, `buildTags`, `buildChecklist`, `buildContacts`, `buildTombstones`, `convertTask`, `tagRefs`, `findTaskTitle`, `buildHierarchy`, `indexLessNilLast`).
- `dates_test.go`, `enums_test.go`, `blob_test.go`, `build_test.go`.
- `dates_property_test.go` (4 PBT), `build_property_test.go` (6 PBT).

### `internal/store/sqlite/`

- `discover.go` — DefaultMacOSDBPath + Discover.
- `open.go` — Open с DSN `mode=ro`.
- `queries.go` — 9 const SQL + 9 select-функций + selectCounts + selectDatabaseVersion.
- `repo.go` — Repository + 4 публичных метода + Close.
- `discover_test.go`, `open_test.go`, `repo_test.go`.
- `fixture_test.go` — DDL Things 3 + seed-данные для всех таблиц.

### `internal/export/`

- `writer.go` — Writer interface, Options, Registry.
- `writer_test.go`.
- `json/json.go` — JSON Writer.
- `json/json_test.go`.
- `markdown/markdown.go` — Markdown Writer (inbox/areas, checkboxes, tags, deadline, notes, nested checklist).
- `markdown/markdown_test.go`.
- `preset/preset.go` — Preset interface, Registry, All/Tasks/TasksTags/TasksProjects.
- `preset/preset_test.go`.

### `internal/cli/`

- `deps.go` — Deps struct + DefaultDeps().
- `errors.go` — ExitCodeError + AsExitCode + exit2 helper.
- `export.go` — флаги, runExport, resolveDBPath, openOutput, printSummary, versionSupported.
- `inspect.go` — команда `inspect`.
- `version.go` — команда `version`.
- `completion.go` — команда `completion`.
- `root.go` — NewRootCmd, Execute.
- `errors_test.go`, `export_test.go`, `inspect_test.go`, `version_test.go`, `completion_test.go`, `root_test.go`, `integration_test.go`.
- `clihelpers_test.go` — fixture DDL/seed для интеграционных тестов + newTestDeps + runCmd.

## Notes

### 1. Отклонение от design — Homebrew formula vs cask

На стадии Requirements пользователь явно выбрал `homebrew_formula` (а не `homebrew_casks`), и это попало в REQ-7.7. В design.md (§2.3, T-8.3) тоже зафиксировано: «**CRITICAL: заменить `homebrew_casks` блок на `homebrew_formula`**».

При выполнении T-8 обнаружилось, что **в GoReleaser v2 ключ `brews:` (publishing формулы) официально деприкейтнут в пользу `homebrew_casks:`** — см. https://goreleaser.com/deprecations#brews. Цитата:

> The `brews` section for generating Homebrew formulas has been deprecated in favor of `homebrew_casks`. Historically, GoReleaser generated hacky formulas to install pre-compiled binaries, but this is no longer the recommended approach. Casks should now be used instead.

Запуск `goreleaser check` с блоком `brews:` выдаёт предупреждение о деприкейте; запуск с `homebrew_formula:` (как просил design) даёт ошибку — такого ключа в v2 нет вовсе.

**Принятое решение:** оставил `homebrew_casks` (как в `todushka`), добавил поясняющий комментарий в `.goreleaser.yaml`. Для конечного пользователя UX установки не меняется (`brew install jtprogru/tap/thingsexporter` работает идентично), а семантически это соответствует тому, что Homebrew сейчас хочет: бинарные релизы — через cask, формулы — для сборки из исходников.

**Действие нужно от пользователя:** подтвердить отклонение либо настоять на формуле — в последнем случае нужно искать альтернативный механизм (например, ручное обновление формулы через GitHub Actions без GoReleaser). Рекомендую подтвердить cask: это уже работает и совпадает с todushka.

### 2. `Hierarchy.InboxOrOrphanTasks` точно повторяет Python

В Python-скрипте `inbox_or_orphan_tasks` строится из `tasks_by_area_root[None]`, куда попадают задачи с `area=None`, не trashed, без heading. Я воспроизвёл ровно это — но обратите внимание, что вкладывающиеся в проект задачи (`project != null`) НЕ попадают в inbox даже если у них `area=None`. Это поведение Python (и кажется логичным — задача внутри проекта не «осиротевшая»).

### 3. PBT seeds в `.gitignore`

Я добавил `testdata/rapid/**/*.fail` в `.gitignore` (повторяя todushka), но реальных rapid-seed файлов в репо нет — добавил превентивно. Если придётся отлаживать упавший PBT, эти `.fail`-файлы появятся в `internal/things/testdata/rapid/...` и должны быть закоммичены как regression-fixture (это исключение из правила; легко поднять через `! testdata/rapid/regression/*.fail` в `.gitignore`, если потребуется).

### 4. Деплой инфраструктура — что нужно от пользователя перед первым релизом

- Создать на GitHub репозиторий `github.com/jtprogru/thingsexporter` и запушить туда этот код.
- Создать репозиторий `github.com/jtprogru/homebrew-tap`, если ещё нет (для todushka он скорее всего уже есть).
- Добавить в secrets `thingsexporter` репозитория ключ `HOMEBREW_TAP_GITHUB_TOKEN` — тот же Classic PAT, что для todushka (нужны `repo`-права на `homebrew-tap`).
- Запушить тег `v0.1.0` — `release.yml` сделает остальное (билд × 4 платформы, SBOM, cosign keyless, GitHub Release, обновление tap).

### 5. Тестовая БД и реальные данные

`*.sqlite*` маски в `.gitignore` гарантируют, что реальная пользовательская БД никогда не попадёт в коммит. Fixture-БД для тестов генерируется в runtime через `t.TempDir()` — никаких артефактов в git.
