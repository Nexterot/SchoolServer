// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package type01

import (
	dt "SchoolServer/libtelco/sessions/data-types"
	ss "SchoolServer/libtelco/sessions/session"
	"bytes"
	"fmt"
	"strconv"
	"unicode"

	gr "github.com/levigross/grequests"
	"golang.org/x/net/html"
)

// GetWeekSchoolMarks возвращает оценки на заданную неделю с сервера первого типа.
func GetWeekSchoolMarks(s *ss.Session, date, studentID string) (*dt.WeekSchoolMarks, error) {
	p := "http://"
	var weekSchoolMarks *dt.WeekSchoolMarks

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
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
	response0, err := s.Sess.Post(p+s.Serv.Link+"/asp/Curriculum/Assignments.asp", requestOptions0)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response0.Close()
	}()
	if err := checkResponse(s, response0); err != nil {
		return nil, err
	}
	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней оценки на заданную неделю.
	parsedHTML, err := html.Parse(bytes.NewReader(response0.Bytes()))
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
	getAllSchoolMarksInfo := func(node *html.Node) ([]dt.DaySchoolMarks, error) {
		if node != nil {
			days := make([]dt.DaySchoolMarks, 0, 7)
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
							lesson.Mark = c2.FirstChild.Data
						}
						lessons = append(lessons, lesson)
					}
				}
			}
			currentDay.Lessons = lessons
			days = append(days, currentDay)
			return days, nil
		}
		return nil, nil
	}

	// Составляет таблицу с днями и их уроками
	makeWeekSchoolMarks := func(node *html.Node) (*dt.WeekSchoolMarks, error) {
		var days dt.WeekSchoolMarks
		var err error
		lessonsNode := searchForSchoolMarksNode(node)
		days.Data, err = getAllSchoolMarksInfo(lessonsNode)
		return &days, err
	}

	weekSchoolMarks, err = makeWeekSchoolMarks(parsedHTML)
	return weekSchoolMarks, err
}

// GetLessonDescription вовзращает подробности урока с сервера первого типа.
func GetLessonDescription(s *ss.Session, date, lessonID, studentID string) (*dt.LessonDescription, error) {
	p := "http://"
	//var lessonDescription *dt.LessonDescription

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
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
	response0, err := s.Sess.Post(p+s.Serv.Link+"/asp/Curriculum/Assignments.asp", requestOptions0)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response0.Close()
	}()
	fmt.Println(string(response0.Bytes()))
	fmt.Println()
	fmt.Println()
	fmt.Println()
	if err := checkResponse(s, response0); err != nil {
		return nil, err
	}

	// 1-ый Post-запрос.
	requestOptions1 := &gr.RequestOptions{
		Data: map[string]string{
			// TODO: get AID.
			"AID":       "217128",
			"AT":        s.AT,
			"CID":       lessonID,
			"PCLID_IUP": "",
			"TP":        "3",
		},
		Headers: map[string]string{
			"Origin":           p + s.Serv.Link,
			"X-Requested-With": "XMLHttpRequest",
			"at":               s.AT,
			"Referer":          p + s.Serv.Link + "/asp/Curriculum/Assignments.asp",
		},
	}
	response1, err := s.Sess.Post(p+s.Serv.Link+"/asp/ajax/Assignments/GetAssignmentInfo.asp", requestOptions1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()
	fmt.Println(string(response1.Bytes()))
	if err := checkResponse(s, response1); err != nil {
		return nil, err
	}
	return nil, nil
}
