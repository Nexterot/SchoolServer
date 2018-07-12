// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

/*
Package sessions - данный файл содержит в себе внутренние функции обработки сайта.
*/
package sessions

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

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

// Report struct - отчёт
type Report struct {
}

// totalMarkReportParser парсит "Отчет об успеваемости".
// находится в inner-funcs, так как отчеты на всех серверах одинаковые.
func totalMarkReportParser(r io.Reader) (*TotalMarkReport, error) {
	parsedHTML, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	// Находит нод с табличкой
	var findNode func(*html.Node) *html.Node
	findNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			for _, a := range node.Attr {
				if (a.Key == "class") && (a.Val == "table-print-num") {
					return node
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findNode(c)
			if n != nil {
				return n
			}
		}
		return nil
	}

	// Формирует отчёт
	var formTotalMarkReport func(*html.Node, map[string][]int)
	formTotalMarkReport = func(node *html.Node, report map[string][]int) {
		for _, a := range node.Attr {
			if (a.Key == "class") && (a.Val == "cell-text") {
				// Нашли урок
				report[node.FirstChild.Data] = make([]int, 7)
				i := 0
				flag := false
				for c := node.NextSibling; c != nil; c = c.NextSibling {
					if c.FirstChild != nil {
						mark, err := strconv.Atoi(c.FirstChild.Data)
						if flag {
							i++
						}
						if (err != nil) || (i > 6) {
							continue
						}
						flag = true
						report[node.FirstChild.Data][i] = mark
					}
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			formTotalMarkReport(c, report)
		}
	}

	// Создаёт отчёт
	makeTotalMarkReport := func(node *html.Node) (*TotalMarkReport, error) {
		var report TotalMarkReport
		tableNode := findNode(node)
		data := make(map[string][]int)
		formTotalMarkReport(tableNode, data)
		report.Data = data

		return &report, nil
	}

	return makeTotalMarkReport(parsedHTML)
}
