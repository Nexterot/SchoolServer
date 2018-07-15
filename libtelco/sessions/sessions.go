// Copyright (C) 2018 Mikhail Masyagin

/*
Package sessions - данный файл содержит в себе сессии на серверах школ.
*/
package sessions

import (
	cp "SchoolServer/libtelco/config-parser"
	"fmt"
	"sync"
	"time"

	gr "github.com/levigross/grequests"
)

const (
	parent = iota
	child  = iota
)

// Session struct содержит в себе описание сессии к одному из школьных серверов.
type Session struct {
	// Общая структура.
	sess        *gr.Session
	Serv        *cp.School
	mu          sync.Mutex
	LastRequest time.Time
	Type        int
	// Только для родителей.
	ChildrenIDS *map[string]string
	// Для серверов первого типа.
	at  string
	ver string
}

// NewSession создает новую сессию на базе информации о школьном сервере,
// к которому предстоит подключиться.
func NewSession(server *cp.School) *Session {
	return &Session{
		sess: gr.NewSession(nil),
		Serv: server,
		mu:   sync.Mutex{},
	}
}

/*
Вход в систему.
*/

// Login логинится к серверу и создает очередную сессию.
func (s *Session) Login() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	switch s.Serv.Type {
	case cp.FirstType:
		err = s.loginFirst()
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return err
}

/*
Получение списка детей.
*/

// GetChildrenMap получает мапу детей в их {UID, CLID}.
func (s *Session) GetChildrenMap() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	switch s.Serv.Type {
	case cp.FirstType:
		err = s.getChildrenMapFirst()
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return err
}

/*
Получение расписания.
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

// GetTimeTable возвращает расписание на n дней, начиная с текущего.
func (s *Session) GetTimeTable(date string, n int) (*TimeTable, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var timeTable *TimeTable
	if (n < 1) || (n > 7) {
		err = fmt.Errorf("Invalid days number")
		return nil, err
	}
	timeTable = &TimeTable{
		Days: make([]DayTimeTable, 0, n),
	}
	for i := 0; i < n; i++ {
		day, err := s.getDayTimeTable(date)
		if err != nil {
			return timeTable, err
		}
		timeTable.Days = append(timeTable.Days, *day)
		date, err = incDate(date)
		if err != nil {
			return nil, err
		}
	}
	return timeTable, err
}

// getDayTimeTable возвращает расписание на один день.
func (s *Session) getDayTimeTable(date string) (*DayTimeTable, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var dayTimeTable *DayTimeTable
	switch s.Serv.Type {
	case cp.FirstType:
		dayTimeTable, err = s.getDayTimeTableFirst(date)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return dayTimeTable, err
}

/*
Получение оценок.
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
	Status bool   `json:"statuc"`
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

// GetWeekSchoolMarks возвращает оценки на заданную неделю.
func (s *Session) GetWeekSchoolMarks(date string) (*WeekSchoolMarks, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var weekSchoolMarks *WeekSchoolMarks
	switch s.Serv.Type {
	case cp.FirstType:
		weekSchoolMarks, err = s.getWeekSchoolMarksFirst(date)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return weekSchoolMarks, err
}

/*
Получение отчетов.
*/

/*
1 тип.
*/

// TotalMarkReport struct - отчет первого типа.
type TotalMarkReport struct {
	Table []TotalMarkReportNote `json:"table"`
}

// TotalMarkReportNote struct -
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

