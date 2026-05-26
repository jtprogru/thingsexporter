// Package things содержит доменные типы Things 3 экспорта,
// конвертеры дат, маппинг enum-кодов и сборку Export-структуры.
package things

// Export — корневой контейнер выгрузки.
// Поля, нерелевантные текущему пресету, опускаются через omitempty.
type Export struct {
	Schema         string          `json:"schema"`
	Meta           Meta            `json:"meta"`
	Areas          []Area          `json:"areas,omitempty"`
	Tags           []Tag           `json:"tags,omitempty"`
	Tasks          []Task          `json:"tasks,omitempty"`
	ChecklistItems []ChecklistItem `json:"checklistItems,omitempty"`
	Contacts       []Contact       `json:"contacts,omitempty"`
	Tombstones     []Tombstone     `json:"tombstones,omitempty"`
	Links          *Links          `json:"links,omitempty"`
	Hierarchy      *Hierarchy      `json:"hierarchy,omitempty"`
}

// Meta — служебные данные экспорта.
type Meta struct {
	Source     string    `json:"source"`
	ExportedAt string    `json:"exportedAt"`
	Counts     Counts    `json:"counts"`
	DBMetaRows []MetaRow `json:"db_meta_rows"`
}

// Counts — счётчики коллекций. Указатели присутствуют только для
// коллекций, реально включённых в Export (ADR-3).
type Counts struct {
	Areas          *int `json:"areas,omitempty"`
	Tags           *int `json:"tags,omitempty"`
	Tasks          *int `json:"tasks,omitempty"`
	ChecklistItems *int `json:"checklistItems,omitempty"`
	Contacts       *int `json:"contacts,omitempty"`
	Tombstones     *int `json:"tombstones,omitempty"`
	TaskTagLinks   *int `json:"taskTagLinks,omitempty"`
	AreaTagLinks   *int `json:"areaTagLinks,omitempty"`
}

// MetaRow — строка таблицы Meta как есть.
type MetaRow struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Area — доменная модель TMArea.
type Area struct {
	UUID         string     `json:"uuid"`
	Title        *string    `json:"title"`
	Visible      *int64     `json:"visible"`
	Index        *int64     `json:"index"`
	CachedTags   *BlobValue `json:"cachedTags"`
	Experimental *BlobValue `json:"experimental"`
	Tags         []TagRef   `json:"tags"`
}

// Tag — доменная модель TMTag.
type Tag struct {
	UUID         string     `json:"uuid"`
	Title        *string    `json:"title"`
	Shortcut     *string    `json:"shortcut"`
	UsedDate     *string    `json:"usedDate"`
	Parent       *string    `json:"parent"`
	Index        *int64     `json:"index"`
	Experimental *BlobValue `json:"experimental"`
	ParentTitle  *string    `json:"parentTitle"`
}

// Task — доменная модель TMTask со всеми обогащёнными полями.
type Task struct {
	UUID                        string  `json:"uuid"`
	LeavesTombstone             *int64  `json:"leavesTombstone"`
	CreationDate                *string `json:"creationDate"`
	UserModificationDate        *string `json:"userModificationDate"`
	StopDate                    *string `json:"stopDate"`
	LastReminderInteractionDate *string `json:"lastReminderInteractionDate"`
	RepeaterMigrationDate       *string `json:"repeaterMigrationDate"`
	Type                        *int64  `json:"type"`
	Status                      *int64  `json:"status"`
	Trashed                     *int64  `json:"trashed"`
	Title                       *string `json:"title"`
	Notes                       *string `json:"notes"`
	NotesSync                   *int64  `json:"notesSync"`

	CachedTags   *BlobValue `json:"cachedTags"`
	Start        *int64     `json:"start"`
	StartBucket  *int64     `json:"startBucket"`
	ReminderTime *int64     `json:"reminderTime"`

	StartDate                  *int64  `json:"startDate"`
	StartDateISO               *string `json:"startDateISO"`
	Deadline                   *int64  `json:"deadline"`
	DeadlineISO                *string `json:"deadlineISO"`
	DeadlineSuppressionDate    *int64  `json:"deadlineSuppressionDate"`
	DeadlineSuppressionDateISO *string `json:"deadlineSuppressionDateISO"`
	T2DeadlineOffset           *int64  `json:"t2_deadlineOffset"`

	Index                   *int64 `json:"index"`
	TodayIndex              *int64 `json:"todayIndex"`
	TodayIndexReferenceDate *int64 `json:"todayIndexReferenceDate"`

	Area    *string `json:"area"`
	Project *string `json:"project"`
	Heading *string `json:"heading"`
	Contact *string `json:"contact"`

	UntrashedLeafActionsCount     *int64 `json:"untrashedLeafActionsCount"`
	OpenUntrashedLeafActionsCount *int64 `json:"openUntrashedLeafActionsCount"`
	ChecklistItemsCount           *int64 `json:"checklistItemsCount"`
	OpenChecklistItemsCount       *int64 `json:"openChecklistItemsCount"`

	Rt1RepeatingTemplate            *string    `json:"rt1_repeatingTemplate"`
	Rt1RecurrenceRule               *BlobValue `json:"rt1_recurrenceRule"`
	Rt1InstanceCreationStartDate    *int64     `json:"rt1_instanceCreationStartDate"`
	Rt1InstanceCreationPaused       *int64     `json:"rt1_instanceCreationPaused"`
	Rt1InstanceCreationCount        *int64     `json:"rt1_instanceCreationCount"`
	Rt1AfterCompletionReferenceDate *int64     `json:"rt1_afterCompletionReferenceDate"`
	Rt1NextInstanceStartDate        *int64     `json:"rt1_nextInstanceStartDate"`

	Experimental *BlobValue `json:"experimental"`
	Repeater     *BlobValue `json:"repeater"`

	// Обогащённые поля
	TypeName     *string         `json:"typeName"`
	StatusName   *string         `json:"statusName"`
	AreaTitle    *string         `json:"areaTitle,omitempty"`
	ProjectTitle *string         `json:"projectTitle,omitempty"`
	HeadingTitle *string         `json:"headingTitle,omitempty"`
	ContactName  *string         `json:"contactName,omitempty"`
	Tags         []TagRef        `json:"tags,omitempty"`
	Checklist    []ChecklistItem `json:"checklist,omitempty"`
}

