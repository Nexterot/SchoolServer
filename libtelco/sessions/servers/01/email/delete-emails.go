// Copyright (C) 2018 Mikhail Masyagin

package email

import (
	"fmt"
	"strings"

	gr "github.com/levigross/grequests"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"
	"github.com/pkg/errors"
)

// DeleteEmails удаяет заданные сообщения с сервера первого типа.
func DeleteEmails(s *dt.Session, boxID string, emailIDs []string) error {
	p := "http://"

	// 0-ой Get-запрос (не дублирующийся).
	r0 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   "http://62.117.74.43/",
			},
			Data: map[string]string{},
		}
		_, err := s.Sess.Get(p+s.Serv.Link+fmt.Sprintf("/asp/Messages/MailBox.asp?AT=%s&VER=%s", s.AT, s.VER), ro)
		return true, err
	}
	_, err := r0()
	if err != nil {
		return errors.Wrap(err, "0 GET")
	}

	// 1-ый POST-запрос.
	r1 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Origin":           p + s.Serv.Link,
				"X-Requested-With": "XMLHttpRequest",
				"Referer":          p + s.Serv.Link + fmt.Sprintf("/asp/Messages/MailBox.asp?AT=%s&VER=%s", s.AT, s.VER),
			},
			Data: map[string]string{},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+
			fmt.Sprintf("/asp/ajax/GetMessagesAjax.asp?AT=%v&nBoxID=1&jtStartIndex=0&jtPageSize=10&jtSorting=Sent%%20DESC", s.AT), ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := check.CheckResponse(s, r)
		return flag, err
	}
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

	// 2-ой POST-запрос.
	r2 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Origin":           p + s.Serv.Link,
				"X-Requested-With": "XMLHttpRequest",
				"at":               s.AT,
				"Content-Type":     "application/x-www-form-urlencoded",
				"Referer":          p + s.Serv.Link + fmt.Sprintf("/asp/Messages/MailBox.asp?AT=%s&VER=%s", s.AT, s.VER),
			},
		}
		body := fmt.Sprintf("AT=%v", s.AT)
		for _, id := range emailIDs {
			body += fmt.Sprintf("&deletedMessages=%v", id)
		}
		body += fmt.Sprintf("&nBoxId=%v", boxID)
		ro.RequestBody = strings.NewReader(body)
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/ajax/DeleteMessagesAjax.asp", ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := check.CheckResponse(s, r)
		return flag, err
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
