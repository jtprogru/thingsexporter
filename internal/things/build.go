package things

import (
	"sort"
	"time"
)

// SchemaVersion фиксирует схему выгрузки. См. ADR-9.
const SchemaVersion = "thingsexporter/v1"

// BuildOptions — параметры сборки Export из RawData.
type BuildOptions struct {
	Source     string
	ExportedAt time.Time
	NoBlobs    bool
}

// Build конвертирует RawData в полный Export (включая все коллекции, Links и Hierarchy).
// Сокращение состава под пресет — задача пакета export/preset.
func Build(raw RawData, opts BuildOptions) Export {
	areas := buildAreas(raw, opts.NoBlobs)
	tags := buildTags(raw, opts.NoBlobs)
	checklist := buildChecklist(raw, opts.NoBlobs)
	contacts := buildContacts(raw)
	tombstones := buildTombstones(raw)

	tagByUUID := make(map[string]*Tag, len(tags))
	for i := range tags {
		tagByUUID[tags[i].UUID] = &tags[i]
	}
	areaByUUID := make(map[string]*Area, len(areas))
	for i := range areas {
		areaByUUID[areas[i].UUID] = &areas[i]
	}
	contactByUUID := make(map[string]*Contact, len(contacts))
	for i := range contacts {
		contactByUUID[contacts[i].UUID] = &contacts[i]
	}

	// teги по задаче/области
	tagsForTask := make(map[string][]string)
	for _, p := range raw.TaskTagPairs {
		tagsForTask[p.Task] = append(tagsForTask[p.Task], p.Tag)
	}
	tagsForArea := make(map[string][]string)
	for _, p := range raw.AreaTagPairs {
		tagsForArea[p.Area] = append(tagsForArea[p.Area], p.Tag)
	}

	// чек-лист по задаче
	checklistForTask := make(map[string][]ChecklistItem)
	for _, c := range checklist {
		if c.Task == nil {
			continue
		}
		checklistForTask[*c.Task] = append(checklistForTask[*c.Task], c)
	}
	for k := range checklistForTask {
		items := checklistForTask[k]
		sort.SliceStable(items, func(i, j int) bool {
			return indexLessNilLast(items[i].Index, items[j].Index)
		})
		checklistForTask[k] = items
	}

	// Resolve parentTitle для тегов.
	for i := range tags {
		if tags[i].Parent == nil {
			continue
		}
		if p, ok := tagByUUID[*tags[i].Parent]; ok && p != nil {
			tags[i].ParentTitle = p.Title
		}
	}

	// Обогащаем области: список тегов.
	for i := range areas {
		areas[i].Tags = tagRefs(tagsForArea[areas[i].UUID], tagByUUID)
	}

	// Конвертация задач + обогащение titlесов/тегов/чек-листа.
	tasks := make([]Task, 0, len(raw.Tasks))
	for _, rt := range raw.Tasks {
		t := convertTask(rt, opts.NoBlobs)
		if rt.Area != nil {
			if a, ok := areaByUUID[*rt.Area]; ok && a != nil {
				t.AreaTitle = a.Title
			}
		}
		if rt.Project != nil {
			if p := findTaskTitle(*rt.Project, raw.Tasks); p != nil {
				t.ProjectTitle = p
			}
		}
		if rt.Heading != nil {
			if h := findTaskTitle(*rt.Heading, raw.Tasks); h != nil {
				t.HeadingTitle = h
			}
		}
		if rt.Contact != nil {
			if c, ok := contactByUUID[*rt.Contact]; ok && c != nil {
				t.ContactName = c.DisplayName
			}
		}
		t.Tags = tagRefs(tagsForTask[t.UUID], tagByUUID)
		if items, ok := checklistForTask[t.UUID]; ok {
			t.Checklist = items
		} else {
			t.Checklist = []ChecklistItem{}
		}
		tasks = append(tasks, t)
	}

	hierarchy := buildHierarchy(tasks, areas)
	links := &Links{TaskTag: raw.TaskTagPairs, AreaTag: raw.AreaTagPairs}

	counts := Counts{
		Areas:          intPtr(len(areas)),
		Tags:           intPtr(len(tags)),
		Tasks:          intPtr(len(tasks)),
		ChecklistItems: intPtr(len(checklist)),
		Contacts:       intPtr(len(contacts)),
		Tombstones:     intPtr(len(tombstones)),
		TaskTagLinks:   intPtr(len(raw.TaskTagPairs)),
		AreaTagLinks:   intPtr(len(raw.AreaTagPairs)),
	}

	return Export{
		Schema: SchemaVersion,
		Meta: Meta{
			Source:     opts.Source,
			ExportedAt: opts.ExportedAt.UTC().Format("2006-01-02T15:04:05.000000-07:00"),
			Counts:     counts,
			DBMetaRows: raw.MetaRows,
		},
		Areas:          areas,
		Tags:           tags,
		Tasks:          tasks,
		ChecklistItems: checklist,
		Contacts:       contacts,
		Tombstones:     tombstones,
		Links:          links,
		Hierarchy:      &hierarchy,
	}
}

