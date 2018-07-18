package inner

import (
	dt "SchoolServer/libtelco/sessions/data-types"
	"errors"
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// ParentInfoLetterReportParser возвращает отчет "Информационное письмо для родителей".
// находится в inner-funcs, так как отчеты на всех серверах одинаковые.
func ParentInfoLetterReportParser(r io.Reader) (*dt.ParentInfoLetterReport, error) {
	parsedHTML, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	// Находит нод с табличкой
	var findParentInfoLetterTableNode func(*html.Node) *html.Node
	findParentInfoLetterTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "table" {
				for _, a := range node.Attr {
					if a.Key == "class" && a.Val == "table-print" {
						return node
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findParentInfoLetterTableNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	// Формирует отчёт
	var formParentInfoLetterReportTable func(*html.Node) ([]dt.ParentInfoLetterReportNote, error)
	formParentInfoLetterReportTable = func(node *html.Node) ([]dt.ParentInfoLetterReportNote, error) {
		notes := make([]dt.ParentInfoLetterReportNote, 0, 10)
		if node != nil {
			// Определяем вид отчёта
			hasPeriodMark := false
			n := node.FirstChild.FirstChild
			k := 0
			for n = n.FirstChild; n != nil; n = n.NextSibling {
				k++
			}
			if k == 3 {
				hasPeriodMark = true
			}

			noteNode := node.FirstChild.FirstChild.NextSibling
			for noteNode = noteNode.NextSibling; noteNode != nil && len(noteNode.Attr) == 0; noteNode = noteNode.NextSibling {
				// Добавляем запись
				note := *new(dt.ParentInfoLetterReportNote)
				c := noteNode.FirstChild

				note.Name = c.FirstChild.Data

				note.Marks = make([]int, 8, 8)
				for i := 0; i < 8; i++ {
					c = c.NextSibling
					if len(c.FirstChild.Data) == 2 && c.FirstChild.Data[0] == 194 {
						note.Marks[i] = -1
					} else {
						note.Marks[i], err = strconv.Atoi(c.FirstChild.Data)
						if err != nil {
							return notes, err
						}
					}
				}

				c = c.NextSibling
				v, err := strconv.ParseFloat(strings.Replace(c.FirstChild.Data, ",", ".", 1), 32)
				if err != nil {
					note.AverageMark = -1.0
				} else {
					note.AverageMark = float32(v)
				}

				if hasPeriodMark {
					c = c.NextSibling
					note.MarkForPeriod, err = strconv.Atoi(c.FirstChild.Data)
					if err != nil {
						note.MarkForPeriod = -1
					}
				}

				notes = append(notes, note)
			}

			if len(noteNode.Attr) == 1 && noteNode.Attr[0].Val == "totals" {
				note := *new(dt.ParentInfoLetterReportNote)
				c := noteNode.FirstChild

				note.Name = c.FirstChild.Data

				note.Marks = make([]int, 8, 8)
				for i := 0; i < 8; i++ {
					c = c.NextSibling
					if len(c.FirstChild.Data) == 2 && c.FirstChild.Data[0] == 194 {
						note.Marks[i] = -1
					} else {
						note.Marks[i], err = strconv.Atoi(c.FirstChild.Data)
						if err != nil {
							return notes, err
						}
					}
				}

				c = c.NextSibling
				v, err := strconv.ParseFloat(strings.Replace(c.FirstChild.Data, ",", ".", 1), 32)
				if err != nil {
					note.AverageMark = -1.0
				} else {
					note.AverageMark = float32(v)
				}

				if hasPeriodMark {
					c = c.NextSibling
					note.MarkForPeriod, err = strconv.Atoi(c.FirstChild.Data)
					if err != nil {
						note.MarkForPeriod = -1
					}
				} else {
					note.MarkForPeriod = -1
				}

				notes = append(notes, note)
			}
		} else {
			return notes, errors.New("Node is nil in func formParentInfoLetterReportTable")
		}

		return notes, nil
	}

	// Создаёт отчёт
	makeParentInfoLetterReport := func(node *html.Node) (*dt.ParentInfoLetterReport, error) {
		var report dt.ParentInfoLetterReport
		tableNode := findParentInfoLetterTableNode(node)
		report.Data, err = formParentInfoLetterReportTable(tableNode)

		return &report, err
	}

	var report *dt.ParentInfoLetterReport
	report, err = makeParentInfoLetterReport(parsedHTML)
	if err != nil {
		return nil, err
	}

	return report, nil
}
