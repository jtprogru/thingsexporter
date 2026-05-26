package sqlite

import (
	"context"
	"database/sql"

	"github.com/jtprogru/thingsexporter/internal/things"
)

// Repository обёртка над *sql.DB с типизированными методами чтения Things 3.
type Repository struct {
	db *sql.DB
}

// NewRepository — конструктор. Сам *sql.DB обычно создан Open().
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// ReadAll вычитывает все таблицы в RawData. Не делает никаких преобразований.
func (r *Repository) ReadAll(ctx context.Context) (things.RawData, error) {
	var out things.RawData
	var err error
	if out.Areas, err = selectAreas(ctx, r.db); err != nil {
		return out, err
	}
	if out.Tags, err = selectTags(ctx, r.db); err != nil {
		return out, err
	}
	if out.Tasks, err = selectTasks(ctx, r.db); err != nil {
		return out, err
	}
	if out.Checklist, err = selectChecklist(ctx, r.db); err != nil {
		return out, err
	}
	if out.Contacts, err = selectContacts(ctx, r.db); err != nil {
		return out, err
	}
	if out.Tombstones, err = selectTombstones(ctx, r.db); err != nil {
		return out, err
	}
	if out.TaskTagPairs, err = selectTaskTags(ctx, r.db); err != nil {
		return out, err
	}
	if out.AreaTagPairs, err = selectAreaTags(ctx, r.db); err != nil {
		return out, err
	}
	if out.MetaRows, err = selectMetaRows(ctx, r.db); err != nil {
		return out, err
	}
	return out, nil
}

// ReadCounts возвращает только COUNT(*) по каждой таблице.
func (r *Repository) ReadCounts(ctx context.Context) (things.Counts, error) {
	return selectCounts(ctx, r.db)
}

// DatabaseVersion возвращает значение databaseVersion из таблицы Meta.
func (r *Repository) DatabaseVersion(ctx context.Context) (*int, error) {
	return selectDatabaseVersion(ctx, r.db)
}

// Close закрывает underlying *sql.DB.
func (r *Repository) Close() error { return r.db.Close() }
