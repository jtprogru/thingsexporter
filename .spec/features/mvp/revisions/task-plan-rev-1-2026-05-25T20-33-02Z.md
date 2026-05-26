# thingsexporter MVP — Task Plan

**Status:** Draft
**Author:** Mikhail Savin (через AI-ассистента в режиме spec-driven-dev)
**Date:** 2026-05-25
**Inputs:** `.spec/features/mvp/explore.md`, `.spec/features/mvp/requirements.md`, `.spec/features/mvp/design.md`

## Work Type Classification

**Тип работ:** **Pure feature** — greenfield-разработка. Репозиторий `thingsexporter` пуст (нет ни Go-кода, ни тестов, ни инфраструктуры). Поведение, которое нужно «сохранить», отсутствует — preservation-тестов нет, RED-тесты в смысле «зафиксировать дефект» неприменимы. Все тесты пишутся в форме GREEN-стабов (ожидаемое поведение), затем реализация подгоняется под них.

**Task order (по правилам шаблона для Pure feature):**
```
GREEN (тест-стабы) → CODE (реализация снизу-вверх) → GREEN (полные тесты) → GATE
```

В этом плане GREEN/CODE/GREEN-секвенция инлайнится внутри каждой top-level задачи: subtasks внутри одной задачи идут парами «тест → код», по одному файлу за раз.

## Test Style Source

**Test Style Source:** Tier 3 (нет существующих тестов в `thingsexporter`; референс — соседний проект `todushka`)
- Evidence: `find /Users/jtprogru/Work/github/jtprogru/thingsexporter -name '*_test.go'` → ничего. `find /Users/jtprogru/Work/github/jtprogru/todushka -name '*_test.go'` → есть, образцы из тех же доменов (CLI на cobra, storage с in-memory fakes, доменные конвертеры с table-driven, rapid PBT).
- Key patterns:
  - **Framework:** `github.com/stretchr/testify/require` + `assert`.
  - **PBT:** `pgregory.net/rapid` (используется в `todushka/internal/tui/testdata/rapid` и `todushka/internal/config/testdata/rapid`).
  - **Naming:** `TestXxx` для unit, `PropXxx` для rapid PBT.
  - **Structure:** табличные тесты через `t.Run(name, func(t *testing.T){...})`, fixture через `t.TempDir()`, mock-стримов через `bytes.Buffer`.
  - **CLI test seam:** структура `Deps` с инжектируемыми `Stdout/Stderr/Stdin/Env/Clock/...`, в тестах подменяется на буферы и фейк-clock.
  - **Rapid seeds:** локальные `.fail`-файлы идут в `.gitignore` (повторяем todushka).

## Commands

| Action          | Command                                                                  | Source             |
|-----------------|--------------------------------------------------------------------------|--------------------|
| Test            | `task test`                                                              | design.md §2.8     |
| Test-race       | `task test-race`                                                         | design.md §2.8     |
| Build           | `task build`                                                             | design.md §2.8     |
| Cross-compile   | `task cross-compile`                                                     | design.md §2.8     |
| Lint            | `task lint`                                                              | design.md §2.8     |
| Fmt             | `task fmt`                                                               | design.md §2.8     |
| Tidy            | `task tidy`                                                              | design.md §2.8     |
| Vuln scan       | `govulncheck ./...`                                                      | design.md §2.8     |
| Release check   | `goreleaser check && goreleaser build --snapshot --clean --single-target` | design.md §2.8     |

Все subtasks ниже используют ровно эти команды; никаких плейсхолдеров.

## Coverage Matrix

| Requirement | Task(s)       | Correctness Property                         |
|-------------|---------------|----------------------------------------------|
| REQ-1.1     | T-4           | CP-1 (Absence)                               |
| REQ-1.2     | T-4           | CP-15 (Propagation)                          |
| REQ-1.3     | T-4, T-6      | CP-15 + CLI integration test                 |
| REQ-1.4     | T-4, T-6      | Error Handling integration test              |
| REQ-1.5     | T-4           | CP-17 (Absence)                              |
| REQ-1.6     | T-4, T-6      | CLI integration test (warning в stderr)      |
| REQ-2.1     | T-2           | CP-2 (Round-trip)                            |
| REQ-2.2     | T-2           | CP-2                                         |
| REQ-2.3     | T-2           | CP-3 (Equivalence)                           |
| REQ-2.4     | T-2           | CP-4 (Absence)                               |
| REQ-2.5     | T-2           | CP-5 (Equivalence)                           |
| REQ-2.6     | T-2           | CP-5                                         |
| REQ-2.7     | T-2           | CP-6 (Equivalence)                           |
| REQ-2.8     | T-2, T-3      | CP-6, CP-18 (Propagation)                    |
| REQ-3.1     | T-3           | CP-8 (Propagation) + integration             |
| REQ-3.2     | T-3           | CP-8                                         |
| REQ-3.3     | T-3           | Integration test (preset all)                |
| REQ-3.4     | T-3           | Integration test (preset all)                |
| REQ-3.5     | T-3           | CP-9 (Absence), CP-10 (Equivalence)          |
| REQ-3.6     | T-3           | CP-7 (Equivalence)                           |
| REQ-4.1     | T-5           | CP-13 (Equivalence)                          |
| REQ-4.2     | T-5           | CP-13                                        |
| REQ-4.3     | T-5           | CP-14 (Propagation) + integration            |
| REQ-4.4     | T-5           | Integration test (markdown без тегов)        |
| REQ-4.5     | T-5           | CP-12 (Equivalence)                          |
| REQ-5.1     | T-5           | CP-8, CP-11 (Exclusion)                      |
| REQ-5.2     | T-5           | CP-11                                        |
| REQ-5.3     | T-5           | CP-11, CP-8                                  |
| REQ-5.4     | T-5           | CP-11                                        |
| REQ-5.5     | T-5           | CP-12                                        |
| REQ-6.1     | T-6           | CLI integration test                         |
| REQ-6.2     | T-6           | CLI integration test                         |
| REQ-6.3     | T-6           | CLI integration test (stdout vs file)        |
| REQ-6.4     | T-6           | CP-19 (Exclusion) + integration              |
| REQ-6.5     | T-6           | CLI integration test (inspect)               |
| REQ-6.6     | T-6           | Unit test (version output)                   |
| REQ-6.7     | T-6           | Smoke test (cobra completion)                |
| REQ-6.8     | T-6           | Smoke test (--help)                          |
| REQ-6.9     | T-6           | CP-16 (Equivalence)                          |
| REQ-7.1     | T-1, T-8      | Goreleaser-check job                         |
| REQ-7.2     | T-8           | CI workflow run                              |
| REQ-7.3     | T-8           | `task lint` в CI                             |
| REQ-7.4     | T-8           | CI workflow file present + actions pinned    |
| REQ-7.5     | T-8           | Release workflow file present                |
| REQ-7.6     | T-8           | `.goreleaser.yaml` соответствует             |
| REQ-7.7     | T-8           | `.goreleaser.yaml` homebrew_formula блок     |
| REQ-7.8     | T-8           | Документировано в README + manual smoke      |
| REQ-7.9     | T-1, T-8      | Unit test (`version` command output)         |
| REQ-7.10    | T-8           | `.github/dependabot.yml` present             |
| REQ-8.1     | T-2, T-3, T-4 | Все unit-тесты в T-2..T-6 (meta-требование)  |
| REQ-8.2     | T-3, T-7      | Integration JSON test (через fixture)        |
| REQ-8.3     | T-5, T-7      | Integration Markdown test                    |
| REQ-8.4     | T-6           | CLI test через `Deps` seam                   |
| REQ-8.5     | T-1           | `.gitignore` записи                          |
| ADR-9 (CP-20) | T-3, T-5    | CP-20 (Equivalence) — schema field           |

