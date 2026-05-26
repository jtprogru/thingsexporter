package sqlite_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/stretchr/testify/require"
)

const fixtureDDL = `
CREATE TABLE "Meta" (
    "key" TEXT PRIMARY KEY,
    "value" TEXT
);
CREATE TABLE "TMArea" (
    "uuid" TEXT PRIMARY KEY,
    "title" TEXT,
    "visible" INTEGER,
    "index" INTEGER,
    "cachedTags" BLOB,
    "experimental" BLOB
);
CREATE TABLE "TMTag" (
    "uuid" TEXT PRIMARY KEY,
    "title" TEXT,
    "shortcut" TEXT,
    "usedDate" REAL,
    "parent" TEXT,
    "index" INTEGER,
    "experimental" BLOB
);
CREATE TABLE "TMContact" (
    "uuid" TEXT PRIMARY KEY,
    "displayName" TEXT,
    "firstName" TEXT,
    "lastName" TEXT,
    "emails" TEXT,
    "appleAddressBookId" TEXT,
    "index" INTEGER
);
CREATE TABLE "TMTaskTag" (
    "tasks" TEXT NOT NULL,
    "tags" TEXT NOT NULL
);
CREATE TABLE "TMAreaTag" (
    "areas" TEXT NOT NULL,
    "tags" TEXT NOT NULL
);
CREATE TABLE "TMChecklistItem" (
    "uuid" TEXT PRIMARY KEY,
    "userModificationDate" REAL,
    "creationDate" REAL,
    "title" TEXT,
    "status" INTEGER,
    "stopDate" REAL,
    "index" INTEGER,
    "task" TEXT,
    "leavesTombstone" INTEGER,
    "experimental" BLOB
);
CREATE TABLE "TMTombstone" (
    "uuid" TEXT PRIMARY KEY,
    "deletionDate" REAL,
    "deletedObjectUUID" TEXT
);
CREATE TABLE "TMTask" (
    "uuid" TEXT PRIMARY KEY,
    "leavesTombstone" INTEGER,
    "creationDate" REAL,
    "userModificationDate" REAL,
    "type" INTEGER,
    "status" INTEGER,
    "stopDate" REAL,
    "trashed" INTEGER,
    "title" TEXT,
    "notes" TEXT,
    "notesSync" INTEGER,
    "cachedTags" BLOB,
    "start" INTEGER,
    "startDate" INTEGER,
    "startBucket" INTEGER,
    "reminderTime" INTEGER,
    "lastReminderInteractionDate" REAL,
    "deadline" INTEGER,
    "deadlineSuppressionDate" INTEGER,
    "t2_deadlineOffset" INTEGER,
    "index" INTEGER,
    "todayIndex" INTEGER,
    "todayIndexReferenceDate" INTEGER,
    "area" TEXT,
    "project" TEXT,
    "heading" TEXT,
    "contact" TEXT,
    "untrashedLeafActionsCount" INTEGER,
    "openUntrashedLeafActionsCount" INTEGER,
    "checklistItemsCount" INTEGER,
    "openChecklistItemsCount" INTEGER,
    "rt1_repeatingTemplate" TEXT,
    "rt1_recurrenceRule" BLOB,
    "rt1_instanceCreationStartDate" INTEGER,
    "rt1_instanceCreationPaused" INTEGER,
    "rt1_instanceCreationCount" INTEGER,
    "rt1_afterCompletionReferenceDate" INTEGER,
    "rt1_nextInstanceStartDate" INTEGER,
    "experimental" BLOB,
    "repeater" BLOB,
    "repeaterMigrationDate" REAL
);
`

