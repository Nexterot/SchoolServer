package dataTypes

/*
Оценки.
*/

// WeekSchoolMarks struct содержит в себе оценки и ДЗ на текущую неделю.
type WeekSchoolMarks struct {
	Data []DaySchoolMarks `json:"days"`
}

// DaySchoolMarks struct содержит в себе оценки и ДЗ на текущий день.
type DaySchoolMarks struct {
	Date    string       `json:"date"`
	Lessons []SchoolMark `json:"lessons"`
}

// SchoolMark struct содержит в себе оценку и ДЗ по одному уроку.
type SchoolMark struct {
	AID    int
	CID    int
	TP     int
	Status int    `json:"status"`
	InTime bool   `json:"inTime"`
	Name   string `json:"name"`
	Author string `json:"author"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	Mark   string `json:"mark"`
	Weight string `json:"weight"`
	// Временный костыль.
	ID int `json:"id"`
}
