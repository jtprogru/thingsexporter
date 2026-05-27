# thingsexporter

CLI-утилита на Go для экспорта локальной БД macOS-приложения [Things 3](https://culturedcode.com/things/) в JSON или Markdown. Читает `main.sqlite` строго в read-only, не зависит от Python, единый статический бинарь без CGO.

## Установка

### Homebrew (macOS / Linux)

```sh
brew install jtprogru/tap/thingsexporter
```

### `go install`

```sh
go install github.com/jtprogru/thingsexporter/cmd/thingsexporter@latest
```

### Из исходников

```sh
git clone https://github.com/jtprogru/thingsexporter
cd thingsexporter
task build     # → bin/thingsexporter
```

## Использование

### Дефолт — полный JSON в stdout

На macOS путь к БД определяется автоматически:

```sh
thingsexporter > things.json
```

На Linux/Windows (или если БД лежит в нестандартном месте):

```sh
thingsexporter --db /path/to/main.sqlite > things.json
```

### Markdown

```sh
thingsexporter --format markdown --out tasks.md
```

Вывод — иерархия `# Inbox` → `# Areas → ## <area> → ### <project>` с GFM-чекбоксами `[ ]` / `[x]` / `[-]` (canceled), inline-тегами `#tag` и дедлайнами `⏰ YYYY-MM-DD`.

### Подмножество данных

```sh
# только задачи без связей
thingsexporter --include tasks

# задачи + теги
thingsexporter --include tasks+tags

# задачи + области, проекты, заголовки
thingsexporter --include tasks+projects

# оглавление — области, теги и иерархия без тел задач
thingsexporter --include structure

# всё (default)
thingsexporter --include all
```

### Полезные флаги

```
--db <path>          путь к main.sqlite (default: auto-discover на macOS)
--out <path|->       выходной файл, '-' = stdout (default: -)
--format json|markdown   формат вывода (default: json)
--include <preset>   состав: all|structure|tasks|tasks+tags|tasks+projects (default: all)
--indent <int>       отступ JSON, 0 = компактный (default: 2)
--no-blobs           не выводить BLOB-поля (по умолчанию они идут как hex)
--quiet              подавить сводку в stderr
```

### Подкоманды

```sh
thingsexporter inspect              # счётчики и databaseVersion без выгрузки
thingsexporter version              # версия + commit + дата сборки
thingsexporter completion bash      # shell-completion для bash/zsh/fish/powershell
```

## Поддерживаемая версия БД

На момент релиза поддерживается `databaseVersion = 26`. Если открыть БД другой версии — утилита выдаст warning в stderr, но продолжит экспорт. Сообщайте о новых версиях в issues.

## Эксплуатация

- БД всегда открывается строго в read-only (`mode=ro`), поэтому утилита безопасна для запуска при работающем Things 3.
- Никакие данные никуда не отправляются — обработка только локальная.
- BLOB-поля (`cachedTags`, `experimental`, `recurrenceRule`) по умолчанию сериализуются как `{"__blob_hex__": "<hex>"}`. Парсинг плистов/правил повтора — вне MVP.
- Trashed-задачи попадают в коллекцию `tasks`, но исключаются из `hierarchy` (как в референсном Python-скрипте).

## Лицензия

[MIT](./LICENSE) © Mikhail Savin
