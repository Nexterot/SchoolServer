// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

/*
Package type01 - данный файл содержит в себе проверку ответа на ошибки.
*/
package type01

import (
	ss "SchoolServer/libtelco/sessions/session"
	"fmt"
	"strings"

	gr "github.com/levigross/grequests"
)

// checkResponse проверяет, не выкинуло ли нас с сервера и не пришло ли предупреждение.
func checkResponse(s *ss.Session, response *gr.Response) (bool, error) {
	str := response.String()
	// Проверяем на Warning.
	if (response.StatusCode == 200) && strings.Contains(str, `ACTION="/asp/SecurityWarning.asp"`) {
		if err := retry(s, str); err != nil {
			return false, err
		}
		return false, nil
	}

	// Проверка на "выброс с сервера".
	if (response.StatusCode != 200) || strings.Contains(strings.ToLower(str), "<title></title>") {
		return false, fmt.Errorf("You was logged out from server")
	}
	return true, nil
}

func retry(s *ss.Session, body string) error {
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
	response0, err := s.Sess.Post(p+s.Serv.Link+"/asp/SecurityWarning.asp", requestOptions0)
	if err != nil {
		return err
	}
	return nil
}
