// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package type01

import (
	dt "SchoolServer/libtelco/sessions/data-types"
	ss "SchoolServer/libtelco/sessions/session"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
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

// GetLessonDescription вовзращает подробности урока с сервера первого типа.
func GetLessonDescription(s *ss.Session, AID, CID, TP int, studentID string) (*dt.LessonDescription, error) {
	p := "http://"
	var lessonDescription *dt.LessonDescription

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
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
	response0, err := s.Sess.Post(p+s.Serv.Link+"/asp/ajax/Assignments/GetAssignmentInfo.asp", requestOptions0)
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
	// находящуюся в теле ответа, и найти в ней подробности урока.

	// Получаем таблицу с подробностями урока из полученной json-структуры
	responseMap := make(map[string]interface{})
	err = json.Unmarshal(response0.Bytes(), &responseMap)
	if err != nil {
		return nil, err
	}
	strs := responseMap["data"].(map[string]interface{})
	var responseString string
	for k, v := range strs {
		if k == "strTable" {
			responseString = v.(string)
		}
	}

	parsedHTML, err := html.Parse(strings.NewReader(responseString))
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

	// Находит из переданной строки путь к файлу и его ID
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

	formLessonDescription := func(node *html.Node) (*dt.LessonDescription, error) {
		details := *new(dt.LessonDescription)
		if node != nil {

			var authorNode *html.Node
			if node.Parent != nil && node.Parent.Parent != nil && node.Parent.Parent.Parent != nil && node.Parent.Parent.Parent.PrevSibling != nil {
				authorNode = node.Parent.Parent.Parent.PrevSibling
				if authorNode.FirstChild != nil && authorNode.FirstChild.FirstChild != nil && authorNode.FirstChild.FirstChild.NextSibling != nil {
					authorNode = authorNode.FirstChild.FirstChild.NextSibling
					if authorNode.FirstChild != nil {
						author := authorNode.FirstChild.Data
						var start, end int
						for i := 0; i < len(author); i++ {
							if author[i:i+1] == "(" {
								start = i + 1
							} else {
								if author[i:i+1] == ")" {
									end = i
								}
							}
						}
						details.Author = author[start:end]
					}
				}
			}

			tableNode := node.FirstChild.FirstChild.NextSibling.NextSibling
			commentNode := tableNode.FirstChild.NextSibling
			if commentNode.FirstChild != nil {
				commentNode = commentNode.FirstChild
				for commentNode != nil {
					if commentNode.FirstChild != nil {
						details.Description += commentNode.FirstChild.Data
					} else {
						if commentNode.Data == "br" {
							details.Description += "\n"
						} else {
							if !(len(commentNode.Data) == 2 && commentNode.Data[0] == 194 && commentNode.Data[1] == 160) {
								details.Description += commentNode.Data
							}
						}
					}

					commentNode = commentNode.NextSibling
				}
			}

			tableNode = tableNode.NextSibling
			if tableNode.FirstChild.NextSibling.FirstChild != nil {
				if tableNode.FirstChild.NextSibling.FirstChild.FirstChild != nil {
					if tableNode.FirstChild.NextSibling.FirstChild.FirstChild.FirstChild != nil {
						details.FileName = tableNode.FirstChild.NextSibling.FirstChild.FirstChild.FirstChild.Data
					}
					for _, a := range tableNode.FirstChild.NextSibling.FirstChild.FirstChild.Attr {
						if a.Key == "href" {
							details.File, details.ID, err = findURLAndID(a.Val)
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

		return &details, errors.New("Node is nil in func formLessonDescription")
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

// getFile выкачивает файл по заданной ссылке в заданную директорию (если его там ещё нет) и возвращает
// - true, если файл был скачан;
// - false, если файл уже был в директории;
// с сервера первого типа.
func getFile(s *ss.Session, link, path, filename string, attachmentID string) (bool, error) {
	p := "http://"

	// Проверка, есть ли файл на диске.
	// Если есть, то не будем ничего скачивать.
	if _, err := os.Stat(path + filename); err == nil {
		return false, nil
	}
	// 0-ой POST-запрос.
	requestOptions0 := &gr.RequestOptions{
		Data: map[string]string{
			"VER":          s.VER,
			"at":           s.AT,
			"attachmentId": attachmentID,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Curriculum/Assignments.asp",
		},
	}
	response0, err := s.Sess.Post(link, requestOptions0)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = response0.Close()
	}()
	if err := checkResponse(s, response0); err != nil {
		return false, err
	}
	// Сохранение файла на диск.
	// Создание папок, если надо.
	if err = os.MkdirAll(path, 0700); err != nil {
		return false, err
	}
	// Создание файла.
	file, err := os.Create(path + filename)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = file.Close()
	}()
	// Запись в файл.
	if _, err = file.Write(response0.Bytes()[:len(response0.Bytes())]); err != nil {
		return false, err
	}
	return true, nil
}
