package inner

import (
	dt "SchoolServer/libtelco/sessions/data-types"
	"errors"
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// StudentTotalReportParser возвращает отчет о посещениях ученика.
// находится в inner-funcs, так как отчеты на всех серверах одинаковые.
func StudentTotalReportParser(r io.Reader) (*dt.StudentTotalReport, error) {
	parsedHTML, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	// Находит нод с табличкой
	var findPerformanceAndAttendanceTableNode func(*html.Node) *html.Node
	findPerformanceAndAttendanceTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			for _, a := range node.Attr {
				if (a.Key == "class") && (a.Val == "table-print") {
					return node
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findPerformanceAndAttendanceTableNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	// Разделяет строку по пробелу
	splitBySpace := func(str string) []string {
		strings := make([]string, 0, 3)
		start := 0
		for i := 0; i < len(str); i++ {
			if (str[i] == 194) || (str[i] == 160) {
				if (i < len(str)-1) && (str[i+1] == 160) {
					strings = append(strings, str[start:i])
					start = i + 2
				}
			}
		}
		if start < len(str) {
			strings = append(strings, str[start:])
		}

		return strings
	}

	// Формирует отчёт
	formStudentTotalReportTable := func(node *html.Node) ([]dt.Month, []dt.SubjectAverageMark, error) {
		months := make([]dt.Month, 0, 3)
		averageMarks := make([]dt.SubjectAverageMark, 0, 10)

		if node != nil {
			// Добавляем месяцы
			monthNode := node.FirstChild.FirstChild.FirstChild
			for monthNode = monthNode.NextSibling; monthNode != nil && len(monthNode.Attr) == 1 && monthNode.Attr[0].Key == "colspan"; monthNode = monthNode.NextSibling {
				month := *new(dt.Month)
				if monthNode.FirstChild != nil {
					month.Name = monthNode.FirstChild.Data
				}
				numberOfDaysInMonth, err := strconv.Atoi(monthNode.Attr[0].Val)
				if err != nil {
					return months, averageMarks, err
				}
				month.Days = make([]dt.Day, numberOfDaysInMonth)
				months = append(months, month)
			}

			// Добавляем дни
			dayNode := node.FirstChild.FirstChild.NextSibling
			// Текущий месяц в months
			currentMonth := 0
			// Сколько дней добавили для текущего месяца
			dayNumberInMonth := 0
			// Всего дней в отчёте
			overallNumberOfDays := 0
			for dayNode = dayNode.FirstChild; dayNode != nil; dayNode = dayNode.NextSibling {
				if dayNumberInMonth == len(months[currentMonth].Days) {
					currentMonth++
					dayNumberInMonth = 0
				}

				day := *new(dt.Day)
				if dayNode.FirstChild != nil {
					day.Number, err = strconv.Atoi(dayNode.FirstChild.Data)
				}
				day.Subjects = make([]dt.SubjectMarks, 0, 1)
				if err != nil {
					return months, averageMarks, err
				}
				months[currentMonth].Days[dayNumberInMonth] = day

				dayNumberInMonth++
				overallNumberOfDays++
			}

			// Идём по остальной части таблицы
			noteNode := node.FirstChild.FirstChild.NextSibling
			for noteNode = noteNode.NextSibling; noteNode != nil; noteNode = noteNode.NextSibling {
				currentMonth = 0
				dayNumberInMonth = 0
				c := noteNode.FirstChild
				var subjectName string
				if c.FirstChild != nil {
					subjectName = c.FirstChild.Data
				}
				for i := 0; i < overallNumberOfDays; i++ {
					if dayNumberInMonth == len(months[currentMonth].Days) {
						currentMonth++
						dayNumberInMonth = 0
					}

					c = c.NextSibling
					var marks []string
					if c.FirstChild.FirstChild != nil {
						for c2 := c.FirstChild; c2 != nil; c2 = c2.NextSibling {
							var s []string
							if c2.FirstChild != nil {
								s = splitBySpace(c2.FirstChild.Data)
							} else {
								s = splitBySpace(c2.Data)
							}
							for k := 0; k < len(s); k++ {
								marks = append(marks, s[k])

							}
						}
					} else {
						if c.FirstChild != nil {
							marks = splitBySpace(c.FirstChild.Data)
						}
					}

					// Избавляемся от строк из непечатаемых символом
					finalMarks := make([]string, 0, 1)
					for el := range marks {
						if len([]byte(marks[el])) != 0 {
							finalMarks = append(finalMarks, marks[el])
						}
					}
					if len(finalMarks) != 0 {
						subjectMarks := *new(dt.SubjectMarks)
						subjectMarks.Name = subjectName
						subjectMarks.Marks = finalMarks

						months[currentMonth].Days[dayNumberInMonth].Subjects = append(months[currentMonth].Days[dayNumberInMonth].Subjects, subjectMarks)
					}

					dayNumberInMonth++
				}
				averageSubjectMark := *new(dt.SubjectAverageMark)
				averageSubjectMark.Name = subjectName
				if c.NextSibling.FirstChild != nil {
					v, err := strconv.ParseFloat(strings.Replace(c.NextSibling.FirstChild.Data, ",", ".", 1), 32)
					if err != nil {
						v = -1.0
					}
					averageSubjectMark.Mark = float32(v)
				} else {
					averageSubjectMark.Mark = -1.0
				}

				averageMarks = append(averageMarks, averageSubjectMark)
			}
		} else {
			return months, averageMarks, errors.New("Node is nil in func formStudentTotalReportTable")
		}

		return months, averageMarks, nil
	}

	// Создаёт отчёт
	makeStudentTotalReportTable := func(node *html.Node) (*dt.StudentTotalReport, error) {
		var report dt.StudentTotalReport
		tableNode := findPerformanceAndAttendanceTableNode(node)
		report.MainTable, report.AverageMarks, err = formStudentTotalReportTable(tableNode)

		return &report, err
	}

	var report *dt.StudentTotalReport
	report, err = makeStudentTotalReportTable(parsedHTML)
	if err != nil {
		return nil, err
	}

	return report, nil
}
