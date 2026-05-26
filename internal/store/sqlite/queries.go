package sqlite

import (
	"context"
	"database/sql"
	"regexp"
	"strconv"

	"github.com/jtprogru/thingsexporter/internal/things"
)

const selectAreasSQL = `SELECT
	"uuid", "title", "visible", "index", "cachedTags", "experimental"
FROM "TMArea"`

const selectTagsSQL = `SELECT
	"uuid", "title", "shortcut", "usedDate", "parent", "index", "experimental"
FROM "TMTag"`

const selectTasksSQL = `SELECT
	"uuid", "leavesTombstone",
	"creationDate", "userModificationDate", "stopDate",
	"lastReminderInteractionDate", "repeaterMigrationDate",
	"type", "status", "trashed",
	"title", "notes", "notesSync",
	"cachedTags", "start", "startBucket", "reminderTime",
	"startDate", "deadline", "deadlineSuppressionDate", "t2_deadlineOffset",
	"index", "todayIndex", "todayIndexReferenceDate",
	"area", "project", "heading", "contact",
	"untrashedLeafActionsCount", "openUntrashedLeafActionsCount",
	"checklistItemsCount", "openChecklistItemsCount",
	"rt1_repeatingTemplate", "rt1_recurrenceRule",
	"rt1_instanceCreationStartDate", "rt1_instanceCreationPaused",
	"rt1_instanceCreationCount", "rt1_afterCompletionReferenceDate",
	"rt1_nextInstanceStartDate",
	"experimental", "repeater"
FROM "TMTask"`

const selectChecklistSQL = `SELECT
	"uuid", "userModificationDate", "creationDate",
	"title", "status", "stopDate", "index", "task",
	"leavesTombstone", "experimental"
FROM "TMChecklistItem"`

const selectContactsSQL = `SELECT
	"uuid", "displayName", "firstName", "lastName",
	"emails", "appleAddressBookId", "index"
FROM "TMContact"`

const selectTombstonesSQL = `SELECT
	"uuid", "deletionDate", "deletedObjectUUID"
FROM "TMTombstone"`

const selectTaskTagsSQL = `SELECT "tasks", "tags" FROM "TMTaskTag"`
const selectAreaTagsSQL = `SELECT "areas", "tags" FROM "TMAreaTag"`
const selectMetaRowsSQL = `SELECT "key", "value" FROM "Meta"`

const selectMetaDatabaseVersionSQL = `SELECT "value" FROM "Meta" WHERE "key" = 'databaseVersion'`

func nullStr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func nullInt(ns sql.NullInt64) *int64 {
	if !ns.Valid {
		return nil
	}
	v := ns.Int64
	return &v
}

func nullFloat(ns sql.NullFloat64) *float64 {
	if !ns.Valid {
		return nil
	}
	v := ns.Float64
	return &v
}