// GetTotalMarkReport возвращает успеваемость ученика.
func (s *Session) GetTotalMarkReport() (*TotalMarkReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var finalMarkReport *TotalMarkReport
	switch s.Serv.Type {
	case cp.FirstType:
		finalMarkReport, err = s.getTotalMarkReportFirst()
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return finalMarkReport, err
}

/*
2 тип.
*/

// REDO!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// AverageMarkReport struct - отчет второго типа.
type AverageMarkReport struct {
	Student map[string]string
	Class   map[string]string
}

// GetAverageMarkReport возвращает средние баллы ученика.
func (s *Session) GetAverageMarkReport(dateBegin, dateEnd, Type string) (*AverageMarkReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var averageMarkReport *AverageMarkReport
	switch s.Serv.Type {
	case cp.FirstType:
		averageMarkReport, err = s.getAverageMarkReportFirst(dateBegin, dateEnd, Type)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return averageMarkReport, err
}

/*
3 тип.
*/

// AverageMarkDynReport struct - отчет третьего типа.
type AverageMarkDynReport struct {
	Data []AverageMarkDynReportNote
}

// AverageMarkDynReportNote struct - одна запись в отчёте "Динамика среднего балла".
type AverageMarkDynReportNote struct {
	Date               string
	StudentWorksAmount int
	StudentAverageMark string
	ClassWorksAmount   int
	ClassAverageMark   string
}

// GetAverageMarkDynReport возвращает динамику среднего балла ученика.
func (s *Session) GetAverageMarkDynReport(dateBegin, dateEnd, Type string) (*AverageMarkDynReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var averageMarkDynReport *AverageMarkDynReport
	switch s.Serv.Type {
	case cp.FirstType:
		averageMarkDynReport, err = s.getAverageMarkDynReportFirst(dateBegin, dateEnd, Type)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return averageMarkDynReport, err
}

/*
4 тип.
*/

// StudentGradeReport struct - отчет четвертого типа.
type StudentGradeReport struct {
	Data []StudentGradeReportNote
}

// StudentGradeReportNote struct - одна запись в отчете об успеваемости.
type StudentGradeReportNote struct {
	Type             string
	Theme            string
	DateOfCompletion string
	Mark             int
}

// GetStudentGradeReport возвращает отчет об успеваемости ученика по предмету.
func (s *Session) GetStudentGradeReport(dateBegin, dateEnd, SubjectName string) (*StudentGradeReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var studentGradeReport *StudentGradeReport
	switch s.Serv.Type {
	case cp.FirstType:
		studentGradeReport, err = s.getStudentGradeReportFirst(dateBegin, dateEnd, SubjectName)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return studentGradeReport, err
}

/*
5 тип.
*/

// REDO!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// StudentTotalReport struct - отчет пятого типа.
type StudentTotalReport struct {
	MainTable    []Month
	AverageMarks map[string]string
}

// Day struct - один день в отчёте об успеваемости и посещаемости
type Day struct {
	Number int
	Marks  map[string][]string
}

// Month struct - один месяц в отчёте об успеваемости и посещаемости
type Month struct {
	Name string
	Days []Day
}

// GetStudentTotalReport возвращает отчет о посещениях ученика.
func (s *Session) GetStudentTotalReport(dateBegin, dateEnd string) (*StudentTotalReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var studentTotalReport *StudentTotalReport
	switch s.Serv.Type {
	case cp.FirstType:
		studentTotalReport, err = s.getStudentTotalReportFirst(dateBegin, dateEnd)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return studentTotalReport, err
}

/*
6 тип.
*/

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// StudentTotalReport struct - отчет шестого типа пока что пропускаем.
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

/*
7 тип.
*/

// JournalAccessReport struct - отчет седьмого типа.
type JournalAccessReport struct {
}

// GetJournalAccessReport возвращает отчет о доступе к журналу.
func (s *Session) GetJournalAccessReport() (*JournalAccessReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var studentTotalReport *JournalAccessReport
	switch s.Serv.Type {
	case cp.FirstType:
		studentTotalReport, err = s.getJournalAccessReportFirst()
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return studentTotalReport, err
}

/*
8 тип.
*/

// ParentInfoLetterReport struct - отчет седьмого типа.
type ParentInfoLetterReport struct {
}

// GetParentInfoLetterReport возвращает шаблон письма родителям.
func (s *Session) GetParentInfoLetterReport(studentID, reportTypeID, periodID string) (*ParentInfoLetterReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	var parentInfoLetterRepport *ParentInfoLetterReport
	switch s.Serv.Type {
	case cp.FirstType:
		parentInfoLetterRepport, err = s.getParentInfoLetterReportFirst(studentID, reportTypeID, periodID)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return parentInfoLetterRepport, err
}

// checkPage проверяет, была ли ошибка при запросе.
func (s *Session) checkResponse(response *gr.Response) error {
	switch s.Serv.Type {
	case cp.FirstType:
		return s.checkResponseFirst(response)
	default:
		return fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
}
