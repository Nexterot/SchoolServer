// Copyright (C) 2018 Mikhail Masyagin

package sessions

import (
	"fmt"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/data-types"
	t01 "github.com/masyagin1998/SchoolServer/libtelco/sessions/type-01"

	"github.com/pkg/errors"
)

/*
Получение отчетов.
*/

/*
01 тип.
*/

// GetTotalMarkReport возвращает успеваемость ученика.
func (s *Session) GetTotalMarkReport(studentID string) (*dt.TotalMarkReport, error) {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	if studentID == "" {
		studentID = s.Base.Child.SID
	}
	var err error
	var finalMarkReport *dt.TotalMarkReport
	switch s.Base.Serv.Type {
	case cp.FirstType:
		finalMarkReport, err = t01.GetTotalMarkReport(s.Base, studentID)
		err = errors.Wrap(err, "type-01")
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return finalMarkReport, errors.Wrap(err, "from GetTotalMarkReport")
}

/*
02 тип.
*/

// GetAverageMarkReport возвращает средние баллы ученика.
func (s *Session) GetAverageMarkReport(dateBegin, dateEnd, Type, studentID string) (*dt.AverageMarkReport, error) {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	if studentID == "" {
		studentID = s.Base.Child.SID
	}
	var err error
	var averageMarkReport *dt.AverageMarkReport
	switch s.Base.Serv.Type {
	case cp.FirstType:
		averageMarkReport, err = t01.GetAverageMarkReport(s.Base, dateBegin, dateEnd, Type, studentID)
		err = errors.Wrap(err, "type-01")
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return averageMarkReport, errors.Wrap(err, "from GetAverageMarkReport")
}

/*
03 тип.
*/

// GetAverageMarkDynReport возвращает динамику среднего балла ученика.
func (s *Session) GetAverageMarkDynReport(dateBegin, dateEnd, Type, studentID string) (*dt.AverageMarkDynReport, error) {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	if studentID == "" {
		studentID = s.Base.Child.SID
	}
	var err error
	var averageMarkDynReport *dt.AverageMarkDynReport
	switch s.Base.Serv.Type {
	case cp.FirstType:
		averageMarkDynReport, err = t01.GetAverageMarkDynReport(s.Base, dateBegin, dateEnd, Type, studentID)
		err = errors.Wrap(err, "type-01")
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return averageMarkDynReport, errors.Wrap(err, "from GetAverageMarkReportDyn")
}

/*
04 тип.
*/

// GetStudentGradeReport возвращает отчет об успеваемости ученика по предмету.
func (s *Session) GetStudentGradeReport(dateBegin, dateEnd, subjectID, studentID string) (*dt.StudentGradeReport, error) {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	if studentID == "" {
		studentID = s.Base.Child.SID
	}
	var err error
	var studentGradeReport *dt.StudentGradeReport
	switch s.Base.Serv.Type {
	case cp.FirstType:
		studentGradeReport, err = t01.GetStudentGradeReport(s.Base, dateBegin, dateEnd, subjectID, studentID)
		err = errors.Wrap(err, "type-01")
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return studentGradeReport, errors.Wrap(err, "from GetStudentGradeReport")
}

/*
05 тип.
*/

// GetStudentTotalReport возвращает отчет о посещениях ученика.
func (s *Session) GetStudentTotalReport(dateBegin, dateEnd, studentID string) (*dt.StudentTotalReport, error) {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	if studentID == "" {
		studentID = s.Base.Child.SID
	}
	var err error
	var studentTotalReport *dt.StudentTotalReport
	switch s.Base.Serv.Type {
	case cp.FirstType:
		studentTotalReport, err = t01.GetStudentTotalReport(s.Base, dateBegin, dateEnd, studentID)
		err = errors.Wrap(err, "type-01")
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return studentTotalReport, errors.Wrap(err, "from GetStudentTotalReport")
}

/*
06 тип.
*/

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// GetStudentAttendanceGradeReport - отчет шестого типа пока что пропускаем.
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

/*
07 тип.
*/

// GetJournalAccessReport возвращает отчет о доступе к журналу.
func (s *Session) GetJournalAccessReport(studentID string) (*dt.JournalAccessReport, error) {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	if studentID == "" {
		studentID = s.Base.Child.SID
	}
	var err error
	var studentTotalReport *dt.JournalAccessReport
	switch s.Base.Serv.Type {
	case cp.FirstType:
		studentTotalReport, err = t01.GetJournalAccessReport(s.Base, studentID)
		err = errors.Wrap(err, "type-01")
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return studentTotalReport, errors.Wrap(err, "from GetJournalAccessReport")
}

/*
08 тип.
*/

// GetParentInfoLetterReport возвращает шаблон письма родителям.
func (s *Session) GetParentInfoLetterReport(reportTypeID, periodID, studentID string) (*dt.ParentInfoLetterReport, error) {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	if studentID == "" {
		studentID = s.Base.Child.SID
	}
	var err error
	var parentInfoLetterRepport *dt.ParentInfoLetterReport
	switch s.Base.Serv.Type {
	case cp.FirstType:
		parentInfoLetterRepport, err = t01.GetParentInfoLetterReport(s.Base, reportTypeID, periodID, studentID)
		err = errors.Wrap(err, "type-01")
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return parentInfoLetterRepport, errors.Wrap(err, "from GetParentInfoLetterReport")
}