const fixtureSeed = `
INSERT INTO "Meta" ("key","value") VALUES (
  'databaseVersion',
  '<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<integer>26</integer>
</plist>'
);

INSERT INTO "TMArea" ("uuid","title","visible","index","cachedTags","experimental") VALUES
  ('A-work','Work',NULL,-100,NULL,NULL),
  ('A-home','Home',NULL,50,NULL,NULL);

INSERT INTO "TMTag" ("uuid","title","shortcut","usedDate","parent","index","experimental") VALUES
  ('T-p1','P1',NULL,724000000.0,NULL,1,NULL),
  ('T-p2','P2',NULL,NULL,'T-p1',2,NULL),
  ('T-work','work',NULL,NULL,NULL,3,NULL);

INSERT INTO "TMTask" (
  "uuid","leavesTombstone","creationDate","userModificationDate","type","status","stopDate","trashed",
  "title","notes","notesSync","cachedTags","start","startDate","startBucket","reminderTime",
  "lastReminderInteractionDate","deadline","deadlineSuppressionDate","t2_deadlineOffset",
  "index","todayIndex","todayIndexReferenceDate","area","project","heading","contact",
  "untrashedLeafActionsCount","openUntrashedLeafActionsCount","checklistItemsCount","openChecklistItemsCount",
  "rt1_repeatingTemplate","rt1_recurrenceRule","rt1_instanceCreationStartDate","rt1_instanceCreationPaused",
  "rt1_instanceCreationCount","rt1_afterCompletionReferenceDate","rt1_nextInstanceStartDate",
  "experimental","repeater","repeaterMigrationDate"
) VALUES
  ('task-1',0,724000000.0,724000000.0,0,0,NULL,0,'buy milk','',1,NULL,1,NULL,0,NULL,NULL,NULL,NULL,0,10,0,NULL,'A-work',NULL,NULL,NULL,-1,-1,2,1,NULL,NULL,NULL,0,0,NULL,NULL,NULL,NULL,NULL),
  ('task-trashed',0,724000000.0,724000000.0,0,3,724100000.0,1,'garbage','',1,NULL,1,NULL,0,NULL,NULL,NULL,NULL,0,20,0,NULL,'A-work',NULL,NULL,NULL,-1,-1,0,0,NULL,NULL,NULL,0,0,NULL,NULL,NULL,NULL,NULL),
  ('proj-1',0,724000000.0,724000000.0,1,0,NULL,0,'Build deck','',1,NULL,1,NULL,0,NULL,NULL,NULL,NULL,0,5,0,NULL,'A-home',NULL,NULL,NULL,-1,-1,0,0,NULL,NULL,NULL,0,0,NULL,NULL,NULL,NULL,NULL),
  ('task-inbox',0,724000000.0,724000000.0,0,0,NULL,0,'call dentist','',1,NULL,1,NULL,0,NULL,NULL,NULL,NULL,0,1,0,NULL,NULL,NULL,NULL,NULL,-1,-1,0,0,NULL,NULL,NULL,0,0,NULL,NULL,NULL,NULL,NULL),
  ('task-in-proj',0,724000000.0,724000000.0,0,0,NULL,0,'buy wood','',1,NULL,1,NULL,0,NULL,NULL,NULL,NULL,0,0,0,NULL,NULL,'proj-1',NULL,NULL,-1,-1,0,0,NULL,NULL,NULL,0,0,NULL,NULL,NULL,NULL,NULL);

INSERT INTO "TMChecklistItem" ("uuid","userModificationDate","creationDate","title","status","stopDate","index","task","leavesTombstone","experimental") VALUES
  ('cl-1',724000000.0,724000000.0,'check brand',0,NULL,0,'task-1',0,NULL),
  ('cl-2',724000000.0,724000000.0,'verify expiry',3,724100000.0,1,'task-1',0,NULL);

INSERT INTO "TMContact" ("uuid","displayName","firstName","lastName","emails","appleAddressBookId","index") VALUES
  ('C-1','Alice','Alice','Doe','alice@example.com',NULL,1);

INSERT INTO "TMTombstone" ("uuid","deletionDate","deletedObjectUUID") VALUES
  ('tomb-1',724000000.0,'gone-uuid');

INSERT INTO "TMTaskTag" ("tasks","tags") VALUES
  ('task-1','T-p1'),
  ('task-1','T-work'),
  ('proj-1','T-p2');

INSERT INTO "TMAreaTag" ("areas","tags") VALUES
  ('A-work','T-p1');
`

// buildFixture создаёт временный SQLite-файл со схемой Things 3 и
// контролируемым набором данных. Возвращает путь к файлу.
func buildFixture(t testing.TB) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "things.sqlite")
	db, err := sql.Open("sqlite", "file:"+path+"?mode=rwc")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, fixtureDDL); err != nil {
		t.Fatalf("DDL: %v", err)
	}
	if _, err := db.ExecContext(ctx, fixtureSeed); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	return path
}
