// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package email

import (
	"fmt"

	gr "github.com/levigross/grequests"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"
	"github.com/pkg/errors"
)

// GetAddressBook возвращает список всех возможных адресатов с сервера первого типа.
func GetAddressBook(s *dt.Session) (*dt.AddressBook, error) {
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
		return nil, errors.Wrap(err, "0 GET")
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
		return nil, errors.Wrap(err, "1 POST")
	}
	if !flag {
		flag, err = r1()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 1 POST")
		}
		if !flag {
			return nil, fmt.Errorf("retry didn't work for 1 POST")
		}
	}

	// 2-ой Get-запрос (не дублирующийся).
	r2 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + fmt.Sprintf("/asp/Messages/MailBox.asp?AT=%s&VER=%s", s.AT, s.VER),
			},
			Data: map[string]string{},
		}
		_, err := s.Sess.Get(p+s.Serv.Link+fmt.Sprintf("/asp/Messages/composemessage.asp?at=%v&ver=%v", s.AT, s.VER), ro)
		return true, err
	}
	_, err = r2()
	if err != nil {
		return nil, errors.Wrap(err, "2 GET")
	}

	// 3-ий Get-запрос (не дублирующийся).
	r3 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"X-Requested-With": "XMLHttpRequest",
				"Referer":          p + s.Serv.Link + fmt.Sprintf("/asp/Messages/composemessage.asp?at=%v&ver=%v", s.AT, s.VER),
			},
			Data: map[string]string{},
		}
		_, err := s.Sess.Get(p+s.Serv.Link+"/vendor/pages/css/print-tables.min.css", ro)
		return true, err
	}
	_, err = r3()
	if err != nil {
		return nil, errors.Wrap(err, "3 GET")
	}

	// 4-ый Get-запрос (не дублирующийся).
	r4 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + fmt.Sprintf("/asp/Messages/composemessage.asp?at=%v&ver=%v", s.AT, s.VER),
			},
			Data: map[string]string{},
		}
		_, err := s.Sess.Get(p+s.Serv.Link+fmt.Sprintf("/asp/messages/addressbook.asp?at=%v&ver=%v&F=COMPOSE&FN=ATO&FA=LTO", s.AT, s.VER), ro)
		return true, err
	}
	_, err = r4()
	if err != nil {
		return nil, errors.Wrap(err, "4 GET")
	}

	// 5-ый Get-запрос (не дублирующийся).
	r5 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + fmt.Sprintf("/asp/messages/addressbook.asp?at=%v&ver=%v&F=COMPOSE&FN=ATO&FA=LTO", s.AT, s.VER),
			},
			Data: map[string]string{},
		}
		_, err := s.Sess.Get(p+s.Serv.Link+fmt.Sprintf("/asp/Messages/addrbkbottom.asp?AT=%v&VER=%v", s.AT, s.VER), ro)
		return true, err
	}
	_, err = r5()
	if err != nil {
		return nil, errors.Wrap(err, "5 GET")
	}

	// 6-ой Get-запрос (не дублирующийся).
	r6 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + fmt.Sprintf("/asp/messages/addressbook.asp?at=%v&ver=%v&F=COMPOSE&FN=ATO&FA=LTO", s.AT, s.VER),
			},
			Data: map[string]string{},
		}
		_, err := s.Sess.Get(p+s.Serv.Link+fmt.Sprintf("/asp/Messages/addrbkright.asp?AT=%v&F=COMPOSE&FN=ATO&FA=LTO&VER=%v", s.AT, s.VER), ro)
		return true, err
	}
	_, err = r6()
	if err != nil {
		return nil, errors.Wrap(err, "6 GET")
	}

	// 7-ой Get-запрос (не дублирующийся).
	r7 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Headers: map[string]string{
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + fmt.Sprintf("/asp/messages/addressbook.asp?at=%v&ver=%v&F=COMPOSE&FN=ATO&FA=LTO", s.AT, s.VER),
			},
			Data: map[string]string{},
		}
		b, err := s.Sess.Get(p+s.Serv.Link+fmt.Sprintf("/asp/Messages/addrbkleft.asp?AT=%v&VER=%v", s.AT, s.VER), ro)
		fmt.Println(b.String())
		return true, err
	}
	_, err = r7()
	if err != nil {
		return nil, errors.Wrap(err, "7 GET")
	}

	return nil, nil
}
