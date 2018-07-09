// Copyright (C) 2018 Mikhail Masyagin

/*
Package parser - данный файл содержит в себе функции для обработки 1 типа сайтов.
*/
package parser

import (
	cp "SchoolServer/libtelco/config-parser"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	gr "github.com/levigross/grequests"
	"golang.org/x/net/html"
)

// login логинится к серверу и создает очередную сессию.
func (s *session) login() error {
	var err error
	switch s.serv.Type {
	case cp.FirstType:
		err = s.firstTypeLogin()
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.serv.Type)
	}
	return err
}

// firstTypeLogin логинится к серверу первого типа и создает очередную сессию.
func (s *session) firstTypeLogin() error {
	// Создание сессии.
	s.sess = gr.NewSession(nil)
	p := "http://"

	// Полчение формы авторизации.
	// 0-ой Get-запрос.
	response0, err := s.sess.Get(p+s.serv.Link+"/asp/ajax/getloginviewdata.asp", nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = response0.Close()
	}()

	// 1-ый Post-запрос.
	response1, err := s.sess.Post(p+s.serv.Link+"/webapi/auth/getdata", nil)
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
	pw := s.serv.Password
	hasher := md5.New()
	if _, err = hasher.Write([]byte(pw)); err != nil {
		return err
	}
	pw = hex.EncodeToString(hasher.Sum(nil))
	hasher = md5.New()
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
			"PW":        pw[:len(s.serv.Password)],
			"SCID":      "2",
			"SFT":       "2",
			"SID":       "77",
			"UN":        s.serv.Login,
			"lt":        fa.Lt,
			"pw2":       pw,
			"ver":       fa.Ver,
		},
		Headers: map[string]string{
			"Referer": p + s.serv.Link + "/",
		},
	}
	response2, err := s.sess.Post(p+s.serv.Link+"/asp/postlogin.asp", requestOption)
	if err != nil {
		return err
	}
	defer func() {
		_ = response2.Close()
	}()

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа и найти в ней "AT".
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
	s.at = f(parsedHTML, "AT")
	s.ver = f(parsedHTML, "VER")
	if (s.at == "") || (s.ver == "") {
		return fmt.Errorf("Problems on school server: %s", s.serv.Link)
	}

	return nil
}

func (s *session) getDayTimeTableFirst(date string) (*DayTimeTable, error) {
	p := "http://"
	var dayTimeTable *DayTimeTable

	// 0-ой Get-запрос.
	RequestOption0 := &gr.RequestOptions{
		Headers: map[string]string{
			"at":      s.at,
			"Referer": p + s.serv.Link + "/asp/Calendar/DayViewS.asp",
		},
	}
	_, err := s.sess.Get(p+s.serv.Link+"/asp/ajax/GetCalendar.asp?AT="+s.at+"&startDate=01.09.2017&endDate=31.08.2018", RequestOption0)
	if err != nil {
		return dayTimeTable, err
	}

	// 1-ый Post-запрос.
	requestOption1 := &gr.RequestOptions{
		Data: map[string]string{
			"AT":        s.at,
			"BackPage":  "/asp/Calendar/DayViewS.asp",
			"DATE":      date,
			"EventID":   "",
			"EventType": "",
			"FOO":       "",
			"LoginType": "0",
			"PCLID_UP":  "10169_0",
			"SID":       "11198",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Referer": p + s.serv.Link + "/asp/Calendar/DayViewS.asp",
		},
	}
	response, err := s.sess.Post(p+s.serv.Link+"/asp/Calendar/DayViewS.asp", requestOption1)
	if err != nil {
		return dayTimeTable, err
	}
	defer func() {
		_ = response.Close()
	}()

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа и найти в ней расписание на текущий день.

	// Андрей, твоя задача по образу и подобию логина сделать вложенную функцию парсера и записать результат ее работы в dayTimeTable.

	return dayTimeTable, nil
}
