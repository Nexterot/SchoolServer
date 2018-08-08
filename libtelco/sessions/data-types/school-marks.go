// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

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
	// Для последующего AJAX-запроса.
	Date string `json:"date"`
	AID  int    `json:"AID"`
	CID  int    `json:"CID"`
	TP   int    `json:"TP"`
	// Собственно ответ.
	Status int    `json:"status"`
	InTime bool   `json:"inTime"`
	Name   string `json:"name"`
	Author string `json:"author"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	Mark   string `json:"mark"`
	Weight string `json:"weight"`
}

// LessonDescription struct содержит в себе подробности задания.
type LessonDescription struct {
	Description string `json:"description"`
	Author      string `json:"string"`
	File        string `json:"file"`
	FileName    string `json:"fileName"`
}
