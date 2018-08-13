package inner

import (
	"errors"
	"io"
	"strconv"

	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/data-types"

	"golang.org/x/net/html"
)

// JournalAccessReportParser возвращает отчет о доступе к классному журналу.
// находится в inner-funcs, так как отчеты на всех серверах одинаковые.
func JournalAccessReportParser(r io.Reader) (*dt.JournalAccessReport, error) {
	parsedHTML, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	// Находит нод с табличкой
	var findJournalAccessTableNode func(*html.Node) *html.Node
	findJournalAccessTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "table" {
				for _, a := range node.Attr {
					if a.Key == "class" && a.Val == "table-print" {
						// Проверяем, что нашли нужную таблицу(их две на страницах отчёта)
						c := node.FirstChild.FirstChild
						i := 0
						for c = c.FirstChild; c != nil; c = c.NextSibling {
							i++
						}
						if i == 7 {
							return node
						}
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findJournalAccessTableNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	// Формирует отчёт
	var formJournalAccessReportTable func(*html.Node) ([]dt.JournalAccessReportNote, error)
	formJournalAccessReportTable = func(node *html.Node) ([]dt.JournalAccessReportNote, error) {
		notes := make([]dt.JournalAccessReportNote, 0, 10)
		if node != nil {
			noteNode := node.FirstChild.FirstChild
			for noteNode = noteNode.NextSibling; noteNode != nil; noteNode = noteNode.NextSibling {
				// Добавляем запись
				note := *new(dt.JournalAccessReportNote)
				c := noteNode.FirstChild
				if c.FirstChild != nil {
					note.Class, err = strconv.Atoi(c.FirstChild.Data)
					if err != nil {
						note.Class = -1
					}
				} else {
					note.Class = -1
				}
				c = c.NextSibling
				if c.FirstChild != nil {
					note.Subject = c.FirstChild.Data
				} else {
					note.Subject = ""
				}
				c = c.NextSibling
				if c.FirstChild != nil {
					note.Date = c.FirstChild.Data
				} else {
					note.Date = ""
				}
				c = c.NextSibling
				if c.FirstChild != nil {
					note.User = c.FirstChild.Data
				} else {
					note.User = ""
				}
				c = c.NextSibling
				if c.FirstChild != nil {
					note.LessonDate = c.FirstChild.Data
				} else {
					note.LessonDate = ""
				}
				c = c.NextSibling
				if c.FirstChild != nil {
					note.Period = c.FirstChild.Data
				} else {
					note.Period = ""
				}
				c = c.NextSibling
				if c.FirstChild != nil {
					note.Action = c.FirstChild.Data
				} else {
					note.Action = ""
				}

				notes = append(notes, note)
			}
		} else {
			return nil, errors.New("Node is nil in func formJournalAccessReportTable")
		}

		return notes, nil
	}

	// Создаёт отчёт
	makeJournalAccessReport := func(node *html.Node) (*dt.JournalAccessReport, error) {
		var report dt.JournalAccessReport
		tableNode := findJournalAccessTableNode(node)
		report.Data, err = formJournalAccessReportTable(tableNode)

		return &report, err
	}

	var report *dt.JournalAccessReport
	report, err = makeJournalAccessReport(parsedHTML)
	if err != nil {
		return nil, err
	}

	return report, nil
}
