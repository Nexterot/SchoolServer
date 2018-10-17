// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package schoolmarks

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"

	"github.com/pkg/errors"

	gr "github.com/levigross/grequests"
	"golang.org/x/net/html"
)

// GetWeekSchoolMarks возвращает оценки на заданную неделю с сервера первого типа.
func GetWeekSchoolMarks(s *dt.Session, date, studentID string) (*dt.WeekSchoolMarks, error) {
	p := "http://"

	// 0-ой Post-запрос.
	r0 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"Date":      date,
				"LoginType": "0",
				"PCLID_IUP": "",
				"SID":       studentID,
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Curriculum/Assignments.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Curriculum/Assignments.asp", ro)
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
	// находящуюся в теле ответа, и найти в ней оценки на заданную неделю.
	parsedHTML, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	// Объявляем нужные функции

	// Находит node, в котором находятся оценки.
	var searchForSchoolMarksNode func(*html.Node) *html.Node
	searchForSchoolMarksNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			for _, a := range node.Attr {
				if a.Key == "class" && a.Val == "table table-bordered table-thin table-xs print-block" {
					return node
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := searchForSchoolMarksNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	// Находит дату дня, когда проходил урок, если этот урок является первым, и возвращает эту дату
	var newDayDate func(*html.Node) string
	newDayDate = func(node *html.Node) string {
		for _, a := range node.Attr {
			if a.Key == "title" && a.Val == "Посмотреть расписание на этот день" {
				return node.FirstChild.Data
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := newDayDate(c)
			if n != "" {
				return n
			}
		}
		return ""
	}

	// Находит ID урока
	findID := func(str string) (int, int, int, error) {
		var start int
		var AID, CID, TP int
		for i := 0; i < len(str); i++ {
			if unicode.IsDigit(rune(str[i])) {
				start = i
				i++
				for unicode.IsDigit(rune(str[i])) {
					i++
				}
				var err error
				AID, err = strconv.Atoi(str[start:i])
				if err != nil {
					return 0, 0, 0, err
				}
				i++
				start = i
				for unicode.IsDigit(rune(str[i])) {
					i++
				}
				CID, err = strconv.Atoi(str[start:i])
				if err != nil {
					return 0, 0, 0, err
				}
				i++
				start = i
				for unicode.IsDigit(rune(str[i])) {
					i++
				}
				TP, err = strconv.Atoi(str[start:i])
				if err != nil {
					return 0, 0, 0, err
				}
			}
		}
		return AID, CID, TP, nil
	}

	// Получает всю информацию о уроках из переданного нода.
	getAllSchoolMarksInfo := func(node *html.Node, requestDate string) ([]dt.DaySchoolMarks, error) {
		days := make([]dt.DaySchoolMarks, 0, 7)
		if node != nil {
			lessons := make([]dt.SchoolMark, 0, 10)
			var currentDay dt.DaySchoolMarks
			date := ""
			var lesson dt.SchoolMark
			node = node.FirstChild.NextSibling
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				if len(c.Attr) == 1 {
					if c.Attr[0].Key == "bgcolor" {
						date = newDayDate(c)
						c2 := c.FirstChild.NextSibling
						if date != "" {
							// Новый день
							if currentDay.Date != "" {
								currentDay.Lessons = lessons
								days = append(days, currentDay)
							}
							currentDay = *new(dt.DaySchoolMarks)
							currentDay.Date = date
							lessons = make([]dt.SchoolMark, 0, 10)

							c2 = c2.NextSibling
						}
						lesson = *new(dt.SchoolMark)
						lesson.Date = requestDate

						if c.Attr[0].Val == "#FFFFFF" {
							lesson.InTime = true
						}
						if c2.FirstChild != nil {
							lesson.Name = c2.FirstChild.Data
						}

						c2 = c2.NextSibling.NextSibling
						c3 := c2.FirstChild
						if c3 != nil {
							lesson.Type = c3.Data
						}

						c2 = c2.NextSibling.NextSibling
						c3 = c2.FirstChild.NextSibling
						if c3 != nil {
							if c3.FirstChild != nil {
								lesson.Title = c3.FirstChild.Data
							}

							for _, a := range c3.Attr {
								if a.Key == "onclick" {
									var err error
									lesson.AID, lesson.CID, lesson.TP, err = findID(a.Val)
									if err != nil {
										return days, err
									}
									break
								}
							}
						}

						c2 = c2.NextSibling
						if c2.FirstChild != nil {
							lesson.Weight = c2.FirstChild.Data
						}
						if c2 != nil {
							c2 = c2.NextSibling
							for c2 != nil && c2.Data != "td" {
								c2 = c2.NextSibling
							}
							if c2 != nil && c2.FirstChild != nil {
								lesson.Mark = c2.FirstChild.Data
							}
						}
						lessons = append(lessons, lesson)
					}
				}
			}
			currentDay.Lessons = lessons
			days = append(days, currentDay)
			return days, nil
		}
		return days, errors.New("Node is nil in func getAllSchoolMarksInfo")
	}

	// Проверяет, является ли день выходным
	var checkWeekend func(*html.Node) *html.Node
	checkWeekend = func(node *html.Node) *html.Node {
		if node != nil {
			if strings.Contains(node.Data, "Нет заданий на этой неделе") {
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

	// Составляет таблицу с днями и их уроками
	makeWeekSchoolMarks := func(node *html.Node, requestDate string) (*dt.WeekSchoolMarks, error) {
		days := dt.NewWeekSchoolMarks()
		var err error
		lessonsNode := searchForSchoolMarksNode(node)
		if lessonsNode == nil {
			// Проверяем, является ли день выходным
			if checkWeekend(node) != nil {
				days.Data = make([]dt.DaySchoolMarks, 0)
				return days, nil
			}
		}
		days.Data, err = getAllSchoolMarksInfo(lessonsNode, requestDate)
		return days, err
	}

	return makeWeekSchoolMarks(parsedHTML, date)
}
