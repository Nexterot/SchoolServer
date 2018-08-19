// Copyright (C) 2018 Mikhail Masyagin

package sessions

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	t01 "github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/timetable"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"

	"github.com/pkg/errors"
)

/*
Получение расписания.
*/

// GetTimeTable возвращает расписание на n дней, начиная с текущего.
func (s *Session) GetTimeTable(date string, n int, studentID string) (*dt.TimeTable, error) {
	s.MU.Lock()
	defer s.MU.Unlock()
	if studentID == "" {
		studentID = s.Child.SID
	}
	var err error
	var timeTable *dt.TimeTable
	if (n < 1) || (n > 7) {
		err = fmt.Errorf("Invalid days number %v", n)
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
		date, err = incDate(date)
		if err != nil {
			return nil, errors.Wrap(err, "incrementing date")
		}
	}
	return timeTable, nil
}

// getDayTimeTable возвращает расписание на один день.
func (s *Session) getDayTimeTable(date, studentID string) (*dt.DayTimeTable, error) {
	var err error
	var dayTimeTable *dt.DayTimeTable
	switch s.Serv.Type {
	case cp.FirstType:
		dayTimeTable, err = t01.GetDayTimeTable(&s.Session, date, studentID)
		err = errors.Wrap(err, "type-01")
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return dayTimeTable, errors.Wrap(err, "from getDayTimeTable")
}

// incDate увеличивает текущую дату на один день.
func incDate(date string) (string, error) {
	// Объявляем необходимые функции

	// Принимает дату, записанную в виде строки, и возвращает в виде int'ов день, месяц и год этой даты
	getStringDates := func(str string) (int, int, int, error) {
		s := strings.Split(str, ".")
		if len(s) != 3 {
			// Дата записана неправильно
			return 0, 0, 0, fmt.Errorf("Wrond date: %s", date)
		}
		var err error
		day := 0
		month := 0
		year := 0
		day, err = strconv.Atoi(s[0])
		if err != nil {
			return 0, 0, 0, err
		}
		month, err = strconv.Atoi(s[1])
		if err != nil {
			return 0, 0, 0, err
		}
		year, err = strconv.Atoi(s[2])
		if err != nil {
			return 0, 0, 0, err
		}
		return day, month, year, nil
	}

	// Преобразует три числа (day, month, year) в строку-дату "day.month.year"
	makeDate := func(day int, month int, year int) string {
		var str string
		if day < 10 {
			str += "0"
		}
		str += strconv.Itoa(day) + "."
		if month < 10 {
			str += "0"
		}
		str += strconv.Itoa(month) + "." + strconv.Itoa(year)
		return str
	}

	// Преобразуем дату в три int'а
	day, month, year, err := getStringDates(date)
	if err != nil {
		return "", err
	}
	if (day == 0) && (month == 0) && (year == 0) {
		return "", fmt.Errorf("Wrond date: %s", date)
	}

	// Создаём из полученных int'ов структуру time.Date и получаем следующий день
	currentDay := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	oneDayLater := currentDay.AddDate(0, 0, 1)
	year2, month2, day2 := oneDayLater.Date()

	// Преобразуем полученные числа обратно в строку
	nextDay := makeDate(day2, int(month2), year2)
	return nextDay, nil
}
