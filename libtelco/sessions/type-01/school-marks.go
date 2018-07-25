// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package type01

import (
	dt "SchoolServer/libtelco/sessions/data-types"
	ss "SchoolServer/libtelco/sessions/session"
	"bytes"
	"errors"
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
		return days, errors.New("Node is nil in func getAllSchoolMarksInfo")
	}

	// Составляет таблицу с днями и их уроками
	makeWeekSchoolMarks := func(node *html.Node, requestDate string) (*dt.WeekSchoolMarks, error) {
		var days dt.WeekSchoolMarks
		var err error
		lessonsNode := searchForSchoolMarksNode(node)
		days.Data, err = getAllSchoolMarksInfo(lessonsNode, requestDate)
		return &days, err
	}

	weekSchoolMarks, err = makeWeekSchoolMarks(parsedHTML, date)
	return weekSchoolMarks, err
}

/*
TODO:
1) чтобы вытащить аякс-запросом подробности урока, мне нужно сначала запросить расписание+оценки, которые будут содержать данный урок, то есть мне нужна "дата". Таким образом, в ответ со списком оценок/уроков тебе надо пихнуть тупо дату, которую передавал пользователь
2) так как там идёт Ajax-запрос, то там не html, а маленькая фигня прилетает, твой парсер эт учитывает?
*/

// GetLessonDescription вовзращает подробности урока с сервера первого типа.
func GetLessonDescription(s *ss.Session, date string, AID, CID, TP int, studentID string) (*dt.LessonDescription, error) {
	p := "http://"
	var lessonDescription *dt.LessonDescription

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

	// 1-ый Post-запрос.
	requestOptions1 := &gr.RequestOptions{
		Data: map[string]string{
			"AID":       strconv.Itoa(AID),
			"AT":        s.AT,
			"CID":       strconv.Itoa(CID),
			"PCLID_IUP": "",
			"TP":        strconv.Itoa(TP),
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
	if err := checkResponse(s, response1); err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней подробности урока.
	parsedHTML, err := html.Parse(bytes.NewReader(response1.Bytes()))
	if err != nil {
		return nil, err
	}

	// Находит нод с табличкой
	var findLessonDescriptionTableNode func(*html.Node) *html.Node
	findLessonDescriptionTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "table" {
				for _, a := range node.Attr {
					if a.Key == "class" && a.Val == "table table-bordered table-condensed" {
						return node
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findLessonDescriptionTableNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	// Находит из переданной строки путь к файлу и его attachmentId
	findURLAndID := func(str string) (string, string, error) {
		var url, id string
		var i, j int
		for i = 0; i < len(str); i++ {
			if str[i:i+1] == "(" {
				break
			}
		}
		for j = i; j < len(str); j++ {
			if str[j:j+1] == "," {
				url = str[i+2 : j-1]
				break
			}
		}
		for i = j + 1; i < len(str); i++ {
			if str[i:i+1] == ")" {
				id = str[j+2 : i]
			}
		}
		if url == "" || id == "" {
			return url, id, errors.New("Couldn't find url path or attachment id of file in task details")
		}

		return url, id, nil
	}

	var formLessonDescription func(*html.Node) (*dt.LessonDescription, error)
	formLessonDescription = func(node *html.Node) (*dt.LessonDescription, error) {
		if node != nil {
			details := *new(dt.LessonDescription)

			tableNode := node.FirstChild.FirstChild
			if tableNode.FirstChild.FirstChild != nil {
				details.ThemeType = tableNode.FirstChild.FirstChild.Data
			}
			if tableNode.FirstChild.NextSibling.FirstChild != nil {
				details.ThemeInfo = tableNode.FirstChild.NextSibling.FirstChild.Data
			}

			tableNode = tableNode.NextSibling
			if tableNode.FirstChild.FirstChild != nil {
				details.DateType = tableNode.FirstChild.FirstChild.Data
			}
			if tableNode.FirstChild.NextSibling.FirstChild != nil {
				details.DateInfo = tableNode.FirstChild.NextSibling.FirstChild.Data
			}

			details.Comments = make([]string, 0, 1)
			tableNode = tableNode.NextSibling
			commentNode := tableNode.FirstChild.NextSibling
			if commentNode.FirstChild != nil {
				commentNode = commentNode.FirstChild
				for commentNode != nil {
					if commentNode.Data != "br" && !(len(commentNode.Data) == 2 && commentNode.Data[0] == 194 && commentNode.Data[1] == 160) {
						details.Comments = append(details.Comments, commentNode.Data)
					}
					commentNode = commentNode.NextSibling
				}
			}

			tableNode = tableNode.NextSibling
			if tableNode.FirstChild.NextSibling.FirstChild != nil {
				if tableNode.FirstChild.NextSibling.FirstChild.FirstChild != nil {
					for _, a := range tableNode.FirstChild.NextSibling.FirstChild.FirstChild.Attr {
						if a.Key == "href" {
							details.File, details.AttachmentID, err = findURLAndID(a.Val)
							if err != nil {
								return &details, err
							}
							break
						}
					}
				}
			}

			return &details, nil
		}

		return nil, errors.New("Node is nil in func formLessonDescription")
	}

	makeLessonDescription := func(node *html.Node) (*dt.LessonDescription, error) {
		var details *dt.LessonDescription
		tableNode := findLessonDescriptionTableNode(node)
		details, err = formLessonDescription(tableNode)

		return details, err
	}

	lessonDescription, err = makeLessonDescription(parsedHTML)
	return lessonDescription, err
}
