// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

/*
Package type01 - данный файл содержит в себе функции для обработки 1 типа сайтов.
*/
package type01

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"

	dt "SchoolServer/libtelco/sessions/data-types"
	ss "SchoolServer/libtelco/sessions/session"

	gr "github.com/levigross/grequests"
	"golang.org/x/net/html"
)

// Login логинится к серверу первого типа и создает очередную сессию.
func Login(s *ss.Session) error {
	// Создание сессии.
	p := "http://"

	// Полчение формы авторизации.
	// 0-ой Get-запрос.
	_, err := s.Sess.Get(p+s.Serv.Link+"/asp/ajax/getloginviewdata.asp", nil)
	if err != nil {
		return err
	}

	// 1-ый Post-запрос.
	response1, err := s.Sess.Post(p+s.Serv.Link+"/webapi/auth/getdata", nil)
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
	response2, err := s.Sess.Post(p+s.Serv.Link+"/asp/postlogin.asp", requestOption)
	if err != nil {
		return err
	}
	defer func() {
		_ = response2.Close()
	}()

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней "AT" и "VER".
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
	s.AT = f(parsedHTML, "AT")
	s.VER = f(parsedHTML, "VER")
	if (s.AT == "") || (s.VER == "") {
		return fmt.Errorf("Problems on school server: %s", s.Serv.Link)
	}
	return nil
}

// GetChildrenMap получает мапу детей в их UID с сервера первого типа.
func GetChildrenMap(s *ss.Session) error {
	p := "http://"

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
		Data: map[string]string{
			"AT":        s.AT,
			"LoginType": "0",
			"RPTID":     "0",
			"ThmID":     "1",
			"VER":       s.VER,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
		},
	}
	response0, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentTotalMarks.asp", requestOptions0)
	fmt.Println(string(response0.Bytes()))
	if err != nil {
		return err
	}
	defer func() {
		_ = response0.Close()
	}()
	if err := checkResponse(s, response0); err != nil {
		return err
	}
	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней мапу детей в их ID.
	parsedHTML, err := html.Parse(bytes.NewReader(response0.Bytes()))
	if err != nil {
		return err
	}

	var getChildrenIDNode func(*html.Node) *html.Node
	getChildrenIDNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "select" {
				for _, a := range node.Attr {
					if a.Key == "name" && a.Val == "SID" {
						return node
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			n := getChildrenIDNode(c)
			if n != nil {
				return n
			}
		}

		return nil
	}

	getChildrenIDs := func(node *html.Node) (map[string]string, error) {
		childrenIDs := make(map[string]string)
		idNode := getChildrenIDNode(node)
		if idNode != nil {
			for n := idNode.FirstChild; n != nil; n = n.NextSibling {
				if len(n.Attr) != 0 {
					for _, a := range n.Attr {
						if a.Key == "value" {
							childrenIDs[n.FirstChild.Data] = a.Val
							if _, err := strconv.Atoi(a.Val); err != nil {
								return childrenIDs, nil
							}
						}
					}
				}
			}
		} else {
			return childrenIDs, fmt.Errorf("Couldn't find children IDs Node")
		}

		return childrenIDs, err
	}

	s.ChildrenIDS, err = getChildrenIDs(parsedHTML)
	if err.Error() == "Couldn't find children IDs Node" {
		s.Type = ss.Student
		err = nil
	} else {
		s.Type = ss.Parent
	}
	return err
}

// TODO: протестить для пездюков.

// GetLessonsMap возвращает мапу предметов в их ID с сервера первого типа.
func GetLessonsMap(s *ss.Session, studentID string) (*dt.LessonsMap, error) {
	p := "http://"

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
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
	response0, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentGrades.asp", requestOptions0)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response0.Close()
	}()
	if err := checkResponse(s, response0); err != nil {
		return nil, err
	}

	// 1-ый Post-запрос.
	requestOptions1 := &gr.RequestOptions{
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
	response1, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentGrades.asp", requestOptions1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()

	if err := checkResponse(s, response1); err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней мапу предметов в их ID.
	parsedHTML, err := html.Parse(bytes.NewReader(response1.Bytes()))
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
						if a.Key == "valлue" {
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