func buildAreas(raw RawData, noBlobs bool) []Area {
	out := make([]Area, 0, len(raw.Areas))
	for _, r := range raw.Areas {
		out = append(out, Area{
			UUID:         r.UUID,
			Title:        r.Title,
			Visible:      r.Visible,
			Index:        r.Index,
			CachedTags:   EncodeBlob(r.CachedTags, noBlobs),
			Experimental: EncodeBlob(r.Experimental, noBlobs),
			Tags:         []TagRef{},
		})
	}
	return out
}

func buildTags(raw RawData, noBlobs bool) []Tag {
	out := make([]Tag, 0, len(raw.Tags))
	for _, r := range raw.Tags {
		out = append(out, Tag{
			UUID:         r.UUID,
			Title:        r.Title,
			Shortcut:     r.Shortcut,
			UsedDate:     CoreDataToISO(r.UsedDate),
			Parent:       r.Parent,
			Index:        r.Index,
			Experimental: EncodeBlob(r.Experimental, noBlobs),
		})
	}
	return out
}

func buildChecklist(raw RawData, noBlobs bool) []ChecklistItem {
	out := make([]ChecklistItem, 0, len(raw.Checklist))
	for _, r := range raw.Checklist {
		out = append(out, ChecklistItem{
			UUID:                 r.UUID,
			UserModificationDate: CoreDataToISO(r.UserModificationDate),
			CreationDate:         CoreDataToISO(r.CreationDate),
			Title:                r.Title,
			Status:               r.Status,
			StopDate:             CoreDataToISO(r.StopDate),
			Index:                r.Index,
			Task:                 r.Task,
			LeavesTombstone:      r.LeavesTombstone,
			Experimental:         EncodeBlob(r.Experimental, noBlobs),
			StatusName:           ChecklistStatusName(r.Status),
		})
	}
	return out
}

func buildContacts(raw RawData) []Contact {
	out := make([]Contact, 0, len(raw.Contacts))
	for _, r := range raw.Contacts {
		// Поля RawContact и Contact сейчас совпадают; используем явное
		// type conversion, чтобы при будущем расхождении ловить ошибку компиляции.
		out = append(out, Contact(r))
	}
	return out
}

func buildTombstones(raw RawData) []Tombstone {
	out := make([]Tombstone, 0, len(raw.Tombstones))
	for _, r := range raw.Tombstones {
		out = append(out, Tombstone{
			UUID:              r.UUID,
			DeletionDate:      CoreDataToISO(r.DeletionDate),
			DeletedObjectUUID: r.DeletedObjectUUID,
		})
	}
	return out
}

func convertTask(r RawTask, noBlobs bool) Task {
	return Task{
		UUID:                        r.UUID,
		LeavesTombstone:             r.LeavesTombstone,
		CreationDate:                CoreDataToISO(r.CreationDate),
		UserModificationDate:        CoreDataToISO(r.UserModificationDate),
		StopDate:                    CoreDataToISO(r.StopDate),
		LastReminderInteractionDate: CoreDataToISO(r.LastReminderInteractionDate),
		RepeaterMigrationDate:       CoreDataToISO(r.RepeaterMigrationDate),
		Type:                        r.Type,
		Status:                      r.Status,
		Trashed:                     r.Trashed,
		Title:                       r.Title,
		Notes:                       r.Notes,
		NotesSync:                   r.NotesSync,

		CachedTags:   EncodeBlob(r.CachedTags, noBlobs),
		Start:        r.Start,
		StartBucket:  r.StartBucket,
		ReminderTime: r.ReminderTime,

		StartDate:                  r.StartDate,
		StartDateISO:               PackedDateToISO(r.StartDate),
		Deadline:                   r.Deadline,
		DeadlineISO:                PackedDateToISO(r.Deadline),
		DeadlineSuppressionDate:    r.DeadlineSuppressionDate,
		DeadlineSuppressionDateISO: PackedDateToISO(r.DeadlineSuppressionDate),
		T2DeadlineOffset:           r.T2DeadlineOffset,

		Index:                   r.Index,
		TodayIndex:              r.TodayIndex,
		TodayIndexReferenceDate: r.TodayIndexReferenceDate,

		Area:    r.Area,
		Project: r.Project,
		Heading: r.Heading,
		Contact: r.Contact,

		UntrashedLeafActionsCount:     r.UntrashedLeafActionsCount,
		OpenUntrashedLeafActionsCount: r.OpenUntrashedLeafActionsCount,
		ChecklistItemsCount:           r.ChecklistItemsCount,
		OpenChecklistItemsCount:       r.OpenChecklistItemsCount,

		Rt1RepeatingTemplate:            r.Rt1RepeatingTemplate,
		Rt1RecurrenceRule:               EncodeBlob(r.Rt1RecurrenceRule, noBlobs),
		Rt1InstanceCreationStartDate:    r.Rt1InstanceCreationStartDate,
		Rt1InstanceCreationPaused:       r.Rt1InstanceCreationPaused,
		Rt1InstanceCreationCount:        r.Rt1InstanceCreationCount,
		Rt1AfterCompletionReferenceDate: r.Rt1AfterCompletionReferenceDate,
		Rt1NextInstanceStartDate:        r.Rt1NextInstanceStartDate,

		Experimental: EncodeBlob(r.Experimental, noBlobs),
		Repeater:     EncodeBlob(r.Repeater, noBlobs),

		TypeName:   TaskTypeName(r.Type),
		StatusName: TaskStatusName(r.Status),
	}
}