Все 40 REQ покрыты, все 20 CP связаны минимум с одной задачей.

## Implementation Plan

### T-1: Bootstrap репозитория

***Complexity: mechanical***
***Requirements: REQ-7.1, REQ-7.9, REQ-8.5***
GOAL: создать пустой Go-проект с правильным module path, заготовкой пакета `internal/version` и базовым `.gitignore`/`LICENSE`, чтобы остальные задачи могли импортироваться.

**Subtasks:**

- **T-1.1 [CODE]** Создать `go.mod` в корне репозитория.
  CRITICAL: содержимое — ровно две строки: `module github.com/jtprogru/thingsexporter` и `go 1.26.3`. NOTE: `go.sum` появится автоматически после первого `go mod tidy` в T-1.5.
- **T-1.2 [CODE]** Создать `.gitignore` в корне репозитория.
  CRITICAL: добавить ровно эти записи (по одной на строку):
  ```
  bin/
  dist/
  vendor/
  *.test
  *.out
  cover.out
  *.sqlite
  *.sqlite-wal
  *.sqlite-shm
  testdata/rapid/**/*.fail
  /pipeline.sh
  ```
  IMPORTANT: `*.sqlite*`-маски страхуют от случайного коммита пользовательской БД (REQ-8.5). `/pipeline.sh` исключает dev-shim spec-driven-dev (ADR-7).
- **T-1.3 [CODE]** Создать `LICENSE` в корне с текстом MIT-лицензии.
  CRITICAL: заголовок строки `Copyright (c) 2026 Mikhail Savin <jtprogru@gmail.com>`. NOTE: использовать стандартный текст MIT с https://opensource.org/license/mit.
- **T-1.4 [CODE]** Создать `internal/version/version.go`.
  CRITICAL: пакет `version`, объявить четыре package-level переменные `Version = "dev"`, `Commit = ""`, `Date = ""`, `BuiltBy = ""` (все типа `string`). DO NOT добавлять функции — только переменные, чтобы ldflags могли инжектировать значения (REQ-7.9).
- **T-1.5 [VERIFY]** Запустить `task tidy`.
  GOAL: убедиться, что `go.mod` валиден и `go.sum` создан (на данный момент пуст). Команда: `task tidy`. Ожидаемое: exit 0, нет ошибок.
  NOTE: Taskfile.yml ещё не создан в T-1 — этот subtask должен выполниться после T-8.1 (создание Taskfile). Временно, если T-8 ещё не выполнен, использовать прямой `go mod tidy` для bootstrap'а.

### T-2: Доменные примитивы — dates, enums, blob

***Complexity: standard***
***Requirements: REQ-2.1, REQ-2.2, REQ-2.3, REQ-2.4, REQ-2.5, REQ-2.6, REQ-2.7, REQ-2.8, REQ-8.1***
***Preservation: CP-2, CP-3, CP-4, CP-5, CP-6***
GOAL: реализовать чистые функции конвертации Core Data timestamps и packed dates, маппинг enum-кодов в имена, кодирование BLOB. Чистые функции, без I/O — самый низкий уровень доменного слоя.

**Subtasks:**

- **T-2.1 [GREEN]** Написать тест `internal/things/dates_test.go` с двумя функциями.
  CRITICAL: имя пакета — `things_test`. Тесты:
  - `TestCoreDataToISO_table` (Property/2, Feature/dates) — table-driven с минимум этих кейсов: `(nil, nil)`; `(ptr(0.0), ptr("2001-01-01T00:00:00Z"))`; `(ptr(746541716.0), ptr("2024-08-27T...Z"))` — точное ожидание ISO; `(ptr(math.NaN()), nil)`; `(ptr(math.Inf(1)), nil)`. Использовать `require.Equal`.
  - `TestPackedDateToISO_known` (Property/3, Feature/dates) — table: `(nil, nil)`; `(ptr(int64(0)), nil)`; `(ptr(packDate(2024,10,28)), ptr("2024-10-28"))` где `packDate := func(y,m,d int) int64 { return int64(y<<16 | m<<12 | d<<7) }` — определить как локальный helper в тесте.
  - `TestPackedDateToISO_invalid` (Property/4) — table: год 1969 → nil; год 2101 → nil; месяц 0 → nil; месяц 13 → nil; день 0 → nil; день 32 → nil.
  IMPORTANT: после написания запустить `task test ./internal/things/...` — все тесты должны упасть с `undefined: things.CoreDataToISO/PackedDateToISO`. Это ожидаемое RED-состояние для pure feature.
  _Test_Style:_ Tier 3 — следовать table-driven style из `todushka/internal/config/loader_test.go`.
- **T-2.2 [CODE]** Реализовать `internal/things/dates.go`.
  CRITICAL: пакет `things`. Объявить константу `coreDataEpochUnix = float64(978307200)` (равно `time.Date(2001,1,1,0,0,0,0,time.UTC).Unix()`).
  Сигнатуры:
  - `func CoreDataToISO(v *float64) *string` — если `v==nil`, NaN или Inf → nil; иначе сконструировать `time.Unix(0, 0).Add(time.Duration(coreDataEpochUnix+*v)*time.Second).UTC()`. **IMPORTANT:** для точного воспроизведения Python-формата (с микросекундами) используйте `time.Unix(int64(secs), int64(frac*1e9)).UTC()` и форматирование `t.Format("2006-01-02T15:04:05.000000Z07:00")`. Если результат сериализации Python был `+00:00` вместо `Z` — соответствует `time.RFC3339Nano` с UTC; проверить по образцу `things3.json`.
  - `func PackedDateToISO(v *int64) *string` — если `v==nil || *v==0` → nil; иначе `n := *v`, `year := int(n>>16)&0xFFFF`, `month := int(n>>12)&0x0F`, `day := int(n>>7)&0x1F`; если хотя бы одно вне диапазонов REQ-2.4 — nil; иначе `fmt.Sprintf("%04d-%02d-%02d", year, month, day)`.
  После реализации запустить `task test ./internal/things/...` — все тесты GREEN.
- **T-2.3 [GREEN]** Написать `internal/things/enums_test.go`.
  CRITICAL: пакет `things_test`. Тесты: `TestTaskTypeName_known` (Property/5), `TestTaskStatusName_known`, `TestChecklistStatusName_known`. Каждый — table-driven по REQ-2.5/2.6 + кейс «nil pointer» + «unknown code». Запустить — должны упасть.
- **T-2.4 [CODE]** Реализовать `internal/things/enums.go`.
  CRITICAL: три функции:
  - `func TaskTypeName(code *int64) *string` с внутренним `var taskTypes = map[int64]string{0:"todo",1:"project",2:"heading"}`.
  - `func TaskStatusName(code *int64) *string` с `{0:"open",2:"canceled",3:"completed"}`.
  - `func ChecklistStatusName(code *int64) *string` с `{0:"open",3:"completed"}`.
  В каждой: `if code == nil { return nil }; if name, ok := m[*code]; ok { return &name }; return nil`. DO NOT возвращать пустую строку для unknown — строго nil. После — `task test`, GREEN.
