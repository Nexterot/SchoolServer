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
		weekSchoolMarks, err = t01.GetWeekSchoolMarks(s.Base, date, studentID)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return weekSchoolMarks, err
}
