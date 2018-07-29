package sessions

import (
	cp "SchoolServer/libtelco/config-parser"
	dt "SchoolServer/libtelco/sessions/data-types"
	"SchoolServer/libtelco/sessions/inner"
	t01 "SchoolServer/libtelco/sessions/type-01"
	"fmt"
)

/*
Получение расписания.
*/

// GetTimeTable возвращает расписание на n дней, начиная с текущего.
func (s *Session) GetTimeTable(date string, n int, studentID string) (*dt.TimeTable, error) {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	if studentID == "" {
		studentID = s.Base.Child.SID
	}
	var err error
	var timeTable *dt.TimeTable
	if (n < 1) || (n > 7) {
		err = fmt.Errorf("Invalid days number")
		return nil, err
	}
	timeTable = &dt.TimeTable{
		Days: make([]dt.DayTimeTable, 0, n),
	}
	for i := 0; i < n; i++ {
		day, err := s.getDayTimeTable(date, studentID)
		if err != nil {
			return timeTable, err
		}
		timeTable.Days = append(timeTable.Days, *day)
		date, err = inner.IncDate(date)
		if err != nil {
			return nil, err
		}
	}
	return timeTable, err
}

// getDayTimeTable возвращает расписание на один день.
func (s *Session) getDayTimeTable(date, studentID string) (*dt.DayTimeTable, error) {
	var err error
	var dayTimeTable *dt.DayTimeTable
	switch s.Base.Serv.Type {
	case cp.FirstType:
		dayTimeTable, err = t01.GetDayTimeTable(s.Base, date, studentID)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return dayTimeTable, err
}