- **T-2.5 [GREEN+CODE]** Создать `internal/things/blob_test.go` + `internal/things/blob.go`.
  CRITICAL: в первом подсубтаске `blob_test.go` — `TestEncodeBlob_table` (Property/6, Feature/blob) с кейсами: `(nil, false) → nil`, `([]byte{}, false) → nil`, `([]byte{0xde,0xad}, false) → &BlobValue{Hex: ptr("dead")}`, `([]byte{0xde}, true) → nil`. Затем `blob.go`: `type BlobValue struct { Hex *string \`json:"__blob_hex__,omitempty"\` }`, `func EncodeBlob(b []byte, drop bool) *BlobValue { if drop || len(b)==0 { return nil }; h := hex.EncodeToString(b); return &BlobValue{Hex:&h} }`. Импорт `encoding/hex`.
  NOTE: оба файла за один subtask — допустимо, т.к. blob — самый тривиальный тип (3 строки логики + 1 тип).
- **T-2.6 [GREEN]** Добавить property-based тесты в `internal/things/dates_property_test.go` и `internal/things/enums_property_test.go` через `pgregory.net/rapid`.
  CRITICAL: импорт `pgregory.net/rapid`.
  - `PropCoreDataRoundTrip` (Property/2) — генератор `rapid.Int64Range(978307200, 4102444800)`, ассерт раунд-трип (см. CP-2).
  - `PropPackedDateValid` (Property/3) — `rapid.IntRange(1970,2100)`, `rapid.IntRange(1,12)`, `rapid.IntRange(1,31)`; ассерт точного совпадения формата.
  - `PropEnumTotality` (Property/5) — известные коды + `rapid.Int64Range(-100, 100).Filter(notInKnownSet)`.
  - `PropBlobEncoding` (Property/6) — `rapid.SliceOfN(rapid.Byte(), 0, 256)` + `rapid.Bool()`.
  Запустить `task test ./internal/things/...`, ожидаемо GREEN.

### T-3: Доменные типы и Build (обогащение + hierarchy)

***Complexity: standard***
***Requirements: REQ-2.8, REQ-3.1, REQ-3.2, REQ-3.3, REQ-3.4, REQ-3.5, REQ-3.6, REQ-8.1, REQ-8.2, ADR-9***
***Preservation: CP-2, CP-3, CP-4, CP-5, CP-6, CP-7, CP-8, CP-9, CP-10, CP-18, CP-20***
GOAL: ввести типы `Export/Area/Tag/Task/...`, `RawData` и функцию `Build(raw, opts) Export`, которая выполняет всё обогащение задач (areaTitle/projectTitle/tags/checklist), сборку Hierarchy и Counts, проброс `--no-blobs`. Это самая логически плотная функция MVP.

**Subtasks:**

- **T-3.1 [CODE]** Создать `internal/things/types.go` со всеми публичными структурами из §2.5 design-документа.
  CRITICAL: все 17 типов с точными JSON-тегами как в design. DO NOT добавлять методы — только поля. NOTE: импорт `encoding/hex` не нужен здесь; BLOB уже определён в `blob.go`. Запустить `task tidy && go build ./internal/things/...` — должно собраться.
- **T-3.2 [CODE]** Создать `internal/things/raw.go` с приватными raw-DTO типами и публичным `RawData`.
  CRITICAL: типы `rawArea`, `rawTag`, `rawTask`, `rawChecklist`, `rawContact`, `rawTombstone` — содержат сырые поля как они приходят из `database/sql`: `*float64` для CoreData timestamps, `*int64` для packed dates и enum-кодов, `[]byte` для BLOB, `*string` для nullable строк. `RawData` — публичная struct с полями `Areas []rawArea`, `Tags []rawTag`, `Tasks []rawTask`, `Checklist []rawChecklist`, `Contacts []rawContact`, `Tombstones []rawTombstone`, `TaskTagPairs []TaskTagLink`, `AreaTagPairs []AreaTagLink`, `MetaRows []MetaRow`. IMPORTANT: `raw*`-типы — package-private (lowercase), `RawData` — публичная.
- **T-3.3 [GREEN]** Создать `internal/things/build_test.go` с базовыми табличными тестами.
  CRITICAL: пакет `things_test`. Helpers: `func newRawData() RawData` с фиксированным набором (2 areas, 3 tags, 5 tasks включая trashed + один с tags, 2 checklist, 1 contact, 1 tombstone, 3 TaskTagPairs, 1 AreaTagPair). Тесты:
  - `TestBuild_enrichTaskTags` (Property/8) — два задачи с разными наборами тегов; ассертить, что `task.Tags` содержит точные `TagRef{UUID, Title}`.
  - `TestBuild_areaProjectHeadingTitles` (Feature/build) — задача с заполненными `Area`/`Project`/`Heading` → результат содержит соответствующие `*Title` поля.
  - `TestBuild_hierarchy_excludesTrashed` (Property/9) — задача `Trashed=1` в области A → `result.Hierarchy.Areas[A].Items` не содержит её UUID.
  - `TestBuild_hierarchy_ordering` (Property/10) — три области с indexes `-100, nil, 50` → порядок `-100, 50, nil`.
  - `TestBuild_counts_match` (Property/7) — после Build для preset-агностичного начала: `Counts` поля соответствуют длинам коллекций.
  - `TestBuild_schemaField` (Property/20) — `result.Schema == "thingsexporter/v1"`.
  - `TestBuild_noBlobs_strips` (Property/18) — задача с непустым `CachedTags []byte{0xff}`; `BuildOptions{NoBlobs:true}` → `task.CachedTags == nil`.
  IMPORTANT: на этом этапе все тесты упадут с `undefined: things.Build`.
- **T-3.4 [CODE]** Реализовать `internal/things/build.go`.
  CRITICAL: пакет `things`. Сигнатура `func Build(raw RawData, opts BuildOptions) Export`, где `BuildOptions struct { Source string; ExportedAt time.Time; NoBlobs bool }`.
  Алгоритм (без выдумок):
  1. Построить index-карты: `tagByUUID map[string]*Tag`, `areaByUUID map[string]*Area`, `contactByUUID map[string]*Contact`, `taskByUUID map[string]*Task` (после первичной конвертации raw → domain).
  2. Сконвертировать все raw-сущности в domain-сущности через `dates.go`/`enums.go`/`blob.go` (с пробросом `opts.NoBlobs` в `EncodeBlob`).
  3. Построить `tagsForTask map[string][]string` и `tagsForArea map[string][]string` из `TaskTagPairs`/`AreaTagPairs`.
  4. Построить `checklistForTask map[string][]ChecklistItem`, отсортированный по `Index` ASC с `nil` в конце.
  5. Обогатить задачи: `AreaTitle/ProjectTitle/HeadingTitle/ContactName` через карты, `Tags` через `tagsForTask`+`tagRefs`, `Checklist` через `checklistForTask`. Tags теги: для каждого UUID — `TagRef{UUID, Title: tagByUUID[uuid].Title}` (nil-safe).
  6. Обогатить области: `Tags` через `tagsForArea`+`tagRefs`.
  7. Построить Hierarchy: отсортировать areas по Index ASC nil-last; для каждой собрать items (только `Trashed==nil||*Trashed==0`, `Project==nil`, `Heading==nil`, `Area==&areaUUID`); отдельно inbox = items с `Area==nil`. Сортировка items — по Index ASC nil-last.
  8. Заполнить Counts: указатели только для коллекций, которые присутствуют в полном Build (на данный момент — все).
  9. Заполнить `Export.Schema = "thingsexporter/v1"`.
  IMPORTANT: каждое поле task/area, которое в Python было обогащено — должно быть обогащено и здесь, чтобы integration test проходил против fixture.
  После реализации — `task test ./internal/things/...`, все GREEN.
