// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

/*
Package inner - данный файл содержит в себе внутренние функции обработки сайта.
*/
package inner

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// IncDate увеличивает текущую дату на один день.
func IncDate(date string) (string, error) {
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
