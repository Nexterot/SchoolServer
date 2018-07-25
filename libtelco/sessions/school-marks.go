package sessions

import (
	cp "SchoolServer/libtelco/config-parser"
	dt "SchoolServer/libtelco/sessions/data-types"
	t01 "SchoolServer/libtelco/sessions/type-01"
	"fmt"
)

/*
Получение оценок.
*/

// GetWeekSchoolMarks возвращает оценки на заданную неделю.
func (s *Session) GetWeekSchoolMarks(date, studentID string) (*dt.WeekSchoolMarks, error) {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	if studentID == "" {
		studentID = s.Base.ID
	}
	var err error
	var weekSchoolMarks *dt.WeekSchoolMarks
	switch s.Base.Serv.Type {
	case cp.FirstType:
		weekSchoolMarks, _, err = t01.GetWeekSchoolMarks(s.Base, date, studentID)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return weekSchoolMarks, err
}

/*
Получение подробностей урока.
*/

// GetLessonDescription вовзращает подробности урока.
func (s *Session) GetLessonDescription(date string, AID, CID, TP int, studentID string) (*dt.LessonDescription, error) {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	if studentID == "" {
		studentID = s.Base.ID
	}
	var err error
	var lessonDescription *dt.LessonDescription
	switch s.Base.Serv.Type {
	case cp.FirstType:
		lessonDescription, err = t01.GetLessonDescription(s.Base, date, AID, CID, TP, studentID)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return lessonDescription, err
}
