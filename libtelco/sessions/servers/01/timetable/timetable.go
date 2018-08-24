// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package timetable

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"

	gr "github.com/levigross/grequests"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

// GetDayTimeTable возвращает расписание на один день c сервера первого типа.
func GetDayTimeTable(s *dt.Session, date, studentID string) (*dt.DayTimeTable, error) {
	p := "http://"
	var dayTimeTable *dt.DayTimeTable

	// 0-ой Post-запрос.
	r0 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"BackPage":  "/asp/Calendar/DayViewS.asp",
				"DATE":      date,
				"EventID":   "",
				"EventType": "",
				"FOO":       "",
				"LoginType": "0",
				"PCLID_UP":  "",
				"SID":       studentID,
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Referer": p + s.Serv.Link + "/asp/Calendar/DayViewS.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Calendar/DayViewS.asp", ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := check.CheckResponse(s, r)
		return r.Bytes(), flag, err
	}
	b, flag, err := r0()
	if err != nil {
		return nil, errors.Wrap(err, "0 POST")
	}
	if !flag {
		b, flag, err = r0()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 0 POST")
		}
		if !flag {
			return nil, fmt.Errorf("retry didn't work for 0 POST")
		}
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней расписание на текущий день.
	parsedHTML, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	// Объявляем нужные функции
	// Получение даты дня, расписание которого мы парсим.
	var getDate func(*html.Node) string
	getDate = func(node *html.Node) string {
		if node.Type == html.ElementNode {
			for _, a := range node.Attr {
				if a.Key == "class" && a.Val == "form-control date-input" {
					for _, a2 := range node.Attr {
						if a2.Key == "value" {
							return a2.Val

						}
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if str := getDate(c); str != "" {
				return str
			}
		}
		return ""
	}

	// Находит node, в котором находится расписание.
	var searchForLessonsNode func(*html.Node) *html.Node
	searchForLessonsNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			for _, a := range node.Attr {
				if a.Key == "class" && a.Val == "schedule-table table table-bordered table-condensed print-block" {
					return node
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := searchForLessonsNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	// Проверяет, является ли строка временем урока.
	isDate := func(str string) bool {
		return len(strings.Split(str, ":")) == 3
	}

	// Получает всю информацию о уроках из переданного нода.
	var getAllLessonsInfo func(*html.Node, *[]string, *[]string, *[]string, *[]string)
	getAllLessonsInfo = func(node *html.Node, starts *[]string, ends *[]string, names *[]string, classrooms *[]string) {
		if node != nil {
			if isDate(node.Data) {
				// Нашли строку, содержащую время некоторого урока.
				str := node.Data

				// Находим начало и конец урока в этой строке.
				var start string
				var end string
				if unicode.IsDigit(rune(str[4])) {
					start = str[:5]
				} else {
					start = str[:4]
				}
				if unicode.IsDigit(rune(str[len(str)-5])) {
					end = str[len(str)-5:]
				} else {
					end = str[len(str)-4:]
				}

				*starts = append(*starts, start)
				*ends = append(*ends, end)
			} else if strings.Contains(node.Data, "Урок: ") {
				// Нашли предмет и кабинет, в котором проходит урок.
				s := strings.Split(node.Data, "[")
				*names = append(*names, (s[0])[10:])
				*classrooms = append(*classrooms, "["+s[1])
			}
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				getAllLessonsInfo(c, starts, ends, names, classrooms)
			}
		}
	}

	// Проверяет, является ли день выходным
	var checkWeekend func(*html.Node) *html.Node
	checkWeekend = func(node *html.Node) *html.Node {
		if node != nil {
			if strings.Contains(node.Data, "Нет занятий на этот день для выбранных условий") {
				return node
			}
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				n := checkWeekend(c)
				if n != nil {
					return n
				}
			}
		}

		return nil
	}

	// Получает уроки.
	getLessons := func(node *html.Node) []dt.Lesson {
		// Максимальное количество уроков.
		n := 10

		lessonsNode := searchForLessonsNode(node)
		if lessonsNode == nil {
			// Проверяем, является ли день выходным
			if checkWeekend(node) != nil {
				lessons := make([]dt.Lesson, 1)
				lessons[0].Begin = "00:00"
				lessons[0].End = "23:59"
				lessons[0].Name = "Выходной день"
				return lessons
			}
		}

		starts := make([]string, 0, n)
		ends := make([]string, 0, n)
		names := make([]string, 0, n)
		classrooms := make([]string, 0, n)

		getAllLessonsInfo(lessonsNode, &starts, &ends, &names, &classrooms)

		lessons := make([]dt.Lesson, 0, len(starts))
		for i := 0; i < len(starts); i++ {
			lessons = append(lessons, *new(dt.Lesson))
		}

		if len(starts) != 0 && len(names) == 0 && len(classrooms) == 0 {
			// Случай, когда в расписании указаны каникулы или праздник
			eventNode := lessonsNode.FirstChild.NextSibling.FirstChild.NextSibling
			for i := 0; i < len(starts) && eventNode != nil && eventNode.FirstChild != nil; i++ {
				infoNode := eventNode.FirstChild
				lessons[i].Begin = "00:00"
				lessons[i].End = "23:59"
				infoNode = infoNode.NextSibling.FirstChild
				var event string
				if infoNode != nil && infoNode.NextSibling.FirstChild != nil {
					event = infoNode.NextSibling.FirstChild.Data
				}

				lessons[i].Name = event

				eventNode = eventNode.NextSibling
			}

		} else {
			for i := 0; i < len(starts); i++ {
				lessons[i].Begin = starts[i]
				lessons[i].End = ends[i]
				lessons[i].Name = names[i]
				lessons[i].ClassRoom = classrooms[i]
			}
		}

		return lessons
	}

	// Составляем расписание дня из распарсенного html-кода
	makeDayTimeTable := func(node *html.Node) *dt.DayTimeTable {
		var day dt.DayTimeTable
		day.Date = getDate(node)
		day.Lessons = getLessons(node)
		return &day
	}

	dayTimeTable = makeDayTimeTable(parsedHTML)
	return dayTimeTable, nil
}
