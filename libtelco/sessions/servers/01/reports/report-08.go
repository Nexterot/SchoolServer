// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package reports

import (
	"bytes"
	"fmt"
	"strconv"

	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/reportsparser"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"
	"golang.org/x/net/html"

	"github.com/pkg/errors"

	gr "github.com/levigross/grequests"
)

/*
08 тип.
*/

// GetParentInfoLetterData возвращает параметры отчета восьмого типа с сервера первого типа.
func GetParentInfoLetterData(s *dt.Session) (*dt.ParentInfoLetterData, error) {
	p := "http://"

	// 0-ой Post-запрос.
	r0 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"LoginType": "0",
				"RPTID":     "4",
				"ThmID":     "2",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportParentInfoLetter.asp", ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := check.CheckResponse(s, r)
		return r.Bytes(), flag, err
	}
	b, flag, err := r0()
	if err != nil {
		return nil, errors.Wrap(err, "0 POST")
	}
	if !flag {
		b, flag, err = r0()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 0 POST")
		}
		if !flag {
			return nil, fmt.Errorf("retry didn't work for 0 POST")
		}
	}
	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа и найти в ней параметры отчета восьмого типа.
	parsedHTML, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	// Находит node с табличкой
	var findParentInfoLetterDataTableNode func(*html.Node) *html.Node
	findParentInfoLetterDataTableNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "div" {
				for _, a := range node.Attr {
					if a.Key == "class" && a.Val == "filters-panel col-md-12" {
						return node
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := findParentInfoLetterDataTableNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	// Формирует данные
	formParentInfoLetterData := func(node *html.Node) (*dt.ParentInfoLetterData, error) {
		data := &dt.ParentInfoLetterData{}
		if node == nil {
			return nil, errors.New("Node is nil")
		}
		if node.FirstChild == nil {
			return nil, errors.New("Couldn't find parent info letter data")
		}
		infoNode := node.FirstChild
		for infoNode != nil && infoNode.Data != "div" {
			infoNode = infoNode.NextSibling
		}
		if infoNode == nil {
			return nil, errors.New("Couldn't find parent info letter data")
		}
		infoNode = infoNode.NextSibling
		for infoNode != nil && infoNode.Data != "div" {
			infoNode = infoNode.NextSibling
		}
		if infoNode == nil {
			return nil, errors.New("Couldn't find parent info letter data")
		}
		infoNode = infoNode.NextSibling
		for infoNode != nil && infoNode.Data != "div" {
			infoNode = infoNode.NextSibling
		}
		if infoNode == nil {
			return nil, errors.New("Couldn't find parent info letter data")
		}
		if infoNode.FirstChild != nil {
			// Ищем виды отчётов
			dataNode := infoNode.FirstChild
			for dataNode != nil && dataNode.Data != "div" {
				dataNode = dataNode.NextSibling
			}
			if dataNode != nil && dataNode.FirstChild != nil {
				dataNode = dataNode.FirstChild
				for dataNode != nil && dataNode.Data != "select" {
					dataNode = dataNode.NextSibling
				}
				if dataNode != nil && dataNode.FirstChild != nil {
					data.ReportTypes = make([]dt.ReportType, 0, 2)
					for dataNode = dataNode.FirstChild; dataNode != nil; dataNode = dataNode.NextSibling {
						if dataNode.FirstChild != nil {
							reportType := dt.ReportType{}
							reportType.ReportTypeName = dataNode.FirstChild.Data
							for _, a := range dataNode.Attr {
								if a.Key == "value" {
									reportType.ReportTypeID, err = strconv.Atoi(a.Val)
									if err != nil {
										return nil, err
									}
									break
								}
							}
							data.ReportTypes = append(data.ReportTypes, reportType)
						}
					}
				}
			}
		}
		infoNode = infoNode.NextSibling
		if infoNode != nil && infoNode.FirstChild != nil {
			// Ищем периоды
			dataNode := infoNode.FirstChild
			for dataNode != nil && dataNode.Data != "div" {
				dataNode = dataNode.NextSibling
			}
			if dataNode != nil && dataNode.FirstChild != nil {
				dataNode = dataNode.FirstChild
				for dataNode != nil && dataNode.Data != "select" {
					dataNode = dataNode.NextSibling
				}
				if dataNode != nil && dataNode.FirstChild != nil {
					data.Periods = make([]dt.Period, 0, 4)
					for dataNode = dataNode.FirstChild; dataNode != nil; dataNode = dataNode.NextSibling {
						if dataNode.FirstChild != nil {
							period := dt.Period{}
							period.PeriodName = dataNode.FirstChild.Data
							for _, a := range dataNode.Attr {
								if a.Key == "value" {
									period.PeriodID, err = strconv.Atoi(a.Val)
									if err != nil {
										return nil, err
									}
									break
								}
							}
							data.Periods = append(data.Periods, period)
						}
					}
				}
			}
		}

		return data, nil
	}

	// Создаёт данные для восьмого отчёта
	makeParentInfoLetterData := func(node *html.Node) (*dt.ParentInfoLetterData, error) {
		tableNode := findParentInfoLetterDataTableNode(node)
		return formParentInfoLetterData(tableNode)
	}

	return makeParentInfoLetterData(parsedHTML)
}

// GetParentInfoLetterReport возвращает шаблон письма родителям с сервера первого типа.
func GetParentInfoLetterReport(s *dt.Session, reportTypeID, periodID, studentID string) (*dt.ParentInfoLetterReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	r0 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"LoginType": "0",
				"RPTID":     "4",
				"ThmID":     "2",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportParentInfoLetter.asp", ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r.Close()
		}()
		return check.CheckResponse(s, r)
	}
	flag, err := r0()
	if err != nil {
		return nil, errors.Wrap(err, "0 POST")
	}
	if !flag {
		flag, err = r0()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 0 POST")
		}
		if !flag {
			return nil, fmt.Errorf("retry didn't work for 0 POST")
		}
	}

	// 1-ый Post-запрос.
	r1 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"A":          "",
				"AT":         s.AT,
				"BACK":       "/asp/Reports/ReportParentInfoLetter.asp",
				"LoginType":  "0",
				"NA":         "",
				"PCLID":      "",
				"PP":         "/asp/Reports/ReportParentInfoLetter.asp",
				"RP":         "",
				"RPTID":      "4",
				"RT":         "",
				"ReportType": reportTypeID,
				"SID":        studentID,
				"TA":         "",
				"TERMID":     periodID,
				"ThmID":      "2",
				"VER":        s.VER,
				"drWeek":     "",
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/ReportParentInfoLetter.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportParentInfoLetter.asp", ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r.Close()
		}()
		return check.CheckResponse(s, r)
	}
	flag, err = r1()
	if err != nil {
		return nil, errors.Wrap(err, "1 POST")
	}
	if !flag {
		flag, err = r1()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 1 POST")
		}
		if !flag {
			return nil, fmt.Errorf("retry didn't work for 1 POST")
		}
	}

	// 2-ой Post-запрос.
	r2 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"A":          "",
				"AT":         s.AT,
				"BACK":       "/asp/Reports/ReportParentInfoLetter.asp",
				"LoginType":  "0",
				"NA":         "",
				"PCLID":      "",
				"PP":         "/asp/Reports/ReportParentInfoLetter.asp",
				"RP":         "",
				"RPTID":      "4",
				"RT":         "",
				"ReportType": reportTypeID,
				"SID":        studentID,
				"TA":         "",
				"TERMID":     periodID,
				"ThmID":      "2",
				"VER":        s.VER,
				"drWeek":     "",
			},
			Headers: map[string]string{
				"Origin":           p + s.Serv.Link,
				"X-Requested-With": "XMLHttpRequest",
				"at":               s.AT,
				"Referer":          p + s.Serv.Link + "/asp/Reports/ReportParentInfoLetter.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ParentInfoLetter.asp", ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := check.CheckResponse(s, r)
		return r.Bytes(), flag, err
	}
	b, flag, err := r2()
	if err != nil {
		return nil, errors.Wrap(err, "2 POST")
	}
	if !flag {
		b, flag, err = r2()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 2 POST")
		}
		if !flag {
			return nil, fmt.Errorf("retry didn't work for 2 POST")
		}
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа и найти в ней шаблон письма родителям.
	return reportsparser.ParentInfoLetterReportParser(bytes.NewReader(b))
}
