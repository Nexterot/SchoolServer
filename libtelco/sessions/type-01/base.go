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
	"strings"

	"github.com/pkg/errors"

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
		return errors.Wrap(err, "0 GET")
	}

	// 1-ый Post-запрос.
	response1, err := s.Sess.Post(p+s.Serv.Link+"/webapi/auth/getdata", nil)
	if err != nil {
		return errors.Wrap(err, "1 POST")
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
		return errors.Wrap(err, "decoding JSON")
	}

	// 2-ой Post-запрос.
	pw := s.Serv.Password
	hasher := md5.New()
	if _, err = hasher.Write([]byte(fa.Salt + pw)); err != nil {
		return errors.Wrap(err, "md5")
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
		return errors.Wrap(err, "2 POST")
	}
	defer func() {
		_ = response2.Close()
	}()

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней "AT" и "VER".
	parsedHTML, err := html.Parse(bytes.NewReader(response2.Bytes()))
	if err != nil {
		return errors.Wrap(err, "parsing HTML")
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
		return fmt.Errorf("problems on school server: %s", p+s.Serv.Link)
	}
	return nil
}

// Logout выходит с сервера первого типа.
func Logout(s *ss.Session) error {
	p := "http://"

	// 0-ой Post-запрос.
	r0 := func() (bool, error) {
		requestOptions := &gr.RequestOptions{
			Data: map[string]string{
				"AT":  s.AT,
				"VER": s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
			},
		}
		response, err := s.Sess.Post(p+s.Serv.Link+"/asp/logout.asp", requestOptions)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = response.Close()
		}()
		return checkResponse(s, response)
	}
	flag, err := r0()
	if err != nil {
		return errors.Wrap(err, "0 POST")
	}
	if !flag {
		flag, err = r0()
		if err != nil {
			return errors.Wrap(err, "retrying 0 POST")
		}
		if !flag {
			return fmt.Errorf("retry didn't work for 0 POST")
		}
	}
	return nil
}

