// Copyright (C) 2018 Mikhail Masyagin

package base

import (
	"fmt"
	"strings"

	gr "github.com/levigross/grequests"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"
	"github.com/pkg/errors"
)

// ChangePassword меняет пароль на сервере первого типа.
func ChangePassword(s *dt.Session, oldMD5, newMD5 string) error {
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

	str := string(b)
	ind := strings.Index(str, `name="UID" value="`)
	if ind == -1 {
		return fmt.Errorf("invalid HTML\n %v", str)
	}
	str = str[ind+len(`name="UID" value="`):]
	ind = strings.Index(str, `"`)
	if ind == -1 {
		return fmt.Errorf("invalid HTML\n %v", str)
	}
	str = str[:ind]

	// 1-ый Post-запрос.
	r1 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":     s.AT,
				"act":    "prepare",
				"userId": str,
			},
			Headers: map[string]string{
				"Origin":           p + s.Serv.Link,
				"X-Requested-With": "XMLHttpRequest",
				"at":               s.AT,
				"Referer":          p + s.Serv.Link + "/asp/MySettings/MySettings.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/ajax/ChangePassword.asp", ro)
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

	// 2-ой Post-запрос.
	r2 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":     s.AT,
				"NP3":    newMD5,
				"OP2":    oldMD5,
				"act":    "save",
				"userId": str,
			},
			Headers: map[string]string{
				"Origin":           p + s.Serv.Link,
				"X-Requested-With": "XMLHttpRequest",
				"at":               s.AT,
				"Referer":          p + s.Serv.Link + "/asp/MySettings/MySettings.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/ajax/ChangePassword.asp", ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r.Close()
		}()
		return check.CheckResponse(s, r)
	}
	flag, err = r2()
	if err != nil {
		return errors.Wrap(err, "2 POST")
	}
	if !flag {
		flag, err = r2()
		if err != nil {
			return errors.Wrap(err, "retrying 2 POST")
		}
		if !flag {
			return fmt.Errorf("retry didn't work for 2 POST")
		}
	}

	return nil
}
