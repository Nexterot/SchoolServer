// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

/*
Package sessions - данный файл содержит в себе функции для обработки 1 типа сайтов.
*/
package sessions

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"

	gr "github.com/levigross/grequests"
	"golang.org/x/net/html"
)

// firstTypeLogin логинится к серверу первого типа и создает очередную сессию.
func (s *Session) firstTypeLogin() error {
	// Создание сессии.
	s.sess = gr.NewSession(nil)
	p := "http://"

	// Полчение формы авторизации.
	// 0-ой Get-запрос.
	response0, err := s.sess.Get(p+s.Serv.Link+"/asp/ajax/getloginviewdata.asp", nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = response0.Close()
	}()

	// 1-ый Post-запрос.
	response1, err := s.sess.Post(p+s.Serv.Link+"/webapi/auth/getdata", nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = response1.Close()
	}()
	type FirstAnswer struct {
		Lt   string `json:"lt"`
		Ver  string `json:"ver"`
		Salt string `json:"salt"`
	}
	fa := &FirstAnswer{}
	if err = response1.JSON(fa); err != nil {
		return err
	}

	// 2-ой Post-запрос.
	pw := s.Serv.Password
	hasher := md5.New()
	if _, err = hasher.Write([]byte(pw)); err != nil {
		return err
	}
	pw = hex.EncodeToString(hasher.Sum(nil))
	hasher = md5.New()
	if _, err = hasher.Write([]byte(fa.Salt + pw)); err != nil {
		return err
	}
	pw = hex.EncodeToString(hasher.Sum(nil))
	requestOption := &gr.RequestOptions{
		Data: map[string]string{
			"CID":       "2",
			"CN":        "1",
			"LoginType": "1",
			"PID":       "-1",
			"PW":        pw[:len(s.Serv.Password)],
			"SCID":      "2",
			"SFT":       "2",
			"SID":       "77",
			"UN":        s.Serv.Login,
			"lt":        fa.Lt,
			"pw2":       pw,
			"ver":       fa.Ver,
		},
		Headers: map[string]string{
			"Referer": p + s.Serv.Link + "/",
		},
	}
	response2, err := s.sess.Post(p+s.Serv.Link+"/asp/postlogin.asp", requestOption)
	if err != nil {
		return err
	}
	defer func() {
		_ = response2.Close()
	}()

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа и найти в ней "AT".
	parsedHTML, err := html.Parse(bytes.NewReader(response2.Bytes()))
	if err != nil {
		return err
	}
	var f func(*html.Node, string) string
	f = func(node *html.Node, reqAttr string) string {
		if node.Type == html.ElementNode {
			for i := 0; i < len(node.Attr)-1; i++ {
				tmp0 := node.Attr[i]
				tmp1 := node.Attr[i+1]
				if (tmp0.Key == "name") && (tmp0.Val == reqAttr) && (tmp1.Key == "value") {
					return tmp1.Val
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if str := f(c, reqAttr); str != "" {
				return str
			}
		}
		return ""
	}
	s.at = f(parsedHTML, "AT")
	s.ver = f(parsedHTML, "VER")
	if (s.at == "") || (s.ver == "") {
		return fmt.Errorf("Problems on school server: %s", s.Serv.Link)
	}

	return nil
}

func (s *Session) getDayTimeTableFirst(date string) (*DayTimeTable, error) {
	p := "http://"
	var dayTimeTable *DayTimeTable

	// 0-ой Get-запрос.
	RequestOption0 := &gr.RequestOptions{
		Headers: map[string]string{
			"at":      s.at,
			"Referer": p + s.Serv.Link + "/asp/Calendar/DayViewS.asp",
		},
	}
	_, err := s.sess.Get(p+s.Serv.Link+"/asp/ajax/GetCalendar.asp?AT="+s.at+"&startDate=01.09.2017&endDate=31.08.2018", RequestOption0)
	if err != nil {
		return nil, err
	}

	// 1-ый Post-запрос.
	requestOption1 := &gr.RequestOptions{
		Data: map[string]string{
			"AT":        s.at,
			"BackPage":  "/asp/Calendar/DayViewS.asp",
			"DATE":      date,
			"EventID":   "",
			"EventType": "",
			"FOO":       "",
			"LoginType": "0",
			"PCLID_UP":  "10169_0",
			"SID":       "11198",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Referer": p + s.Serv.Link + "/asp/Calendar/DayViewS.asp",
		},
	}
	response, err := s.sess.Post(p+s.Serv.Link+"/asp/Calendar/DayViewS.asp", requestOption1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Close()
	}()

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа и найти в ней расписание на текущий день.
	parsedHTML, err := html.Parse(bytes.NewReader(response.Bytes()))
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

	// Получает уроки.
	getLessons := func(node *html.Node) []Lesson {
		// Максимальное количество уроков.
		n := 10

		lessonsNode := searchForLessonsNode(node)

		starts := make([]string, 0, n)
		ends := make([]string, 0, n)
		names := make([]string, 0, n)
		classrooms := make([]string, 0, n)

		getAllLessonsInfo(lessonsNode, &starts, &ends, &names, &classrooms)

		lessons := make([]Lesson, 0, len(starts))
		for i := 0; i < len(starts); i++ {
			lessons = append(lessons, *new(Lesson))
		}

		for i := 0; i < len(starts); i++ {
			lessons[i].Begin = starts[i]
			lessons[i].End = ends[i]
			lessons[i].Name = names[i]
			lessons[i].ClassRoom = classrooms[i]
		}

		return lessons
	}

	// Составляем расписание дня из распарсенного html-кода
	makeDayTimeTable := func(node *html.Node) *DayTimeTable {
		var day DayTimeTable
		day.Date = getDate(node)
		day.Lessons = getLessons(node)
		return &day
	}

	dayTimeTable = makeDayTimeTable(parsedHTML)
	return dayTimeTable, nil
}

func (s *Session) getFirstFinalMarkReport() (*FinalMarkReport, error) {
	p := "http://"
	//var finalMarkReport *FinalMarkReport

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{}
	_, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentTotalMarks.asp", requestOptions0)
	if err != nil {
		return nil, err
	}

	// 1-ый Post-запрос.
	requestOption1 := &gr.RequestOptions{
		Data: map[string]string{
			"A":         "",
			"AT":        s.at,
			"BACK":      "/asp/Reports/StudentTotalMarks.asp",
			"ISTF":      "0",
			"LoginType": "0",
			"NA":        "",
			"PCLID":     "10171",
			"PP":        "/asp/Reports/StudentTotalMarks.asp",
			"RP":        "0",
			"RPTID":     "0",
			"RT":        "0",
			"SID":       "11207",
			"TA":        "",
			"ThmID":     "1",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Referer": p + s.Serv.Link + "/asp/Reports/StudentTotalMarks.asp",
		},
	}
	response1, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/StudentTotalMarks.asp", requestOption1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()

	fmt.Println(string(response1.Bytes()))

	return nil, nil
}
