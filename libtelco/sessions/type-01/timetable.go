package type01

import (
	dt "SchoolServer/libtelco/sessions/data-types"
	ss "SchoolServer/libtelco/sessions/session"
	"bytes"
	"strings"
	"unicode"

	gr "github.com/levigross/grequests"
	"golang.org/x/net/html"
)

// GetDayTimeTable возвращает расписание на один день c сервера первого типа.
func GetDayTimeTable(s *ss.Session, date string) (*dt.DayTimeTable, error) {
	p := "http://"
	var dayTimeTable *dt.DayTimeTable

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
		Data: map[string]string{
			"AT":        s.AT,
			"BackPage":  "/asp/Calendar/DayViewS.asp",
			"DATE":      date,
			"EventID":   "",
			"EventType": "",
			"FOO":       "",
			"LoginType": "0",
			"PCLID_UP":  "",
			"SID":       "11207",
			"VER":       s.VER,
		},
		Headers: map[string]string{
			"Referer": p + s.Serv.Link + "/asp/Calendar/DayViewS.asp",
		},
	}
	response0, err := s.Sess.Post(p+s.Serv.Link+"/asp/Calendar/DayViewS.asp", requestOptions0)
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
	// находящуюся в теле ответа, и найти в ней расписание на текущий день.
	parsedHTML, err := html.Parse(bytes.NewReader(response0.Bytes()))
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
	getLessons := func(node *html.Node) []dt.Lesson {
		// Максимальное количество уроков.
		n := 10

		lessonsNode := searchForLessonsNode(node)

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
			// Случай, когда в расписании указаны каникулы
			infoNode := lessonsNode.FirstChild.NextSibling.FirstChild.NextSibling.FirstChild
			date := infoNode.FirstChild.Data
			infoNode = infoNode.NextSibling.FirstChild
			event := infoNode.Data + infoNode.NextSibling.FirstChild.Data
			s := strings.Split(date, "-")
			lessons[0].Begin = s[0][:len(s[0])-2]
			lessons[0].End = s[1][2:]
			lessons[0].Name = event
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