package cli_test

import (
	"bytes"
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/jtprogru/thingsexporter/internal/cli"
	"github.com/jtprogru/thingsexporter/internal/export"
	jsonwriter "github.com/jtprogru/thingsexporter/internal/export/json"
	mdwriter "github.com/jtprogru/thingsexporter/internal/export/markdown"
	"github.com/jtprogru/thingsexporter/internal/export/preset"
	sqlitestore "github.com/jtprogru/thingsexporter/internal/store/sqlite"
	"github.com/stretchr/testify/require"
)

// fixtureDDL and fixtureSeed are duplicated minimally so the cli tests
// do not depend on internal packages.
const fixtureDDL = `
CREATE TABLE "Meta" ("key" TEXT PRIMARY KEY, "value" TEXT);
CREATE TABLE "TMArea" ("uuid" TEXT PRIMARY KEY, "title" TEXT, "visible" INTEGER, "index" INTEGER, "cachedTags" BLOB, "experimental" BLOB);
CREATE TABLE "TMTag" ("uuid" TEXT PRIMARY KEY, "title" TEXT, "shortcut" TEXT, "usedDate" REAL, "parent" TEXT, "index" INTEGER, "experimental" BLOB);
CREATE TABLE "TMContact" ("uuid" TEXT PRIMARY KEY, "displayName" TEXT, "firstName" TEXT, "lastName" TEXT, "emails" TEXT, "appleAddressBookId" TEXT, "index" INTEGER);
CREATE TABLE "TMTaskTag" ("tasks" TEXT NOT NULL, "tags" TEXT NOT NULL);
CREATE TABLE "TMAreaTag" ("areas" TEXT NOT NULL, "tags" TEXT NOT NULL);
CREATE TABLE "TMChecklistItem" ("uuid" TEXT PRIMARY KEY, "userModificationDate" REAL, "creationDate" REAL, "title" TEXT, "status" INTEGER, "stopDate" REAL, "index" INTEGER, "task" TEXT, "leavesTombstone" INTEGER, "experimental" BLOB);
CREATE TABLE "TMTombstone" ("uuid" TEXT PRIMARY KEY, "deletionDate" REAL, "deletedObjectUUID" TEXT);
CREATE TABLE "TMTask" (
  "uuid" TEXT PRIMARY KEY, "leavesTombstone" INTEGER,
  "creationDate" REAL, "userModificationDate" REAL,
  "type" INTEGER, "status" INTEGER, "stopDate" REAL, "trashed" INTEGER,
  "title" TEXT, "notes" TEXT, "notesSync" INTEGER,
  "cachedTags" BLOB, "start" INTEGER, "startDate" INTEGER, "startBucket" INTEGER, "reminderTime" INTEGER,
  "lastReminderInteractionDate" REAL, "deadline" INTEGER, "deadlineSuppressionDate" INTEGER, "t2_deadlineOffset" INTEGER,
  "index" INTEGER, "todayIndex" INTEGER, "todayIndexReferenceDate" INTEGER,
  "area" TEXT, "project" TEXT, "heading" TEXT, "contact" TEXT,
  "untrashedLeafActionsCount" INTEGER, "openUntrashedLeafActionsCount" INTEGER,
  "checklistItemsCount" INTEGER, "openChecklistItemsCount" INTEGER,
  "rt1_repeatingTemplate" TEXT, "rt1_recurrenceRule" BLOB,
  "rt1_instanceCreationStartDate" INTEGER, "rt1_instanceCreationPaused" INTEGER,
  "rt1_instanceCreationCount" INTEGER, "rt1_afterCompletionReferenceDate" INTEGER,
  "rt1_nextInstanceStartDate" INTEGER,
  "experimental" BLOB, "repeater" BLOB, "repeaterMigrationDate" REAL
);
`

const fixtureSeed = `
INSERT INTO "Meta" VALUES ('databaseVersion','<integer>26</integer>');
INSERT INTO "TMArea" VALUES ('A-work','Work',NULL,-100,NULL,NULL), ('A-home','Home',NULL,50,NULL,NULL);
INSERT INTO "TMTag" VALUES ('T-p1','P1',NULL,NULL,NULL,1,NULL), ('T-w','work',NULL,NULL,NULL,2,NULL);
INSERT INTO "TMTask" (
  "uuid","leavesTombstone","creationDate","userModificationDate",
  "type","status","stopDate","trashed","title","notes","notesSync",
  "cachedTags","start","startDate","startBucket","reminderTime",
  "lastReminderInteractionDate","deadline","deadlineSuppressionDate","t2_deadlineOffset",
  "index","todayIndex","todayIndexReferenceDate","area","project","heading","contact",
  "untrashedLeafActionsCount","openUntrashedLeafActionsCount","checklistItemsCount","openChecklistItemsCount",
  "rt1_repeatingTemplate","rt1_recurrenceRule","rt1_instanceCreationStartDate","rt1_instanceCreationPaused",
  "rt1_instanceCreationCount","rt1_afterCompletionReferenceDate","rt1_nextInstanceStartDate",
  "experimental","repeater","repeaterMigrationDate"
) VALUES
  ('t1',0,724000000.0,724000000.0,0,0,NULL,0,'buy milk','',1,NULL,1,NULL,0,NULL,NULL,NULL,NULL,0,10,0,NULL,'A-work',NULL,NULL,NULL,-1,-1,0,0,NULL,NULL,NULL,0,0,NULL,NULL,NULL,NULL,NULL),
  ('proj-1',0,724000000.0,724000000.0,1,0,NULL,0,'Build deck','',1,NULL,1,NULL,0,NULL,NULL,NULL,NULL,0,5,0,NULL,'A-home',NULL,NULL,NULL,-1,-1,0,0,NULL,NULL,NULL,0,0,NULL,NULL,NULL,NULL,NULL),
  ('ti',0,724000000.0,724000000.0,0,0,NULL,0,'call dentist','',1,NULL,1,NULL,0,NULL,NULL,NULL,NULL,0,1,0,NULL,NULL,NULL,NULL,NULL,-1,-1,0,0,NULL,NULL,NULL,0,0,NULL,NULL,NULL,NULL,NULL);
INSERT INTO "TMTaskTag" VALUES ('t1','T-p1'), ('t1','T-w');
INSERT INTO "TMAreaTag" VALUES ('A-work','T-p1');
`

func buildFixtureDB(t testing.TB) string {
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

func newTestDeps(t testing.TB) (cli.Deps, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	dbPath := buildFixtureDB(t)
	openRepo := func(p string) (*sqlitestore.Repository, error) {
		db, err := sqlitestore.Open(p)
		if err != nil {
			return nil, err
		}
		return sqlitestore.NewRepository(db), nil
	}
	deps := cli.Deps{
		Stdout:   stdout,
		Stderr:   stderr,
		Clock:    func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) },
		OpenRepo: openRepo,
		DiscoverDB: func() (string, bool) {
			return dbPath, true
		},
		Writers:             export.NewRegistry(jsonwriter.Writer{}, mdwriter.Writer{}),
		Presets:             preset.NewRegistry(preset.All{}, preset.Structure{}, preset.Tasks{}, preset.TasksTags{}, preset.TasksProjects{}),
		SupportedDBVersions: []int{26},
	}
	return deps, stdout, stderr
}

// runCmd runs the root command with the given argv (without the program name).
func runCmd(t testing.TB, deps cli.Deps, argv ...string) error {
	t.Helper()
	root := cli.NewRootCmd(deps)
	root.SetArgs(argv)
	return root.Execute()
}