- **T-3.5 [GREEN]** Добавить property-based тесты `internal/things/build_property_test.go`.
  CRITICAL: PBT по CP-7, CP-8, CP-9, CP-10, CP-11 (отложим в T-5, т.к. там presets), CP-18, CP-20.
  - `PropCountsMatchCollections` (Property/7) — rapid-генератор `genRawData` (см. ниже).
  - `PropTagsEnrichment` (Property/8) — генератор RawData с N задач, M тегов, P TaskTagPairs.
  - `PropHierarchyExcludesTrashed` (Property/9) — задачи с random `Trashed`.
  - `PropHierarchyOrdering` (Property/10) — random `Index` ∈ `nil ∪ int64`.
  - `PropNoBlobsPropagation` (Property/18) — random BLOB-байты, `NoBlobs=true`.
  - `PropSchemaPresent` (Property/20).
  IMPORTANT: вынести `genRawData` как rapid-генератор в этом же файле (test-only). DO NOT использовать его в продакшен-коде.

### T-4: Storage слой — open, queries, repo, fixture

***Complexity: complex***
***Requirements: REQ-1.1, REQ-1.2, REQ-1.3, REQ-1.4, REQ-1.5, REQ-1.6, REQ-8.1***
***Preservation: CP-1, CP-15, CP-17***
GOAL: реализовать чтение Things SQLite в режиме `mode=ro`, авто-определение пути для macOS, генерацию fixture-БД для всех последующих тестов. Самый I/O-плотный слой.

**Subtasks:**

- **T-4.1 [GREEN]** Создать `internal/store/sqlite/discover_test.go`.
  CRITICAL: пакет `sqlite_test`. Тесты:
  - `TestDiscover_matrix` (Property/15, Feature/discover) — табличный с матрицей `(home string, goos string, statFn func(string) error) → (path string, ok bool)`:
    - `("/u", "darwin", okFn) → ("/u/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite", true)`
    - `("/u", "darwin", errFn) → ("", false)`
    - `("/u", "linux", okFn) → ("", false)`
    - `("", "darwin", okFn) → ("", false)`
  `okFn := func(_ string) error { return nil }`, `errFn := func(_ string) error { return os.ErrNotExist }`.
- **T-4.2 [CODE]** Реализовать `internal/store/sqlite/discover.go`.
  CRITICAL: пакет `sqlite`. Константа `const DefaultMacOSDBPath = "Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite"`. Функция `func Discover(home string, goos string, statFn func(string) error) (string, bool)` с поведением:
  ```
  if home == "" || goos != "darwin" { return "", false }
  p := filepath.Join(home, DefaultMacOSDBPath)
  if err := statFn(p); err != nil { return "", false }
  return p, true
  ```
  Запустить `task test ./internal/store/sqlite/...`, GREEN.
- **T-4.3 [GREEN+CODE]** Создать `internal/store/sqlite/open_test.go` + `internal/store/sqlite/open.go`.
  CRITICAL: тест:
  - `TestOpen_readOnlyDSN` (Property/1, Feature/storage) — создать fixture через `BuildFixture(t)` (см. T-4.6, временно использовать `t.TempDir()` + ручной `os.Create`), вызвать `Open(path)`, выполнить `db.PingContext(ctx)`, ассертить успех. Затем попытаться `db.ExecContext(ctx, "INSERT INTO ...")` — ассертить ошибку `attempt to write a readonly database`. Закрыть DB.
  - `TestOpen_missingFile` — `Open("/nonexistent/path.sqlite")` → error содержит `no such file`.
  Реализация `open.go`:
  ```
  import (_ "modernc.org/sqlite"; "database/sql"; "fmt")
  func Open(path string) (*sql.DB, error) {
      if path == "" { return nil, fmt.Errorf("empty db path") }
      dsn := fmt.Sprintf("file:%s?mode=ro", path)
      return sql.Open("sqlite", dsn)
  }
  ```
  IMPORTANT: импорт драйвера через blank import — обязателен. NOTE: `sql.Open` ленив; реальная ошибка для несуществующего файла приходит на `db.PingContext`. Тест должен вызывать `Ping` после `Open`.
- **T-4.4 [GREEN]** Создать `internal/store/sqlite/queries_test.go`.
  CRITICAL: тест:
  - `TestQueriesQuoteReservedWords` (Property/17, Feature/storage) — прочитать содержимое `queries.go` через `os.ReadFile`, для каждого слова `w ∈ {"index","type","status","start"}` запустить `regexp.MustCompile(\`[^"]\b\` + w + \`\b[^"]\`).FindAllString(content, -1)` — ассертить, что **нет** совпадений (т.е. каждое упоминание окружено двойными кавычками).
  - `TestSelectAreas_fixture`, `TestSelectTags_fixture`, `TestSelectTasks_fixture`, `TestSelectChecklist_fixture`, `TestSelectContacts_fixture`, `TestSelectTombstones_fixture`, `TestSelectTaskTags_fixture`, `TestSelectAreaTags_fixture`, `TestSelectMetaRows_fixture` — каждый создаёт fixture, вызывает соответствующую selectXxx-функцию, ассертит длину и хотя бы одно контрольное значение.
  NOTE: на момент написания этого теста `BuildFixture` ещё не существует — temporarily использовать встроенный setup-DDL в helper'е `setupTestDB(t *testing.T) string` внутри `queries_test.go`. После T-4.6 заменить на `BuildFixture(t)`.
- **T-4.5 [CODE]** Реализовать `internal/store/sqlite/queries.go`.
  CRITICAL: все SELECT-ы как `const`-строки, идентификаторы в двойных кавычках:
  ```
  const selectTasksSQL = `SELECT "uuid", "leavesTombstone", "creationDate", ..., "index", ..., "type", ..., "status", ..., "start", ..., FROM "TMTask"`
  ```
  Функции:
  - `func selectAreas(ctx context.Context, db *sql.DB) ([]rawArea, error)` — выполняет `selectAreasSQL`, итерирует `rows.Scan(...)`. **CRITICAL:** для каждой nullable-колонки использовать `sql.NullString/NullFloat64/NullInt64`, потом конвертировать в `*string/*float64/*int64`. Для BLOB — `[]byte` напрямую.
  - Аналогично: `selectTags`, `selectTasks`, `selectChecklist`, `selectContacts`, `selectTombstones`, `selectTaskTags` (возвращает `[]things.TaskTagLink`), `selectAreaTags` (`[]things.AreaTagLink`), `selectMetaRows` (`[]things.MetaRow`), `selectCounts` (выполняет 8 `SELECT COUNT(*) FROM <T>`), `selectDatabaseVersion` (читает `SELECT value FROM Meta WHERE key='databaseVersion'`, парсит substring `<integer>N</integer>` через regex, возвращает `*int` или nil).
  IMPORTANT: импорт `internal/things` (для `RawData`, `TaskTagLink`, `AreaTagLink`, `MetaRow`, `rawArea` и т.д. — последние перенести из package `things` в package `sqlite`? — **нет**, оставить `raw*` в `things`, делать SELECT прямо в `things.raw*`. Это требует, чтобы `raw*` стали публичными или sqlite-пакет был в том же модуле и видел их). **Решение:** в T-3.2 уже сделать `raw*` публичными (`RawArea`, `RawTask`...) для cross-package доступа. Перепроверить T-3.2 — если приватные, T-4.5 потребует исправления.
  NOTE: пересмотр T-3.2 — сделать `raw*` публичными как `RawArea/RawTag/...` (зачёркивает предыдущую формулировку «package-private»). **CRITICAL:** при выполнении T-3.2 — следовать этой поправке.