// ChecklistItem — TMChecklistItem.
type ChecklistItem struct {
	UUID                 string     `json:"uuid"`
	UserModificationDate *string    `json:"userModificationDate"`
	CreationDate         *string    `json:"creationDate"`
	Title                *string    `json:"title"`
	Status               *int64     `json:"status"`
	StopDate             *string    `json:"stopDate"`
	Index                *int64     `json:"index"`
	Task                 *string    `json:"task"`
	LeavesTombstone      *int64     `json:"leavesTombstone"`
	Experimental         *BlobValue `json:"experimental"`
	StatusName           *string    `json:"statusName"`
}

// Contact — TMContact.
type Contact struct {
	UUID               string  `json:"uuid"`
	DisplayName        *string `json:"displayName"`
	FirstName          *string `json:"firstName"`
	LastName           *string `json:"lastName"`
	Emails             *string `json:"emails"`
	AppleAddressBookID *string `json:"appleAddressBookId"`
	Index              *int64  `json:"index"`
}

// Tombstone — TMTombstone.
type Tombstone struct {
	UUID              string  `json:"uuid"`
	DeletionDate      *string `json:"deletionDate"`
	DeletedObjectUUID *string `json:"deletedObjectUUID"`
}

// TagRef — компактная ссылка на тег при обогащении задач/областей.
type TagRef struct {
	UUID  string  `json:"uuid"`
	Title *string `json:"title"`
}

// Links — пары многие-ко-многим как они есть в БД.
type Links struct {
	TaskTag []TaskTagLink `json:"taskTag"`
	AreaTag []AreaTagLink `json:"areaTag"`
}

// TaskTagLink — одна пара TMTaskTag.
type TaskTagLink struct {
	Task string `json:"task"`
	Tag  string `json:"tag"`
}

// AreaTagLink — одна пара TMAreaTag.
type AreaTagLink struct {
	Area string `json:"area"`
	Tag  string `json:"tag"`
}

// Hierarchy — иерархический срез для пресета all.
type Hierarchy struct {
	Areas              []HierarchyArea `json:"areas"`
	InboxOrOrphanTasks []HierarchyItem `json:"inbox_or_orphan_tasks"`
}

// HierarchyArea — область с её корневыми задачами/проектами.
type HierarchyArea struct {
	UUID    string          `json:"uuid"`
	Title   *string         `json:"title"`
	Visible *int64          `json:"visible"`
	Index   *int64          `json:"index"`
	Tags    []TagRef        `json:"tags"`
	Items   []HierarchyItem `json:"items"`
}

// HierarchyItem — компактная карточка задачи/проекта внутри иерархии.
type HierarchyItem struct {
	UUID       string  `json:"uuid"`
	Title      *string `json:"title"`
	TypeName   *string `json:"typeName"`
	StatusName *string `json:"statusName"`
}
