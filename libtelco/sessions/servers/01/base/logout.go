// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package base

import (
	"fmt"

	gr "github.com/levigross/grequests"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"
	"github.com/pkg/errors"
)

// Logout выходит с сервера первого типа.
func Logout(s *dt.Session) error {
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
		return check.CheckResponse(s, response)
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