- **T-4.6 [CODE]** Реализовать `internal/store/sqlite/fixture.go` (build-tag `_test.go` или внутри `_test.go`-файла).
  CRITICAL: файл `internal/store/sqlite/fixture_test.go` (чтобы попадал только в test-build). Функция `func BuildFixture(t testing.TB) string`:
  1. `path := filepath.Join(t.TempDir(), "things.sqlite")`.
  2. `db, _ := sql.Open("sqlite", "file:"+path+"?mode=rwc")`.
  3. Применить DDL ровно из `schema.sql` (создать константу `fixtureDDL` со всеми `CREATE TABLE` из реальной БД Things 3, скопированных из `sqlite3 main.sqlite .schema` вывода в explore-документе).
  4. Вставить контролируемый набор: 2 area, 3 tag (один с parent), 5 task (один trashed, один с heading, один в проекте, один сирота, один в области), 1 checklist, 1 contact, 1 tombstone, 3 TaskTagPairs, 1 AreaTagPair, Meta строка с `databaseVersion=26` в plist-формате.
  5. Сохранить и закрыть. Вернуть `path`.
  IMPORTANT: timestamps в fixture использовать **корректные** Core Data значения (например, `724000000.0`) — это потом проверяется в тестах конверсии.
  NOTE: после реализации, переделать helper из T-4.4 на `BuildFixture(t)` (через `defer` cleanup `t.TempDir` сам сделает).
- **T-4.7 [GREEN+CODE]** Создать `internal/store/sqlite/repo_test.go` + `internal/store/sqlite/repo.go`.
  CRITICAL: тесты:
  - `TestRepositoryReadAll_fixture` (Feature/storage) — `Open` → `NewRepository(db)` → `r.ReadAll(ctx)` → ассертить длины коллекций ровно как в fixture (2/3/5/1/1/1/3/1).
  - `TestRepositoryReadCounts_fixture` — `r.ReadCounts(ctx)` возвращает корректные счётчики.
  - `TestRepositoryDatabaseVersion_meta` — `r.DatabaseVersion(ctx) == ptr(26)`.
  Реализация `repo.go`:
  ```
  type Repository struct { db *sql.DB }
  func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }
  func (r *Repository) ReadAll(ctx context.Context) (things.RawData, error) {
      // параллельно или последовательно вызвать все selectXxx
      // собрать в RawData
  }
  func (r *Repository) ReadCounts(ctx context.Context) (things.Counts, error) { ... }
  func (r *Repository) DatabaseVersion(ctx context.Context) (*int, error) { ... }
  func (r *Repository) Close() error { return r.db.Close() }
  ```
  После — `task test ./internal/store/sqlite/...`, GREEN.

### T-5: Export pipeline — Writer + Preset + JSON + Markdown

***Complexity: standard***
***Requirements: REQ-4.1..REQ-4.5, REQ-5.1..REQ-5.5, REQ-8.3, ADR-9***
***Preservation: CP-7, CP-8, CP-11, CP-12, CP-13, CP-14, CP-20***
GOAL: ввести интерфейсы `Writer`/`Preset`, реестры, две реализации форматов и четыре пресета.

**Subtasks:**

- **T-5.1 [CODE]** Создать `internal/export/writer.go`.
  CRITICAL: пакет `export`. Определения:
  ```
  type Options struct { Indent int }
  type Writer interface {
      Format() string
      Write(out io.Writer, data things.Export, opts Options) error
  }
  type Registry struct { writers map[string]Writer }
  func NewRegistry(ws ...Writer) *Registry { ... // регистрирует все }
  func (r *Registry) Register(w Writer)
  func (r *Registry) Lookup(format string) (Writer, error) // err: "unknown format %q (supported: %s)"
  func (r *Registry) Formats() []string // отсортированный список
  ```
  IMPORTANT: формат сообщения ошибки — ровно `unknown format "<value>" (supported: json, markdown)` (REQ-4.5). Список `supported` строить через `sort.Strings(r.Formats())`.
- **T-5.2 [GREEN+CODE]** Создать `internal/export/writer_test.go`.
  CRITICAL: тесты:
  - `TestRegistryLookup_unknownFormat` (Property/12) — `r.Lookup("yaml")` → `err != nil && strings.Contains(err.Error(), "unknown format")` && `strings.Contains(err.Error(), "yaml")`.
  - `TestRegistry_FormatsSorted` — после регистрации mock-writers, `Formats()` возвращает алфавитно отсортированный список.
  Создать mock-writer в `_test.go` (`type fakeWriter struct{ name string }` с `Format()` и `Write()` возвращающим nil).
- **T-5.3 [GREEN+CODE]** Создать `internal/export/json/json.go` + `internal/export/json/json_test.go`.
  CRITICAL: пакет `json`. Тесты:
  - `TestJsonWriter_indent_compact` (Property/13) — `opts.Indent==0` → результат не содержит `'\n'`.
  - `TestJsonWriter_indent_two` — `opts.Indent==2` → содержит `"  "` indent на втором уровне; парсится обратно `json.Unmarshal` корректно.
  - `TestJsonWriter_noASCIIEscape` — Export с title `"Тест"` → bytes content содержат UTF-8 `Тест`, нет `Тест`.
  - `TestJsonWriter_format_returns_json` — `(&Writer{}).Format() == "json"`.
  Реализация:
  ```
  type Writer struct{}
  func (Writer) Format() string { return "json" }
  func (Writer) Write(out io.Writer, data things.Export, opts export.Options) error {
      enc := json.NewEncoder(out)
      enc.SetEscapeHTML(false)
      if opts.Indent > 0 { enc.SetIndent("", strings.Repeat(" ", opts.Indent)) }
      return enc.Encode(data)
  }
  ```
  NOTE: package `json` имеет конфликт имён с stdlib `encoding/json`. Использовать alias: `encjson "encoding/json"`.
- **T-5.4 [GREEN+CODE]** Создать `internal/export/markdown/markdown.go` + `internal/export/markdown/markdown_test.go`.
  CRITICAL: пакет `markdown`. Тесты:
  - `TestMarkdownWriter_inboxAreas` — fixture с 1 area + 1 inbox task → вывод содержит `# Inbox` и `# Areas`.
  - `TestMarkdownWriter_checkboxes` (Property/14) — задачи со статусами open/completed/canceled → строки `- [ ]` / `- [x]` / `- [-]`.
  - `TestMarkdownWriter_tagsAndDeadline` — задача с тегами и deadline → строка содержит `#tag1 #tag2 ⏰ 2024-10-28`.
  - `TestMarkdownWriter_notesIndent` — задача с notes `"hello\nworld"` → блок:
    ```
    - [ ] title
        hello
        world
    ```
  - `TestMarkdownWriter_checklistNested` — задача с чек-листом → вложенный `  - [ ] item`.
  - `TestMarkdownWriter_format_returns_markdown` — `(&Writer{}).Format() == "markdown"`.
  Реализация: один пакет `markdown` с `type Writer struct{}` и методом `Write` по правилам REQ-4.3 + ADR-8. Использовать `bufio.Writer` для эффективной записи. Алгоритм:
  1. Если `data.Hierarchy != nil`:
     a. Печатать `# Inbox\n\n` + для каждой `InboxOrOrphanTasks[i]` — `renderTaskLine(item)`.
     b. `\n# Areas\n\n` + для каждой `Hierarchy.Areas[i]` — `## <title>\n` + items в правильном порядке (включая под-проекты — для них `### <title>` потом задачи проекта из `data.Tasks`).
  2. Если `Hierarchy == nil` (пресеты tasks/tasks+tags/tasks+projects) — печатать просто плоский список задач из `data.Tasks`.
  3. `renderTaskLine(t)` — `<checkbox> <title><tags-suffix><deadline-suffix>\n` + (notes 4-space indent) + (checklist 2-space nested).
  IMPORTANT: маркер canceled — ровно `[-]` (с пробелом перед скобкой не нужно — обычный markdown task list).
