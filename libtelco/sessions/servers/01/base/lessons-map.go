// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package base

import (
	"bytes"
	"fmt"
	"strconv"

	gr "github.com/levigross/grequests"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

// GetLessonsMap возвращает мапу предметов в их ID с сервера первого типа.
func GetLessonsMap(s *dt.Session, studentID string) (*dt.LessonsMap, error) {
	p := "http://"

	// 0-ой Post-запрос.
	r0 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"LoginType": "0",
				"RPTID":     "0",
				"ThmID":     "2",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentGrades.asp", ro)
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
			return nil, fmt.Errorf("Retry didn't work for 0 POST")
		}
	}

	// 1-ый Post-запрос.
	r1 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"A":         "",
				"ADT":       "01.09.2017",
				"AT":        s.AT,
				"BACK":      "/asp/Reports/ReportStudentGrades.asp",
				"DDT":       "31.08.2018",
				"LoginType": "0",
				"NA":        "",
				"PCLID_IUP": "",
				"PP":        "/asp/Reports/ReportStudentGrades.asp",
				"RP":        "",
				"RPTID":     "0",
				"RT":        "",
				"SCLID":     "",
				"SID":       studentID,
				"TA":        "",
				"ThmID":     "2",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/ReportStudentGrades.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentGrades.asp", ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()

		flag, err := check.CheckResponse(s, r)
		return r.Bytes(), flag, err
	}
	b, flag, err := r1()
	if err != nil {
		return nil, errors.Wrap(err, "1 POST")
	}
	if !flag {
		b, flag, err = r1()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 1 POST")
		}
		if !flag {
			return nil, fmt.Errorf("Retry didn't work for 1 POST")
		}
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней мапу предметов в их ID.
	parsedHTML, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var getSubjectsIDNode func(*html.Node) *html.Node
	getSubjectsIDNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "select" {
				for _, a := range node.Attr {
					if (a.Key == "name") && (a.Val == "SCLID") {
						return node
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := getSubjectsIDNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	getSubjectsIDs := func(node *html.Node) (map[string]string, error) {
		subjectsIDsMap := make(map[string]string)
		idNode := getSubjectsIDNode(node)
		if idNode != nil {
			for n := idNode.FirstChild; n != nil; n = n.NextSibling {
				if len(n.Attr) != 0 {
					for _, a := range n.Attr {
						if a.Key == "value" {
							subjectsIDsMap[n.FirstChild.Data] = a.Val
							if _, err = strconv.Atoi(a.Val); err != nil {
								return nil, err
							}
						}
					}
				}
			}
		} else {
			return nil, fmt.Errorf("Couldn't find SubjectsIDs Node")
		}
		return subjectsIDsMap, nil
	}

	subjectsIDsMap, err := getSubjectsIDs(parsedHTML)
	if err != nil {
		return nil, err
	}
	var lm dt.LessonsMap
	for k, v := range subjectsIDsMap {
		lm.Data = append(lm.Data, dt.LessonMap{k, v})
	}
	return &lm, nil
}
