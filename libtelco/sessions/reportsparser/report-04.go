package reportsparser

import (
	"io"
	"strconv"

	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"

	"golang.org/x/net/html"
)

// StudentGradeReportParser возвращает отчет об успеваемости ученика по предмету.
// находится в inner-funcs, так как отчеты на всех серверах одинаковые.
func StudentGradeReportParser(r io.Reader) (*dt.StudentGradeReport, error) {
	parsedHTML, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	// Находит нод с табличкой
	var findPerformanceTableNode func(*html.Node) *html.Node
	findPerformanceTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			for _, a := range node.Attr {
				if (a.Key == "class") && (a.Val == "table-print") {
					return node
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findPerformanceTableNode(c)
			if n != nil {
				return n
			}
		}
		return nil
	}

	// Формирует отчёт
	formStudentGradeReportTable := func(node *html.Node) ([]dt.StudentGradeReportNote, error) {
		notes := make([]dt.StudentGradeReportNote, 0, 10)
		if node != nil {
			noteNode := node.FirstChild.FirstChild
			for noteNode = noteNode.NextSibling; noteNode != nil && len(noteNode.Attr) == 0; noteNode = noteNode.NextSibling {
				// Добавляем запись
				perfNote := *new(dt.StudentGradeReportNote)
				c := noteNode.FirstChild
				if c.FirstChild != nil {
					perfNote.Type = c.FirstChild.Data
				}
				c = c.NextSibling
				if c.FirstChild != nil {
					perfNote.Theme = c.FirstChild.Data
				}
				c = c.NextSibling
				if c.FirstChild != nil {
					perfNote.DateOfCompletion = c.FirstChild.Data
				}
				c = c.NextSibling
				if c.FirstChild != nil {
					perfNote.Mark, err = strconv.Atoi(c.FirstChild.Data)
					if err != nil {
						perfNote.Mark = -1
					}
				}
				notes = append(notes, perfNote)
			}
			if noteNode != nil {
				// Добавляем запись о количестве заданий и средней оценке.
				perfNote := *new(dt.StudentGradeReportNote)
				perfNote.Type = ""
				c := noteNode.FirstChild.NextSibling
				if c.FirstChild.FirstChild.FirstChild != nil {
					perfNote.Theme = c.FirstChild.FirstChild.FirstChild.Data
				}
				c = c.NextSibling
				if c.FirstChild.FirstChild.FirstChild != nil {
					perfNote.DateOfCompletion = c.FirstChild.FirstChild.FirstChild.Data
				}
				perfNote.Mark = -1
				notes = append(notes, perfNote)
			}
		}
		return notes, nil
	}

	// Создаёт отчёт
	makeStudentGradeReportTable := func(node *html.Node) (*dt.StudentGradeReport, error) {
		report := dt.NewStudentGradeReport()
		tableNode := findPerformanceTableNode(node)
		report.Data, err = formStudentGradeReportTable(tableNode)
		return report, err
	}

	return makeStudentGradeReportTable(parsedHTML)
}
