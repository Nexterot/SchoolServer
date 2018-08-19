// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

/*
Package check содержит в себе проверку ответа на ошибки.
*/
package check

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/pkg/errors"
	"golang.org/x/net/html"

	gr "github.com/levigross/grequests"
)

// CheckResponse проверяет, не выкинуло ли нас с сервера и не пришло ли предупреждение.
func CheckResponse(s *dt.Session, response *gr.Response) (bool, error) {
	str := response.String()
	// Проверяем на Warning.
	if (response.StatusCode == 200) && strings.Contains(str, `ACTION="/asp/SecurityWarning.asp"`) {
		return false, retry(s, str)
	}
	// Проверяем на "безопасный выброс с сервера".
	if (response.StatusCode == 200) && strings.Contains(str, "<!--Jumptologin page-->") {
		return false, login(s)
	}

	// Проверка на "выброс с сервера".
	if (response.StatusCode != 200) || strings.Contains(strings.ToLower(str), "<title></title>") {
		return false, fmt.Errorf("You was logged out from server")
	}
	return true, nil
}

// login является копией одноименной экспортируемой функции Login. Это необходимо, чтобы
// победить циклические зависимости.
func login(s *dt.Session) error {
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

func retry(s *dt.Session, body string) error {
	p := "http://"

	// Подготовка переменных.
	// Поиск WarnType.
	wStr := `NAME="WarnType" VALUE="`
	ind := strings.Index(body, wStr)
	if ind == -1 {
		return fmt.Errorf("Can't find \"WarnType\" in \n%v", body)
	}
	warnType := body[ind+len(wStr):]
	ind = strings.Index(warnType, `"`)
	if ind == -1 {
		return fmt.Errorf("Can't find ending \"WarnType\" \" in \n%v", body)
	}
	warnType = warnType[:ind]
	// Поиск ATLIST.
	atlStr := `NAME="ATLIST" VALUE="`
	ind = strings.Index(body, atlStr)
	if ind == -1 {
		return fmt.Errorf("Can't find \"ATLIST\" in \n%v", body)
	}
	ATLIST := body[ind+len(atlStr):]
	ind = strings.Index(ATLIST, `"`)
	if ind == -1 {
		return fmt.Errorf("Can't find ending \"ATLIST\" \" in \n%v", body)
	}
	ATLIST = ATLIST[:ind]

	var requestOptions0 *gr.RequestOptions
	if warnType == "1" {
		requestOptions0 = &gr.RequestOptions{
			Data: map[string]string{
				"ATLIST":   ATLIST,
				"WarnType": warnType,
				"at":       s.AT,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   "kek",
			},
		}
	} else if warnType == "2" {
		requestOptions0 = &gr.RequestOptions{
			Data: map[string]string{
				"AT":       s.AT,
				"ATLIST":   ATLIST,
				"VER":      s.VER,
				"WarnType": warnType,
			},
			Headers: map[string]string{
				"Upgrade-Insecure-Requests": "1",
				"Origin":                    p + s.Serv.Link,
			},
		}
	} else {
		return fmt.Errorf("Unknown \"WarnType\" \"%v\" in \n%v", warnType, body)
	}
	_, err := s.Sess.Post(p+s.Serv.Link+"/asp/SecurityWarning.asp", requestOptions0)
	return err
}