func tagRefs(uuids []string, tagByUUID map[string]*Tag) []TagRef {
	out := make([]TagRef, 0, len(uuids))
	for _, u := range uuids {
		ref := TagRef{UUID: u}
		if t, ok := tagByUUID[u]; ok && t != nil {
			ref.Title = t.Title
		}
		out = append(out, ref)
	}
	return out
}

func findTaskTitle(uuid string, tasks []RawTask) *string {
	for i := range tasks {
		if tasks[i].UUID == uuid {
			return tasks[i].Title
		}
	}
	return nil
}

// buildHierarchy формирует срез Areas → root items + inbox.
// Включаются только не-trashed корневые задачи/проекты (project=nil, heading=nil).
func buildHierarchy(tasks []Task, areas []Area) Hierarchy {
	areaOrder := make([]Area, len(areas))
	copy(areaOrder, areas)
	sort.SliceStable(areaOrder, func(i, j int) bool {
		return indexLessNilLast(areaOrder[i].Index, areaOrder[j].Index)
	})

	itemsByArea := make(map[string][]Task)
	var inbox []Task
	for _, t := range tasks {
		if t.Trashed != nil && *t.Trashed == 1 {
			continue
		}
		if t.Project != nil {
			continue
		}
		if t.Heading != nil {
			continue
		}
		if t.TypeName == nil {
			continue
		}
		// hierarchy включает только задачи и проекты (typeName ∈ {todo, project})
		if *t.TypeName != "todo" && *t.TypeName != "project" {
			continue
		}
		if t.Area == nil {
			inbox = append(inbox, t)
		} else {
			itemsByArea[*t.Area] = append(itemsByArea[*t.Area], t)
		}
	}
	for k := range itemsByArea {
		items := itemsByArea[k]
		sort.SliceStable(items, func(i, j int) bool {
			return indexLessNilLast(items[i].Index, items[j].Index)
		})
		itemsByArea[k] = items
	}
	sort.SliceStable(inbox, func(i, j int) bool {
		return indexLessNilLast(inbox[i].Index, inbox[j].Index)
	})

	hierAreas := make([]HierarchyArea, 0, len(areaOrder))
	for _, a := range areaOrder {
		items := make([]HierarchyItem, 0, len(itemsByArea[a.UUID]))
		for _, t := range itemsByArea[a.UUID] {
			items = append(items, HierarchyItem{
				UUID:       t.UUID,
				Title:      t.Title,
				TypeName:   t.TypeName,
				StatusName: t.StatusName,
			})
		}
		hierAreas = append(hierAreas, HierarchyArea{
			UUID:    a.UUID,
			Title:   a.Title,
			Visible: a.Visible,
			Index:   a.Index,
			Tags:    a.Tags,
			Items:   items,
		})
	}
	inboxItems := make([]HierarchyItem, 0, len(inbox))
	for _, t := range inbox {
		inboxItems = append(inboxItems, HierarchyItem{
			UUID:       t.UUID,
			Title:      t.Title,
			TypeName:   t.TypeName,
			StatusName: t.StatusName,
		})
	}
	return Hierarchy{Areas: hierAreas, InboxOrOrphanTasks: inboxItems}
}

// indexLessNilLast — порядок сортировки: nil — в конец, иначе по возрастанию int64.
func indexLessNilLast(a, b *int64) bool {
	switch {
	case a == nil && b == nil:
		return false
	case a == nil:
		return false
	case b == nil:
		return true
	default:
		return *a < *b
	}
}

func intPtr(v int) *int { return &v }
