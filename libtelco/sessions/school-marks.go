// Copyright (C) 2018 Mikhail Masyagin

package sessions

import (
	"fmt"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	red "github.com/masyagin1998/SchoolServer/libtelco/in-memory-db"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	t01 "github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/schoolmarks"

	"github.com/pkg/errors"
)

/*
Получение оценок.
*/

// GetWeekSchoolMarks возвращает оценки на заданную неделю.
func (s *Session) GetWeekSchoolMarks(date, studentID string) (*dt.WeekSchoolMarks, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if studentID == "" {
		studentID = s.Child.SID
	}
	var err error
	var weekSchoolMarks *dt.WeekSchoolMarks
	switch s.Serv.Type {
	case cp.FirstType:
		weekSchoolMarks, err = t01.GetWeekSchoolMarks(&s.Session, date, studentID)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return weekSchoolMarks, errors.Wrap(err, "from GetWeekSchoolMarks")
}

/*
Получение подробностей урока.
*/

// GetLessonDescription вовзращает подробности урока.
func (s *Session) GetLessonDescription(AID, CID, TP int, studentID, classID, serverAddr string, db *red.Database) (*dt.LessonDescription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if studentID == "" {
		studentID = s.Child.SID
	}
	var err error
	var lessonDescription *dt.LessonDescription
	switch s.Serv.Type {
	case cp.FirstType:
		lessonDescription, err = t01.GetLessonDescription(&s.Session, AID, CID, TP, studentID, classID, serverAddr, db)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return lessonDescription, errors.Wrap(err, "from GetLessonDescription")
}
