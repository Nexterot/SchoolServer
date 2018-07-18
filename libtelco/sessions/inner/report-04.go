package inner

import (
	dt "SchoolServer/libtelco/sessions/data-types"
	"io"
	"strconv"

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
				perfNote.Type = c.FirstChild.Data
				c = c.NextSibling
				perfNote.Theme = c.FirstChild.Data
				c = c.NextSibling
				perfNote.DateOfCompletion = c.FirstChild.Data
				c = c.NextSibling
				perfNote.Mark, err = strconv.Atoi(c.FirstChild.Data)
				if err != nil {
					return notes, err
				}
				notes = append(notes, perfNote)
			}
		}
		return notes, nil
	}

	// Создаёт отчёт
	makeStudentGradeReportTable := func(node *html.Node) (*dt.StudentGradeReport, error) {
		var report dt.StudentGradeReport
		tableNode := findPerformanceTableNode(node)
		report.Data, err = formStudentGradeReportTable(tableNode)
		return &report, err
	}

	return makeStudentGradeReportTable(parsedHTML)
}
