// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package datatypes

/*
01 тип.
*/

// TotalMarkReport struct - отчет первого типа.
type TotalMarkReport struct {
	Table []TotalMarkReportNote `json:"table"`
}

// TotalMarkReportNote struct - подотчет об одном предмете.
type TotalMarkReportNote struct {
	Subject string `json:"subject"`
	Period1 int    `json:"period1"`
	Period2 int    `json:"period2"`
	Period3 int    `json:"period3"`
	Period4 int    `json:"period4"`
	Year    int    `json:"year"`
	Exam    int    `json:"exam"`
	Final   int    `json:"final"`
}

/*
02 тип.
*/

// AverageMarkReport struct - отчет второго типа.
type AverageMarkReport struct {
	Table []AverageMarkReportNote `json:"table"`
}

// AverageMarkReportNote - подотчет об одном предмете.
type AverageMarkReportNote struct {
	Subject     string  `json:"subject"`
	StudentMark float32 `json:"student_mark"`
	ClassMark   float32 `json:"class_mark"`
}

/*
03 тип.
*/

// AverageMarkDynReport struct - отчет третьего типа.
type AverageMarkDynReport struct {
	Data []AverageMarkDynReportNote `json:"table"`
}

// AverageMarkDynReportNote struct - одна запись в отчёте "Динамика среднего балла".
type AverageMarkDynReportNote struct {
	Date               string  `json:"date"`
	StudentWorksAmount int     `json:"student_works_amount"`
	StudentAverageMark float32 `json:"student_average_mark"`
	ClassWorksAmount   int     `json:"class_works_amount"`
	ClassAverageMark   float32 `json:"class_average_mark"`
}

/*
04 тип.
*/

// StudentGradeReport struct - отчет четвертого типа.
type StudentGradeReport struct {
	Data []StudentGradeReportNote `json:"table"`
}

// StudentGradeReportNote struct - одна запись в отчете об успеваемости.
type StudentGradeReportNote struct {
	Type             string `json:"type"`
	Theme            string `json:"theme"`
	DateOfCompletion string `json:"date_of_completion"`
	Mark             int    `json:"mark"`
}

/*
05 тип.
*/

// StudentTotalReport struct - отчет пятого типа.
type StudentTotalReport struct {
	MainTable    []Month              `json:"months"`
	AverageMarks []SubjectAverageMark `json:"average_marks"`
}

type SubjectMarks struct {
	Name  string   `json:"name"`
	Marks []string `json:"marks"`
}

type Day struct {
	Number   int            `json:"number"`
	Subjects []SubjectMarks `json:"subjects"`
}

type Month struct {
	Name string `json:"name"`
	Days []Day  `json:"days"`
}

type SubjectAverageMark struct {
	Name string  `json:"name"`
	Mark float32 `json:"mark"`
}

/*
06 тип.
*/

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// StudentAttendanceGradeReport - отчет шестого типа пока что пропускаем.
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

/*
07 тип.
*/

// JournalAccessReport struct - отчет седьмого типа.
type JournalAccessReport struct {
	Data []JournalAccessReportNote `json:"table"`
}

// JournalAccessReportNote struct - одна запись в отчёте о доступе к классному журналу
type JournalAccessReportNote struct {
	Class      int    `json:"class"`
	Subject    string `json:"subject"`
	Date       string `json:"date"`
	User       string `json:"user"`
	LessonDate string `json:"lesson_date"`
	Period     string `json:"period"`
	Action     string `json:"action"`
}

/*
08 тип.
*/

// ParentInfoLetterReport struct - отчет восьмого типа.
type ParentInfoLetterReport struct {
	Data []ParentInfoLetterReportNote `json:"table"`
}

// ParentInfoLetterReportNote struct - одна запись в отчёте "Информационное письмо для родителей"
type ParentInfoLetterReportNote struct {
	Name          string  `json:"name"`
	Marks         []int   `json:"marks"`
	AverageMark   float32 `json:"average_mark"`
	MarkForPeriod int     `json:"mark_for_period"`
}