func selectAreas(ctx context.Context, db *sql.DB) ([]things.RawArea, error) {
	rows, err := db.QueryContext(ctx, selectAreasSQL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []things.RawArea
	for rows.Next() {
		var (
			uuid                     string
			title                    sql.NullString
			visible, index           sql.NullInt64
			cachedTags, experimental []byte
		)
		if err := rows.Scan(&uuid, &title, &visible, &index, &cachedTags, &experimental); err != nil {
			return nil, err
		}
		out = append(out, things.RawArea{
			UUID:         uuid,
			Title:        nullStr(title),
			Visible:      nullInt(visible),
			Index:        nullInt(index),
			CachedTags:   cachedTags,
			Experimental: experimental,
		})
	}
	return out, rows.Err()
}

func selectTags(ctx context.Context, db *sql.DB) ([]things.RawTag, error) {
	rows, err := db.QueryContext(ctx, selectTagsSQL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []things.RawTag
	for rows.Next() {
		var (
			uuid                    string
			title, shortcut, parent sql.NullString
			usedDate                sql.NullFloat64
			index                   sql.NullInt64
			experimental            []byte
		)
		if err := rows.Scan(&uuid, &title, &shortcut, &usedDate, &parent, &index, &experimental); err != nil {
			return nil, err
		}
		out = append(out, things.RawTag{
			UUID:         uuid,
			Title:        nullStr(title),
			Shortcut:     nullStr(shortcut),
			UsedDate:     nullFloat(usedDate),
			Parent:       nullStr(parent),
			Index:        nullInt(index),
			Experimental: experimental,
		})
	}
	return out, rows.Err()
}

func selectTasks(ctx context.Context, db *sql.DB) ([]things.RawTask, error) {
	rows, err := db.QueryContext(ctx, selectTasksSQL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []things.RawTask
	for rows.Next() {
		var (
			uuid                                                           string
			leavesTombstone                                                sql.NullInt64
			creationDate, userModificationDate, stopDate                   sql.NullFloat64
			lastReminderInteractionDate, repeaterMigrationDate             sql.NullFloat64
			typ, status, trashed                                           sql.NullInt64
			title, notes                                                   sql.NullString
			notesSync                                                      sql.NullInt64
			cachedTags                                                     []byte
			start, startBucket, reminderTime                               sql.NullInt64
			startDate, deadline, deadlineSuppressionDate, t2DeadlineOffset sql.NullInt64
			idx, todayIndex, todayIndexRefDate                             sql.NullInt64
			area, project, heading, contact                                sql.NullString
			untrashedLeafCount, openUntrashedLeafCount                     sql.NullInt64
			checklistCount, openChecklistCount                             sql.NullInt64
			rt1RepeatingTemplate                                           sql.NullString
			rt1RecurrenceRule                                              []byte
			rt1InstanceCreationStart, rt1InstanceCreationPaused            sql.NullInt64
			rt1InstanceCreationCount, rt1AfterCompletionRefDate            sql.NullInt64
			rt1NextInstanceStartDate                                       sql.NullInt64
			experimental, repeater                                         []byte
		)
		if err := rows.Scan(
			&uuid, &leavesTombstone,
			&creationDate, &userModificationDate, &stopDate,
			&lastReminderInteractionDate, &repeaterMigrationDate,
			&typ, &status, &trashed,
			&title, &notes, &notesSync,
			&cachedTags, &start, &startBucket, &reminderTime,
			&startDate, &deadline, &deadlineSuppressionDate, &t2DeadlineOffset,
			&idx, &todayIndex, &todayIndexRefDate,
			&area, &project, &heading, &contact,
			&untrashedLeafCount, &openUntrashedLeafCount,
			&checklistCount, &openChecklistCount,
			&rt1RepeatingTemplate, &rt1RecurrenceRule,
			&rt1InstanceCreationStart, &rt1InstanceCreationPaused,
			&rt1InstanceCreationCount, &rt1AfterCompletionRefDate,
			&rt1NextInstanceStartDate,
			&experimental, &repeater,
		); err != nil {
			return nil, err
		}
		out = append(out, things.RawTask{
			UUID:                            uuid,
			LeavesTombstone:                 nullInt(leavesTombstone),
			CreationDate:                    nullFloat(creationDate),
			UserModificationDate:            nullFloat(userModificationDate),
			StopDate:                        nullFloat(stopDate),
			LastReminderInteractionDate:     nullFloat(lastReminderInteractionDate),
			RepeaterMigrationDate:           nullFloat(repeaterMigrationDate),
			Type:                            nullInt(typ),
			Status:                          nullInt(status),
			Trashed:                         nullInt(trashed),
			Title:                           nullStr(title),
			Notes:                           nullStr(notes),
			NotesSync:                       nullInt(notesSync),
			CachedTags:                      cachedTags,
			Start:                           nullInt(start),
			StartBucket:                     nullInt(startBucket),
			ReminderTime:                    nullInt(reminderTime),
			StartDate:                       nullInt(startDate),
			Deadline:                        nullInt(deadline),
			DeadlineSuppressionDate:         nullInt(deadlineSuppressionDate),
			T2DeadlineOffset:                nullInt(t2DeadlineOffset),
			Index:                           nullInt(idx),
			TodayIndex:                      nullInt(todayIndex),
			TodayIndexReferenceDate:         nullInt(todayIndexRefDate),
			Area:                            nullStr(area),
			Project:                         nullStr(project),
			Heading:                         nullStr(heading),
			Contact:                         nullStr(contact),
			UntrashedLeafActionsCount:       nullInt(untrashedLeafCount),
			OpenUntrashedLeafActionsCount:   nullInt(openUntrashedLeafCount),
			ChecklistItemsCount:             nullInt(checklistCount),
			OpenChecklistItemsCount:         nullInt(openChecklistCount),
			Rt1RepeatingTemplate:            nullStr(rt1RepeatingTemplate),
			Rt1RecurrenceRule:               rt1RecurrenceRule,
			Rt1InstanceCreationStartDate:    nullInt(rt1InstanceCreationStart),
			Rt1InstanceCreationPaused:       nullInt(rt1InstanceCreationPaused),
			Rt1InstanceCreationCount:        nullInt(rt1InstanceCreationCount),
			Rt1AfterCompletionReferenceDate: nullInt(rt1AfterCompletionRefDate),
			Rt1NextInstanceStartDate:        nullInt(rt1NextInstanceStartDate),
			Experimental:                    experimental,
			Repeater:                        repeater,
		})
	}
	return out, rows.Err()
}

func selectChecklist(ctx context.Context, db *sql.DB) ([]things.RawChecklist, error) {
	rows, err := db.QueryContext(ctx, selectChecklistSQL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []things.RawChecklist
	for rows.Next() {
		var (
			uuid                                         string
			userModificationDate, creationDate, stopDate sql.NullFloat64
			title                                        sql.NullString
			status, index, leavesTombstone               sql.NullInt64
			task                                         sql.NullString
			experimental                                 []byte
		)
		if err := rows.Scan(&uuid, &userModificationDate, &creationDate, &title,
			&status, &stopDate, &index, &task, &leavesTombstone, &experimental); err != nil {
			return nil, err
		}
		out = append(out, things.RawChecklist{
			UUID:                 uuid,
			UserModificationDate: nullFloat(userModificationDate),
			CreationDate:         nullFloat(creationDate),
			Title:                nullStr(title),
			Status:               nullInt(status),
			StopDate:             nullFloat(stopDate),
			Index:                nullInt(index),
			Task:                 nullStr(task),
			LeavesTombstone:      nullInt(leavesTombstone),
			Experimental:         experimental,
		})
	}
	return out, rows.Err()
}

func selectContacts(ctx context.Context, db *sql.DB) ([]things.RawContact, error) {
	rows, err := db.QueryContext(ctx, selectContactsSQL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []things.RawContact
	for rows.Next() {
		var (
			uuid                                                         string
			displayName, firstName, lastName, emails, appleAddressBookID sql.NullString
			index                                                        sql.NullInt64
		)
		if err := rows.Scan(&uuid, &displayName, &firstName, &lastName,
			&emails, &appleAddressBookID, &index); err != nil {
			return nil, err
		}
		out = append(out, things.RawContact{
			UUID:               uuid,
			DisplayName:        nullStr(displayName),
			FirstName:          nullStr(firstName),
			LastName:           nullStr(lastName),
			Emails:             nullStr(emails),
			AppleAddressBookID: nullStr(appleAddressBookID),
			Index:              nullInt(index),
		})
	}
	return out, rows.Err()
}

func selectTombstones(ctx context.Context, db *sql.DB) ([]things.RawTombstone, error) {
	rows, err := db.QueryContext(ctx, selectTombstonesSQL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []things.RawTombstone
	for rows.Next() {
		var (
			uuid              string
			deletionDate      sql.NullFloat64
			deletedObjectUUID sql.NullString
		)
		if err := rows.Scan(&uuid, &deletionDate, &deletedObjectUUID); err != nil {
			return nil, err
		}
		out = append(out, things.RawTombstone{
			UUID:              uuid,
			DeletionDate:      nullFloat(deletionDate),
			DeletedObjectUUID: nullStr(deletedObjectUUID),
		})
	}
	return out, rows.Err()
}

func selectTaskTags(ctx context.Context, db *sql.DB) ([]things.TaskTagLink, error) {
	rows, err := db.QueryContext(ctx, selectTaskTagsSQL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []things.TaskTagLink
	for rows.Next() {
		var task, tag string
		if err := rows.Scan(&task, &tag); err != nil {
			return nil, err
		}
		out = append(out, things.TaskTagLink{Task: task, Tag: tag})
	}
	return out, rows.Err()
}

func selectAreaTags(ctx context.Context, db *sql.DB) ([]things.AreaTagLink, error) {
	rows, err := db.QueryContext(ctx, selectAreaTagsSQL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []things.AreaTagLink
	for rows.Next() {
		var area, tag string
		if err := rows.Scan(&area, &tag); err != nil {
			return nil, err
		}
		out = append(out, things.AreaTagLink{Area: area, Tag: tag})
	}
	return out, rows.Err()
}

func selectMetaRows(ctx context.Context, db *sql.DB) ([]things.MetaRow, error) {
	rows, err := db.QueryContext(ctx, selectMetaRowsSQL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []things.MetaRow
	for rows.Next() {
		var key, value sql.NullString
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		mr := things.MetaRow{}
		if key.Valid {
			mr.Key = key.String
		}
		if value.Valid {
			mr.Value = value.String
		}
		out = append(out, mr)
	}
	return out, rows.Err()
}

var dbVersionRe = regexp.MustCompile(`<integer>\s*(\d+)\s*</integer>`)

func selectDatabaseVersion(ctx context.Context, db *sql.DB) (*int, error) {
	var raw sql.NullString
	err := db.QueryRowContext(ctx, selectMetaDatabaseVersionSQL).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !raw.Valid {
		return nil, nil
	}
	m := dbVersionRe.FindStringSubmatch(raw.String)
	if len(m) != 2 {
		// Может быть просто число "26" без plist.
		if v, err := strconv.Atoi(raw.String); err == nil {
			return &v, nil
		}
		return nil, nil
	}
	v, err := strconv.Atoi(m[1])
	if err != nil {
		return nil, nil
	}
	return &v, nil
}

func selectCounts(ctx context.Context, db *sql.DB) (things.Counts, error) {
	tables := map[string]**int{}
	c := things.Counts{}
	tables[`"TMArea"`] = &c.Areas
	tables[`"TMTag"`] = &c.Tags
	tables[`"TMTask"`] = &c.Tasks
	tables[`"TMChecklistItem"`] = &c.ChecklistItems
	tables[`"TMContact"`] = &c.Contacts
	tables[`"TMTombstone"`] = &c.Tombstones
	tables[`"TMTaskTag"`] = &c.TaskTagLinks
	tables[`"TMAreaTag"`] = &c.AreaTagLinks
	for t, dest := range tables {
		var n int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+t).Scan(&n); err != nil {
			return c, err
		}
		*dest = &n
	}
	return c, nil
}
