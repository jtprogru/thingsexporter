package things

// RawData is the result of reading the DB BEFORE date/enum/BLOB conversion.
// The storage layer fills RawData, and Build consumes it and returns an Export.
type RawData struct {
	Areas        []RawArea
	Tags         []RawTag
	Tasks        []RawTask
	Checklist    []RawChecklist
	Contacts     []RawContact
	Tombstones   []RawTombstone
	TaskTagPairs []TaskTagLink
	AreaTagPairs []AreaTagLink
	MetaRows     []MetaRow
}

// RawArea is a raw TMArea row.
type RawArea struct {
	UUID         string
	Title        *string
	Visible      *int64
	Index        *int64
	CachedTags   []byte
	Experimental []byte
}

// RawTag is a raw TMTag row.
type RawTag struct {
	UUID         string
	Title        *string
	Shortcut     *string
	UsedDate     *float64
	Parent       *string
	Index        *int64
	Experimental []byte
}

// RawTask is a raw TMTask row.
type RawTask struct {
	UUID                        string
	LeavesTombstone             *int64
	CreationDate                *float64
	UserModificationDate        *float64
	StopDate                    *float64
	LastReminderInteractionDate *float64
	RepeaterMigrationDate       *float64
	Type                        *int64
	Status                      *int64
	Trashed                     *int64
	Title                       *string
	Notes                       *string
	NotesSync                   *int64
	CachedTags                  []byte
	Start                       *int64
	StartBucket                 *int64
	ReminderTime                *int64
	StartDate                   *int64
	Deadline                    *int64
	DeadlineSuppressionDate     *int64
	T2DeadlineOffset            *int64
	Index                       *int64
	TodayIndex                  *int64
	TodayIndexReferenceDate     *int64
	Area                        *string
	Project                     *string
	Heading                     *string
	Contact                     *string

	UntrashedLeafActionsCount     *int64
	OpenUntrashedLeafActionsCount *int64
	ChecklistItemsCount           *int64
	OpenChecklistItemsCount       *int64

	Rt1RepeatingTemplate            *string
	Rt1RecurrenceRule               []byte
	Rt1InstanceCreationStartDate    *int64
	Rt1InstanceCreationPaused       *int64
	Rt1InstanceCreationCount        *int64
	Rt1AfterCompletionReferenceDate *int64
	Rt1NextInstanceStartDate        *int64

	Experimental []byte
	Repeater     []byte
}

// RawChecklist is a raw TMChecklistItem row.
type RawChecklist struct {
	UUID                 string
	UserModificationDate *float64
	CreationDate         *float64
	Title                *string
	Status               *int64
	StopDate             *float64
	Index                *int64
	Task                 *string
	LeavesTombstone      *int64
	Experimental         []byte
}

// RawContact is a raw TMContact row.
type RawContact struct {
	UUID               string
	DisplayName        *string
	FirstName          *string
	LastName           *string
	Emails             *string
	AppleAddressBookID *string
	Index              *int64
}

// RawTombstone is a raw TMTombstone row.
type RawTombstone struct {
	UUID              string
	DeletionDate      *float64
	DeletedObjectUUID *string
}
