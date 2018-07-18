package dataTypes

/*
Расписание.
*/

// TimeTable struct содержит в себе расписание на N дней (N = 1, 2, ..., 7).
type TimeTable struct {
	Days []DayTimeTable `json:"days"`
}

// DayTimeTable struct содержит в себе расписание на день.
type DayTimeTable struct {
	Date    string   `json:"date"`
	Lessons []Lesson `json:"lesson"`
}

// Lesson struct содержит в себе один урок.
type Lesson struct {
	Begin     string `json:"begin"`
	End       string `json:"end"`
	Name      string `json:"name"`
	ClassRoom string `json:"classroom"`
}
