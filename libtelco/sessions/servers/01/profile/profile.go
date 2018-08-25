// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package profile

import (
	"fmt"

	gr "github.com/levigross/grequests"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"
	"github.com/pkg/errors"
)

// GetProfile получает подробности профиля с сервера первого типа.
func GetProfile(s *dt.Session) (*dt.Profile, error) {
	p := "http://"

	// 0-ой Post-запрос.
	r0 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"UID": "",
				"VER": s.VER,
				"at":  s.AT,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link,
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/MySettings/MySettings.asp", ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := check.CheckResponse(s, r)
		if err != nil {
			return nil, false, err
		}
		return r.Bytes(), flag, nil
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
	// находящуюся в теле ответа, и найти в ней мапу детей в их ID.
	/*
		parsedHTML, err := html.Parse(bytes.NewReader(b))
		if err != nil {
			return nil, errors.Wrap(err, "parsing HTML")
		}
	*/
	fmt.Println(string(b))

	return nil, nil
}
