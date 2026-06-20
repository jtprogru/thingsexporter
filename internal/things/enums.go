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

// TaskTypeName maps a task type code to a human-readable name.
// Returns nil for nil input and unknown codes.
func TaskTypeName(code *int64) *string {
	return lookupEnum(code, taskTypes)
}

// TaskStatusName maps a task status code to a human-readable name.
func TaskStatusName(code *int64) *string {
	return lookupEnum(code, taskStatuses)
}

// ChecklistStatusName maps a checklist item status code to a name.
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
