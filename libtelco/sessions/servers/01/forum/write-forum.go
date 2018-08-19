package forum

import (
	"fmt"

	gr "github.com/levigross/grequests"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/check"
	"github.com/pkg/errors"
)

// CreateForumTheme создаёт новую тему на форуме на сервере первого типа.
func CreateForumTheme(s *dt.Session, page, name, message string) error {
	p := "http://"

	// 0-ой POST-запрос.
	r0 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"BACK":      "",
				"DELARR":    "",
				"LoginType": "0",
				"PAGE":      page,
				"PAGESIZE":  "25",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + fmt.Sprintf("/asp/Forum/Forum.asp?AT=%s&VER=%s", s.AT, s.VER),
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Forum/NewThread.asp", ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := check.CheckResponse(s, r)
		return flag, err
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

	// 1-ый POST-запрос.
	r1 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"LoginType": "0",
				"MESSAGE":   message,
				"NAME":      name,
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Forum/NewThread.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Forum/SaveNewThread.asp?PAGE=1&PAGESIZE=25&TPAGE=1", ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := check.CheckResponse(s, r)
		return flag, err
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

	return nil
}

// CreateForumThemeMessage создаёт новое сообщение в теме на форуме на сервере первого типа.
func CreateForumThemeMessage(s *dt.Session, page, message, TID string) error {
	p := "http://"

	// 0-ой POST-запрос.
	r0 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"DELARR":    "",
				"LoginType": "0",
				"MESSAGE":   message,
				"PAGE":      page,
				"PAGESIZE":  "25",
				"TID":       TID,
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + fmt.Sprintf("/asp/Forum/Forum.asp?AT=%s&VER=%s", s.AT, s.VER),
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+fmt.Sprintf("/asp/Forum/AddReply.asp?TPAGE=%s", page), ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := check.CheckResponse(s, r)
		return flag, err
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
