package things

var (
	taskTypes = map[int64]string{
		0: "todo",
		1: "project",
		2: "heading",
	}
	taskStatuses = map[int64]string{
		0: "open",
		2: "canceled",
		3: "completed",
	}
	checklistStatuses = map[int64]string{
		0: "open",
		3: "completed",
	}
)

// TaskTypeName маппит код типа задачи в человекочитаемое имя.
// Возвращает nil для nil-входа и неизвестного кода.
func TaskTypeName(code *int64) *string {
	return lookupEnum(code, taskTypes)
}

// TaskStatusName маппит код статуса задачи в человекочитаемое имя.
func TaskStatusName(code *int64) *string {
	return lookupEnum(code, taskStatuses)
}

// ChecklistStatusName маппит код статуса checklist-айтема в имя.
func ChecklistStatusName(code *int64) *string {
	return lookupEnum(code, checklistStatuses)
}

func lookupEnum(code *int64, m map[int64]string) *string {
	if code == nil {
		return nil
	}
	if name, ok := m[*code]; ok {
		return &name
	}
	return nil
}
