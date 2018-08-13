// Copyright (C) 2018 Mikhail Masyagin

package sessions

import (
	cp "SchoolServer/libtelco/config-parser"
	red "SchoolServer/libtelco/in-memory-db"
	dt "SchoolServer/libtelco/sessions/data-types"
	t01 "SchoolServer/libtelco/sessions/type-01"
	"fmt"

	"github.com/pkg/errors"
)

/*
Получение оценок.
*/

// GetWeekSchoolMarks возвращает оценки на заданную неделю.
func (s *Session) GetWeekSchoolMarks(date, studentID string) (*dt.WeekSchoolMarks, error) {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	if studentID == "" {
		studentID = s.Base.Child.SID
	}
	var err error
	var weekSchoolMarks *dt.WeekSchoolMarks
	switch s.Base.Serv.Type {
	case cp.FirstType:
		weekSchoolMarks, err = t01.GetWeekSchoolMarks(s.Base, date, studentID)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return weekSchoolMarks, errors.Wrap(err, "from GetWeekSchoolMarks")
}

/*
Получение подробностей урока.
*/

// GetLessonDescription вовзращает подробности урока.
func (s *Session) GetLessonDescription(AID, CID, TP int, studentID, classID, serverAddr string, db *red.Database) (*dt.LessonDescription, error) {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	if studentID == "" {
		studentID = s.Base.Child.SID
	}
	var err error
	var lessonDescription *dt.LessonDescription
	switch s.Base.Serv.Type {
	case cp.FirstType:
		lessonDescription, err = t01.GetLessonDescription(s.Base, AID, CID, TP, studentID, classID, serverAddr, db)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return lessonDescription, errors.Wrap(err, "from GetLessonDescription")
}