- **T-5.5 [GREEN+CODE]** Создать `internal/export/preset/preset.go` + `internal/export/preset/preset_test.go`.
  CRITICAL: пакет `preset`. Определения:
  ```
  type Preset interface {
      Name() string
      Apply(in things.Export) things.Export
  }
  type Registry struct { presets map[string]Preset }
  func NewRegistry(ps ...Preset) *Registry
  func (r *Registry) Register(p Preset)
  func (r *Registry) Lookup(name string) (Preset, error)
  func (r *Registry) Names() []string
  ```
  Реализации (по одной типу-структуре в файле, либо в одном — выбор: один файл `presets.go`):
  - `presetAll struct{}` — `Apply(in) in` (identity).
  - `presetTasks struct{}` — оставить только Tasks, у каждой задачи занулить `Tags/Checklist/AreaTitle/ProjectTitle/HeadingTitle/ContactName`. Counts: только `Tasks`. Hierarchy/Links — nil. Other collections — nil.
  - `presetTasksTags struct{}` — оставить Tasks (с Tags), Tags коллекция. Counts: `Tasks`, `Tags`.
  - `presetTasksProjects struct{}` — оставить Areas + Tasks (с AreaTitle/ProjectTitle/HeadingTitle/ContactName, без Tags/Checklist). Counts: `Areas`, `Tasks`.
  Сообщение об ошибке для Lookup: `unknown include preset "<value>" (supported: all, tasks, tasks+projects, tasks+tags)` (алфавитная сортировка).
  Тесты:
  - `TestPresetTasks_strips` (Property/11) — Build → presetTasks.Apply → ассертить exclusion-инварианты.
  - `TestPresetTasksTags_strips` — ассертить, что Tags остаются + Areas/Checklist/Links/Hierarchy nil.
  - `TestPresetTasksProjects_strips` — ассертить Areas/Tasks (с титулами) + остальные nil.
  - `TestPresetAll_identity` — Apply == identity.
  - `TestPresetRegistryLookup_unknown` — error содержит `unknown include preset` и `"foo"`.
- **T-5.6 [GREEN]** Добавить PBT в `internal/export/preset/preset_property_test.go` и `internal/export/json/json_property_test.go`.
  CRITICAL:
  - `PropPresetExclusions` (Property/11) — rapid-генератор RawData (импортировать из `things` или дублировать) → Build → перебрать все 4 presets → проверка exclusion.
  - `PropFormatLookup` (Property/12) — random string как format-arg.
  - `PropJsonIndentLayout` (Property/13) — small Export + indent ∈ {0..8}.
  - `PropMarkdownCheckboxes` (Property/14) — random StatusName → правильный маркер.
  - `PropSchemaPresent` (Property/20) — Build → JSON → парсинг → ключ `"schema"` присутствует.

### T-6: CLI слой — root, deps, export, inspect, version, completion, errors

***Complexity: complex***
***Requirements: REQ-1.3, REQ-1.4, REQ-1.6, REQ-6.1..REQ-6.9, REQ-8.4***
***Preservation: CP-15, CP-16, CP-19***
GOAL: Cobra-команды с инжектируемыми зависимостями (`Deps`), маппинг ошибок в exit-коды, integration-тесты CLI-поверхности через buffer-streams.

**Subtasks:**

- **T-6.1 [CODE]** Создать `internal/cli/deps.go`.
  CRITICAL: пакет `cli`. Структура:
  ```
  type Deps struct {
      Stdout, Stderr io.Writer
      Stdin          io.Reader
      Env            func(string) string
      Goos           string                                                    // обычно runtime.GOOS
      Clock          func() time.Time
      OpenRepo       func(path string) (*sqlitestore.Repository, error)        // sqlitestore = alias
      DiscoverDB     func() (string, bool)
      Writers        *export.Registry
      Presets        *presetpkg.Registry
  }
  func DefaultDeps() Deps {
      // регистрирует jsonwriter.Writer{}, mdwriter.Writer{}
      // регистрирует presetAll{}, presetTasks{}, presetTasksTags{}, presetTasksProjects{}
      // OpenRepo := func(p) (*sqlitestore.Repository, error) {
      //     db, err := sqlitestore.Open(p); if err != nil { return nil, err }
      //     return sqlitestore.NewRepository(db), nil
      // }
      // DiscoverDB := func() (string, bool) {
      //     home, _ := os.UserHomeDir()
      //     return sqlitestore.Discover(home, runtime.GOOS, func(p string) error { _, err := os.Stat(p); return err })
      // }
      // Clock := time.Now
  }
  ```
  IMPORTANT: используется alias-импорт `sqlitestore "github.com/jtprogru/thingsexporter/internal/store/sqlite"` чтобы избежать конфликта с `database/sql`.
- **T-6.2 [CODE]** Создать `internal/cli/errors.go`.
  CRITICAL: типы и функция:
  ```
  type ExitCodeError struct { Code int; Err error }
  func (e *ExitCodeError) Error() string { return e.Err.Error() }
  func (e *ExitCodeError) Unwrap() error { return e.Err }
  func AsExitCode(err error) int {
      if err == nil { return 0 }
      var ec *ExitCodeError
      if errors.As(err, &ec) { return ec.Code }
      return 2 // default для CLI/IO ошибок
  }
  ```
  NOTE: код 1 — только при панике в `main` (обрабатывается там через `recover`).
- **T-6.3 [GREEN]** Создать `internal/cli/errors_test.go`.
  CRITICAL: `TestAsExitCode_table` (Property/16, Feature/cli) — table-driven:
  ```
  (nil, 0)
  (errors.New("x"), 2)
  (&ExitCodeError{Code: 2, Err: io.EOF}, 2)
  (&ExitCodeError{Code: 0, Err: io.EOF}, 0)
  (fmt.Errorf("wrapped: %w", &ExitCodeError{Code:2, Err: io.EOF}), 2)
  ```
  Дополнительно: `PropExitCodes` (rapid) в `errors_property_test.go`.
- **T-6.4 [CODE]** Создать `internal/cli/export.go`.
  CRITICAL: функция `newExportCmd(deps Deps) *cobra.Command`. Флаги:
  ```
  --db        string   default ""    -> resolveDBPath(deps, db)
  --out       string   default "-"
  --format    string   default "json"
  --include   string   default "all"
  --indent    int      default 2
  --no-blobs  bool     default false
  --quiet     bool     default false
  ```
  `RunE`: вызывает приватный `runExport(ctx context.Context, deps Deps, opts exportOpts) error`. Алгоритм `runExport`:
  1. Резолв пути: `--db` приоритетнее `Discover`. Если оба пусты — вернуть `&ExitCodeError{Code:2, Err: errors.New("error: --db is required (no Things 3 database found at default path)")}`. Если пришёл через discover и не `--quiet` — `fmt.Fprintln(deps.Stderr, "using DB:", path)`.
  2. Lookup format/preset через registries — при ошибке оборачивать в `&ExitCodeError{Code:2,...}`.
  3. Lookup `--indent` — если < 0 → ошибка.
  4. `OpenRepo(path)` → при ошибке оборачивать в exit-2.
  5. `repo.DatabaseVersion(ctx)` → если значение и не равно одному из supported (`{26}`) и не `--quiet` → `fmt.Fprintf(deps.Stderr, "warning: unsupported Things 3 databaseVersion=%d, output may be incomplete\n", *v)`.
  6. `repo.ReadAll(ctx)` → ошибка → exit-2.
  7. `Build(raw, BuildOptions{Source: path, ExportedAt: deps.Clock().UTC(), NoBlobs: --no-blobs})`.
  8. `preset.Apply(export)`.
  9. Открыть `--out`: `-` → `deps.Stdout`, иначе `os.OpenFile(path, O_WRONLY|O_CREATE|O_TRUNC, 0644)` (defer Close). Ошибка → exit-2.
  10. `writer.Write(out, exportData, export.Options{Indent: indent})` → ошибка → exit-2.
  11. Если не `--quiet` — печатать в stderr отчёт `OK -> <path>\n  <counts...>`.
