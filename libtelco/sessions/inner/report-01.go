package inner

import (
	dt "SchoolServer/libtelco/sessions/data-types"
	"io"
	"strconv"

	"golang.org/x/net/html"
)

// TotalMarkReportParser возвращает успеваемость ученика.
// находится в inner-funcs, так как отчеты на всех серверах одинаковые.
func TotalMarkReportParser(r io.Reader) (*dt.TotalMarkReport, error) {
	parsedHTML, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	// Находит нод с табличкой
	var findNode func(*html.Node) *html.Node
	findNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			for _, a := range node.Attr {
				if (a.Key == "class") && (a.Val == "table-print-num") {
					return node
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findNode(c)
			if n != nil {
				return n
			}
		}
		return nil
	}

	// Формирует отчёт
	var formTotalMarkReport func(*html.Node, map[string][]int)
	formTotalMarkReport = func(node *html.Node, report map[string][]int) {
		for _, a := range node.Attr {
			if (a.Key == "class") && (a.Val == "cell-text") {
				// Нашли урок
				report[node.FirstChild.Data] = make([]int, 7)
				i := 0
				flag := false
				for c := node.NextSibling; c != nil; c = c.NextSibling {
					if c.FirstChild != nil {
						mark, err := strconv.Atoi(c.FirstChild.Data)
						if flag {
							i++
						}
						if (err != nil) || (i > 6) {
							continue
						}
						flag = true
						report[node.FirstChild.Data][i] = mark
					}
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			formTotalMarkReport(c, report)
		}
	}

	// Создаёт отчёт
	makeTotalMarkReport := func(node *html.Node) (*dt.TotalMarkReport, error) {
		var report dt.TotalMarkReport
		tableNode := findNode(node)
		data := make(map[string][]int)
		formTotalMarkReport(tableNode, data)
		for k, v := range data {
			innerReport := dt.TotalMarkReportNote{k, v[0], v[1], v[2], v[3], v[4], v[5], v[6]}
			report.Table = append(report.Table, innerReport)
		}
		return &report, nil
	}

	return makeTotalMarkReport(parsedHTML)
}
