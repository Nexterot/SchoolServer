package inner

import (
	dt "SchoolServer/libtelco/sessions/data-types"
	"errors"
	"io"
	"strconv"

	"golang.org/x/net/html"
)

// AverageMarkDynReportParser возвращает динамику среднего балла ученика.
// находится в inner-funcs, так как отчеты на всех серверах одинаковые.
func AverageMarkDynReportParser(r io.Reader) (*dt.AverageMarkDynReport, error) {
	parsedHTML, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	// Находит нод с табличкой
	var findAverageMarkDynTableNode func(*html.Node) *html.Node
	findAverageMarkDynTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			for _, a := range node.Attr {
				if (a.Key == "class") && (a.Val == "table-print-num") {
					return node
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findAverageMarkDynTableNode(c)
			if n != nil {
				return n
			}
		}
		return nil
	}

	// Формирует отчёт
	formAverageMarkDynReportTable := func(node *html.Node, data *[]dt.AverageMarkDynReportNote) error {
		if node != nil {
			stage := node.FirstChild.FirstChild
			var studentWorksAmount, studentAverageMark, classWorksAmount, classAverageMark *html.Node
			hasWorks := false

			// Проверяем вид отчёта
			if stage.NextSibling.NextSibling.NextSibling != nil {
				hasWorks = true
			}
			if hasWorks {
				studentWorksAmount = stage.NextSibling
				studentAverageMark = studentWorksAmount.NextSibling
				classWorksAmount = studentAverageMark.NextSibling
				classAverageMark = classWorksAmount.NextSibling

				studentWorksAmount = studentWorksAmount.FirstChild.NextSibling
				studentAverageMark = studentAverageMark.FirstChild.NextSibling
				classWorksAmount = classWorksAmount.FirstChild.NextSibling
				classAverageMark = classAverageMark.FirstChild.NextSibling
			} else {
				studentAverageMark = stage.NextSibling
				classAverageMark = studentAverageMark.NextSibling

				studentAverageMark = studentAverageMark.FirstChild.NextSibling
				classAverageMark = classAverageMark.FirstChild.NextSibling
			}
			stage = stage.FirstChild.NextSibling
			for stage != nil {
				var note dt.AverageMarkDynReportNote
				note.Date = stage.FirstChild.Data
				note.StudentAverageMark = studentAverageMark.FirstChild.Data
				note.ClassAverageMark = classAverageMark.FirstChild.Data

				if hasWorks {
					note.StudentWorksAmount, err = strconv.Atoi(studentWorksAmount.FirstChild.Data)
					if err != nil {
						return err
					}
					note.ClassWorksAmount, err = strconv.Atoi(classWorksAmount.FirstChild.Data)
					if err != nil {
						return err
					}
					studentWorksAmount = studentWorksAmount.NextSibling
					classWorksAmount = classWorksAmount.NextSibling
				}
				(*data) = append((*data), note)
				stage = stage.NextSibling
				studentAverageMark = studentAverageMark.NextSibling
				classAverageMark = classAverageMark.NextSibling
			}

			return nil
		}
		return errors.New("Node is nil in func formAverageMarkDynReportTable")
	}

	// Создаёт отчёт
	makeAverageMarkDynReportTable := func(node *html.Node) (*dt.AverageMarkDynReport, error) {
		var report dt.AverageMarkDynReport
		tableNode := findAverageMarkDynTableNode(node)
		report.Data = make([]dt.AverageMarkDynReportNote, 0, 10)
		err := formAverageMarkDynReportTable(tableNode, &report.Data)
		return &report, err
	}
	return makeAverageMarkDynReportTable(parsedHTML)
}