- **T-6.5 [GREEN]** Создать `internal/cli/export_test.go`.
  CRITICAL: тесты — все вызывают через `Execute(deps)` с buffer-streams:
  - `TestExportCmd_rootDefaults` — root без аргументов → JSON в stdout, parsed валидный, содержит `schema`.
  - `TestExportCmd_format_markdown` — `--format markdown` → stdout начинается с `# Inbox` или `# Areas`.
  - `TestExportCmd_includeTasks_strips` — `--include tasks` → парсинг JSON → нет ключа `areas`.
  - `TestExportCmd_unknownFormat_exit2` — `--format yaml` → exit code 2 (через `AsExitCode(Execute(deps))`).
  - `TestExportCmd_missingDBNonMac` (Property/15) — `Deps.Goos="linux"`, `DiscoverDB` возвращает `("", false)` → exit 2, stderr содержит `--db is required`.
  - `TestExportCmd_quietSuppresses` (Property/19) — `--quiet` → stderr пустой при успехе.
  - `TestExportCmd_outToFile` — `--out /tmp/t.json` → файл создан, valid JSON внутри.
  - `TestExportCmd_indentZero` — `--indent 0` → stdout без `\n`.
  - `TestExportCmd_noBlobs` (Property/18) — `--no-blobs` → парсинг JSON → BLOB-поля `null`.
  IMPORTANT: использовать fake `OpenRepo` в Deps, который возвращает Repository, открытый на fixture-БД (T-4.6 BuildFixture).
- **T-6.6 [GREEN+CODE]** Создать `internal/cli/inspect.go` + `internal/cli/inspect_test.go`.
  CRITICAL: команда `inspect` с флагами `--db`, `--quiet`. `RunE`: открывает Repository, вызывает `ReadCounts` + `DatabaseVersion`, печатает в stdout JSON:
  ```
  {"path":"<resolved>","databaseVersion":26,"counts":{"areas":N,...}}
  ```
  Тесты:
  - `TestInspectCmd_outputsCounts` — fixture → stdout — валидный JSON с теми же ключами.
  - `TestInspectCmd_databaseVersionWarning` — fixture с `databaseVersion=99` → stderr содержит `warning: unsupported`.
- **T-6.7 [GREEN+CODE]** Создать `internal/cli/version.go` + `internal/cli/version_test.go`.
  CRITICAL: команда `version`. `RunE` печатает в `deps.Stdout`:
  ```
  thingsexporter <version.Version>
    commit:    <orDash(version.Commit)>
    built:     <orDash(version.Date)>
    built by:  <orDash(version.BuiltBy)>
    go:        <runtime.Version()>
    platform:  <runtime.GOOS>/<runtime.GOARCH>
  ```
  Helper: `func orDash(s string) string { if s == "" { return "-" }; return s }`. (REQ-6.6).
  Тест `TestVersionCmd_outputFormat` — regex match на стартовой строке `^thingsexporter \S+\n  commit:\s+\S+`.
- **T-6.8 [CODE]** Создать `internal/cli/completion.go`.
  CRITICAL: стандартный cobra-генератор:
  ```
  return &cobra.Command{
      Use: "completion <bash|zsh|fish|powershell>",
      Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
      ValidArgs: []string{"bash","zsh","fish","powershell"},
      RunE: func(cmd *cobra.Command, args []string) error {
          switch args[0] {
          case "bash": return cmd.Root().GenBashCompletion(deps.Stdout)
          case "zsh": return cmd.Root().GenZshCompletion(deps.Stdout)
          case "fish": return cmd.Root().GenFishCompletion(deps.Stdout, true)
          case "powershell": return cmd.Root().GenPowerShellCompletion(deps.Stdout)
          }
          return nil
      },
  }
  ```
  Smoke-тест `TestCompletionCmd_bash` — вызвать с `bash` arg, проверить exit 0, stdout не пуст.
- **T-6.9 [CODE]** Создать `internal/cli/root.go`.
  CRITICAL:
  ```
  func NewRootCmd(deps Deps) *cobra.Command {
      root := &cobra.Command{
          Use: "thingsexporter",
          Short: "Export Things 3 SQLite database to JSON or Markdown",
          SilenceUsage: true,
          SilenceErrors: true,
          RunE: func(cmd *cobra.Command, args []string) error {
              // Поведение root БЕЗ подкоманды = export с дефолтами
              return runExportWithDefaults(cmd.Context(), deps)
          },
      }
      root.SetOut(deps.Stdout); root.SetErr(deps.Stderr); root.SetIn(deps.Stdin)
      root.AddCommand(newExportCmd(deps), newInspectCmd(deps), newVersionCmd(deps), newCompletionCmd(deps))
      return root
  }
  func Execute(deps Deps) error {
      root := NewRootCmd(deps)
      err := root.Execute()
      if err != nil { fmt.Fprintln(deps.Stderr, "error:", err.Error()) }
      return err
  }
  ```
  IMPORTANT: `runExportWithDefaults` извлекает дефолтные значения флагов и вызывает тот же `runExport(...)` из export.go.
- **T-6.10 [GREEN]** Тесты root-команды: `internal/cli/root_test.go`.
  CRITICAL:
  - `TestRoot_helpExits0` — `--help` → exit 0, stdout содержит `thingsexporter`.
  - `TestRoot_subcommandsRegistered` — `root.Find([]string{"export"})` находит, аналогично для `inspect/version/completion`.

### T-7: Wiring — cmd/thingsexporter/main.go + smoke

***Complexity: mechanical***
***Requirements: REQ-6.9, REQ-7.9, REQ-8.2, REQ-8.3***
***Preservation: CP-16***
GOAL: главная точка входа, использующая `DefaultDeps()`.

**Subtasks:**

- **T-7.1 [CODE]** Создать `cmd/thingsexporter/main.go`.
  CRITICAL:
  ```
  package main
  import ( "fmt", "os", "runtime/debug", "github.com/jtprogru/thingsexporter/internal/cli" )
  func main() {
      exitCode := 0
      defer func() {
          if r := recover(); r != nil {
              fmt.Fprintf(os.Stderr, "thingsexporter: panic: %v\n%s\n", r, debug.Stack())
              os.Exit(1)
          }
          os.Exit(exitCode)
      }()
      deps := cli.DefaultDeps()
      if err := cli.Execute(deps); err != nil {
          exitCode = cli.AsExitCode(err)
      }
  }
  ```
  IMPORTANT: точно такой же паттерн что и в `todushka/cmd/todushka/main.go` (recover + AsExitCode).
- **T-7.2 [GREEN]** Создать integration smoke-тест `internal/cli/integration_test.go`.
  CRITICAL: тест `TestIntegration_jsonExport_fixture` — build fixture, вызвать `Execute(deps)` с реальным `Deps` (но `OpenRepo` указывает на fixture path), стандартное `--include all`, парсить вывод как JSON, ассертить `meta.counts.tasks == 5` (или сколько в fixture), `meta.counts.areas == 2`, `len(hierarchy.areas) == 2`, `schema == "thingsexporter/v1"`. (REQ-8.2)
  И `TestIntegration_markdownExport_fixture` — `--format markdown`, ассертить наличие `# Inbox`, `# Areas`, хотя бы одной `## <area>` и одной `### <project>` строки. (REQ-8.3)