// GetChildrenMap получает мапу детей в их UID с сервера первого типа.
func GetChildrenMap(s *ss.Session) error {
	p := "http://"

	// 0-ой Post-запрос.
	r0 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
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
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentTotalMarks.asp", ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := checkResponse(s, r)
		if err != nil {
			return nil, false, err
		}
		return r.Bytes(), flag, nil
	}
	b, flag, err := r0()
	if err != nil {
		return errors.Wrap(err, "0 POST")
	}
	if !flag {
		b, flag, err = r0()
		if err != nil {
			return errors.Wrap(err, "retrying 0 POST")
		}
		if !flag {
			return fmt.Errorf("retry didn't work for 0 POST")
		}
	}
	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней мапу детей в их ID.
	parsedHTML, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "parsing HTML")
	}

	var getChildrenIDNode func(*html.Node) *html.Node
	getChildrenIDNode = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			if node.Data == "select" || node.Data == "input" {
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

	// Находит ID учеников и также определяет тип сессии
	getChildrenIDs := func(node *html.Node) (map[string]ss.Student, bool, error) {
		// Находим ID учеников/ученика.
		childrenIDs := make(map[string]ss.Student)
		idNode := getChildrenIDNode(node)
		if idNode != nil {
			if idNode.FirstChild == nil {
				for _, a := range idNode.Attr {
					if a.Key == "value" {
						if idNode.PrevSibling != nil {
							nameNode := idNode.PrevSibling
							flag := true
							for nameNode != nil && flag {
								for _, b := range nameNode.Attr {
									if b.Key == "type" && b.Val == "text" {
										flag = false
										break
									}
								}
								if !flag {
									break
								}
								nameNode = nameNode.PrevSibling
							}
							if nameNode != nil && !flag {
								for _, a2 := range nameNode.Attr {
									if a2.Key == "value" {
										childrenIDs[a2.Val] = ss.Student{a.Val, ""}
										if _, err := strconv.Atoi(a.Val); err != nil {
											return nil, false, fmt.Errorf("ID has incorrect format \"%v\"", a.Val)
										}
									}
								}
							}
						}
					}
				}
			} else {
				for n := idNode.FirstChild; n != nil; n = n.NextSibling {
					if len(n.Attr) != 0 {
						for _, a := range n.Attr {
							if a.Key == "value" {
								childrenIDs[n.FirstChild.Data] = ss.Student{a.Val, ""}
								if _, err := strconv.Atoi(a.Val); err != nil {
									return nil, false, fmt.Errorf("ID has incorrect format \"%v\"", a.Val)
								}
							}
						}
					}
				}
			}
		} else {
			return nil, false, fmt.Errorf("Couldn't find children IDs Node")
		}

		// Находим тип сессии.
		sessTypeNode := idNode.Parent
		if sessTypeNode != nil && sessTypeNode.Data == "select" {
			sessTypeNode = sessTypeNode.Parent
		}
		isParent := false
		for sessTypeNode != nil && sessTypeNode.Data != "label" {
			sessTypeNode = sessTypeNode.PrevSibling
		}
		if sessTypeNode != nil && sessTypeNode.FirstChild != nil {
			if sessTypeNode.FirstChild.Data == "Ученики" {
				isParent = true
			}
		} else {
			return nil, false, errors.New("Couldn't find type of session")
		}

		return childrenIDs, isParent, nil
	}

	var isParent bool
	s.Children, isParent, err = getChildrenIDs(parsedHTML)
	if err != nil {
		return errors.Wrap(err, "parsing")
	}
	if isParent {
		s.Type = ss.Parent
	} else {
		s.Type = ss.Child
	}

	// 1-ый Post-запрос.
	r1 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"LoginType": "0",
				"RPTID":     "3",
				"ThmID":     "2",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/JournalAccess.asp", ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r.Close()
		}()
		return checkResponse(s, r)
	}

	type Filter struct {
		ID    string `json:"filterId"`
		Value string `json:"filterValue"`
	}

	type SelectedData struct {
		SelectedData []Filter `json:"selectedData"`
	}

	// 2-ой Post-запрос.
	r2 := func(json SelectedData) ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			JSON: json,
			Headers: map[string]string{
				"Origin":           p + s.Serv.Link,
				"X-Requested-With": "XMLHttpRequest",
				"at":               s.AT,
				"Referer":          p + s.Serv.Link + "/asp/Reports/ReportJournalAccess.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/webapi/reports/journal_access/initfilters", ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := checkResponse(s, r)
		return r.Bytes(), flag, err
	}

	// Если мы дошли до этого места, то можно начать искать CLID каждого ребенка.
	for k, v := range s.Children {
		flag, err := r1()
		if err != nil {
			return errors.Wrap(err, "1 POST")
		}
		if !flag {
			flag, err = r1()
			if err != nil {
				return errors.Wrap(err, "retrying 1 POST")
			}
			if !flag {
				return fmt.Errorf("retry didn't work for 1 POST")
			}
		}
		json := SelectedData{
			SelectedData: []Filter{Filter{"SID", v.SID}},
		}
		b, flag, err := r2(json)
		if err != nil {
			return err
		}
		if !flag {
			b, flag, err = r2(json)
			if err != nil {
				return err
			}
			if !flag {
				return fmt.Errorf("Retry didn't work")
			}
		}
		CLID := string(b)
		index := strings.Index(CLID, "\"value\":\"")
		if index == -1 {
			return fmt.Errorf("Invalid SID")
		}
		CLID = CLID[index+len("\"value\":\""):]
		index = strings.Index(CLID, "\"")
		if index == -1 {
			return fmt.Errorf("Invalid SID")
		}
		CLID = CLID[:index]
		v.CLID = CLID
		s.Children[k] = v
	}

	if len(s.Children) == 1 {
		for _, v := range s.Children {
			s.Child = v
		}
	}
	return nil
}

// GetLessonsMap возвращает мапу предметов в их ID с сервера первого типа.
func GetLessonsMap(s *ss.Session, studentID string) (*dt.LessonsMap, error) {
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
		return checkResponse(s, r)
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

		flag, err := checkResponse(s, r)
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
