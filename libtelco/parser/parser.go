// Copyright (C) 2018 Mikhail Masyagin

/*
Package parser содержит набор парсеров различных школьных сайтов и
"ООП-обертку" для одного расписания".
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

// TimeTable struct содержит в себе полное описание
// расписания одной школы в независимости от типа ее сервера.
type TimeTable struct {
}

// NewTimeTable создает новое расписание.
func NewTimeTable() *TimeTable {
	return &TimeTable{}
}

// ParseSchoolServer полностью парсит расписание с одного школьного сервера.
func (timeTable *TimeTable) ParseSchoolServer(schoolServer *cp.SchoolServer) error {
	switch schoolServer.Type {
	case cp.FirstType:
		return timeTable.firstTypeParser(schoolServer)
	default:
		err := fmt.Errorf("Unknown SchoolServer Type: %d", schoolServer.Type)
		return err
	}
}

// firstTypeParser парсит сервера первого типа.
func (timeTable *TimeTable) firstTypeParser(schoolServer *cp.SchoolServer) error {
	at, ver, err := timeTable.firstTypeLogin(schoolServer)
	fmt.Println("AT", at)
	fmt.Println("VER", ver)
	return err
}

// firstTypeLogin заходит на сервера первого типа.
func (timeTable *TimeTable) firstTypeLogin(schoolServer *cp.SchoolServer) (at, ver string, err error) {
	// Создание сессии.

	session := gr.NewSession(nil)
	p := "http://"
	// Полчение формы авторизации.
	// 0-ой Get-запрос.
	response0, err := session.Get(p+schoolServer.Link+"/asp/ajax/getloginviewdata.asp", nil)
	if err != nil {
		return
	}
	defer func() {
		_ = response0.Close()
	}()
	// 1-ый Post-запрос.
	response1, err := session.Post(p+schoolServer.Link+"/webapi/auth/getdata", nil)
	if err != nil {
		return
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
		return
	}
	// 2-ой Post-запрос.
	pw := schoolServer.Password
	hasher := md5.New()
	if _, err = hasher.Write([]byte(pw)); err != nil {
		return
	}
	pw = hex.EncodeToString(hasher.Sum(nil))
	hasher = md5.New()
	if _, err = hasher.Write([]byte(fa.Salt + pw)); err != nil {
		return
	}
	pw = hex.EncodeToString(hasher.Sum(nil))
	requestOption := &gr.RequestOptions{
		Data: map[string]string{
			"CID":       "2",
			"CN":        "1",
			"LoginType": "1",
			"PID":       "-1",
			"PW":        pw[:len(schoolServer.Password)],
			"SCID":      "2",
			"SFT":       "2",
			"SID":       "77",
			"UN":        schoolServer.Login,
			"lt":        fa.Lt,
			"pw2":       pw,
			"ver":       fa.Ver,
		},
		Headers: map[string]string{
			"Referer": p + schoolServer.Link + "/",
		},
	}
	response2, err := session.Post(p+schoolServer.Link+"/asp/postlogin.asp", requestOption)
	if err != nil {
		return
	}
	defer func() {
		_ = response2.Close()
	}()
	fmt.Println(string(response2.Bytes()))
	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа и найти в ней "AT".
	parsedHTML, err := html.Parse(bytes.NewReader(response2.Bytes()))
	if err != nil {
		return
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

	at = f(parsedHTML, "AT")
	ver = f(parsedHTML, "VER")
	return
}

// firstTypeLogout выходит с сервера первого типа.
// Использование данной функции вызвано необходимостью как-то избавиться от
// постоянных сообщений о нескольких одинаковых пользователях в системе, а также от
// ограничения на время работы - 45 минут.
func (timeTable *TimeTable) firstTypeLogout(schoolServer *cp.SchoolServer) error {
	return nil
}