- **T-7.3 [VERIFY]** Запустить полный `task test-race` и `task lint`.
  GOAL: убедиться, что весь test-suite зелёный с race-детектором и линтер не возражает. Команды: `task test-race && task lint`. Ожидаемое: оба exit 0.

### T-8: Инфраструктура — Taskfile + golangci + goreleaser + workflows + dependabot + README

***Complexity: standard***
***Requirements: REQ-7.1..REQ-7.10, REQ-8.5***
GOAL: dev-инфра и релизный конвейер, идентичные `todushka` с подстановкой имени `thingsexporter`.

**Subtasks:**

- **T-8.1 [CODE]** Создать `Taskfile.yml` в корне.
  CRITICAL: скопировать ровно содержимое `todushka/Taskfile.yml`, заменить:
  - `BIN_DIR: bin` (оставить)
  - `CMD_PATH: ./cmd/thingsexporter`
  - все вхождения `todushka` → `thingsexporter`
  Дополнительный таргет `vuln`: `govulncheck ./...` (используется в CI). Добавить таргет `release-check`: `goreleaser check && goreleaser build --snapshot --clean --single-target`.
- **T-8.2 [CODE]** Создать `.golangci.yml` в корне.
  CRITICAL: скопировать ровно содержимое `todushka/.golangci.yml` (REQ-7.3 — без изменений).
- **T-8.3 [CODE]** Создать `.goreleaser.yaml` в корне.
  CRITICAL: скопировать `todushka/.goreleaser.yaml`, изменить:
  - `project_name: thingsexporter`
  - `builds[0].id: thingsexporter`, `main: ./cmd/thingsexporter`, `binary: thingsexporter`.
  - ldflags: `-X github.com/jtprogru/thingsexporter/internal/version.Version={{.Version}}` и т.д.
  - **CRITICAL: заменить `homebrew_casks` блок на `homebrew_formula`**:
    ```yaml
    homebrew_formula:
      - name: thingsexporter
        repository:
          owner: jtprogru
          name: homebrew-tap
          branch: main
          token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
        homepage: "https://github.com/jtprogru/thingsexporter"
        description: "Export Things 3 SQLite database to JSON or Markdown"
        license: "MIT"
        commit_author:
          name: goreleaserbot
          email: bot@goreleaser.com
        commit_msg_template: "chore(thingsexporter): bring formula to {{ .Tag }}"
        install: |
          bin.install "thingsexporter"
        test: |
          system "#{bin}/thingsexporter", "version"
    ```
  - `release.github.name: thingsexporter`.
  - Убрать `homebrew_casks.hooks.post.install` (макос-quarantine не нужен для CLI-формулы).
  IMPORTANT: оставить `sboms` и `signs` (cosign keyless) — без изменений (REQ-7.6).
- **T-8.4 [CODE]** Создать `.github/workflows/ci.yml`.
  CRITICAL: скопировать `todushka/.github/workflows/ci.yml` без изменений, кроме имени и команды `goreleaser-check` job — он уже не зависит от имени бинаря. Actions запиннены по SHA (оставить как есть).
- **T-8.5 [CODE]** Создать `.github/workflows/release.yml`.
  CRITICAL: скопировать `todushka/.github/workflows/release.yml`, без изменений. Permissions `contents: write` + `id-token: write`. Secrets — `GITHUB_TOKEN` и `HOMEBREW_TAP_GITHUB_TOKEN`. NOTE: secret `HOMEBREW_TAP_GITHUB_TOKEN` пользователь добавит вручную в repo settings перед первым `git tag v0.1.0` (REQ-7.8).
- **T-8.6 [CODE]** Создать `.github/dependabot.yml`.
  CRITICAL: скопировать `todushka/.github/dependabot.yml` без изменений (REQ-7.10).
- **T-8.7 [CODE]** Создать `README.md` в корне.
  CRITICAL: разделы:
  - Title + 1-абзац описание.
  - **Install:** `brew install jtprogru/tap/thingsexporter` + `go install github.com/jtprogru/thingsexporter/cmd/thingsexporter@latest` + `git clone ... && task build`.
  - **Usage:** примеры команд: дефолт (auto-discover на macOS), `--db ... --format markdown --out tasks.md`, `--include tasks+tags`, `inspect`, `version`.
  - **Поддерживаемая версия БД:** `databaseVersion=26` (обновится по мере появления новых).
  - **License:** MIT.
  - Краткая ссылка на референсный Python-скрипт как авторская атрибуция.
- **T-8.8 [VERIFY]** Локально проверить инфраструктуру.
  GOAL: убедиться, что CI-conf валиден ещё до push. Команды:
  ```
  task lint
  task test-race
  goreleaser check
  goreleaser build --snapshot --clean --single-target
  ```
  Все четыре должны вернуть exit 0. IMPORTANT: `goreleaser check` ловит синтаксические ошибки в `.goreleaser.yaml`; `goreleaser build --snapshot` проверяет, что код собирается с CGO=0 под текущую платформу. Если упадёт — НЕ переходить к T-9.

### T-9: GATE — финальный checkpoint

***Complexity: mechanical***
***Requirements: все 40 REQ***
GOAL: подтвердить, что всё реализовано, всё зелёное, артефакт фазы Implementation готов к review.

**Subtasks:**

- **T-9.1 [VERIFY]** Прогнать ВЕСЬ test-suite с race-детектором.
  Команда: `task test-race`. Ожидаемое: 0 failed, coverage report сохранён в `cover.out`. CRITICAL: если хоть один тест падает — НЕ переходить к Implementation Report; фиксить, повторять.
- **T-9.2 [VERIFY]** Прогнать линтер и vuln-сканер.
  Команды: `task lint && govulncheck ./...`. Ожидаемое: оба exit 0. Никаких high-severity vulnerabilities.
- **T-9.3 [VERIFY]** Прогнать goreleaser dry-run.
  Команда: `goreleaser check && goreleaser build --snapshot --clean --single-target`. Ожидаемое: exit 0, в `dist/` появился snapshot-бинарь.
- **T-9.4 [VERIFY]** Manual smoke с реальной БД.
  Команды (на macOS пользователя):
  ```
  ./bin/thingsexporter --db /Users/jtprogru/Work/tmp/things3db/main.sqlite --include all --out /tmp/te.json
  jq '.meta.counts' /tmp/te.json
  diff <(jq -S '.meta.counts' /tmp/te.json) <(jq -S '.meta.counts' /Users/jtprogru/Work/tmp/things3db/things3.json)
  ./bin/thingsexporter --db /Users/jtprogru/Work/tmp/things3db/main.sqlite --format markdown --include all | head -30
  ./bin/thingsexporter version
  ```
  CRITICAL: счётчики должны совпадать (areas=4, tags=3, tasks=613, checklistItems=22, contacts=0, tombstones=388, taskTagLinks=55, areaTagLinks=0). Markdown — визуальная проверка читаемости. Version — содержит непустые Commit и Date после `task build` (если ldflags переданы).
  NOTE: если `task build` без ldflags — Version=dev, Commit=-, Date=- — это OK для локальной сборки; на релизе через goreleaser значения подставятся.
- **T-9.5 [VERIFY]** Подтвердить покрытие требований.
  GOAL: вручную пройтись по Coverage Matrix, убедиться, что каждый REQ имеет завершённую соответствующую задачу. Если что-то не реализовано — вернуться к соответствующему T-N. Ожидаемое: 40/40 покрыто.
