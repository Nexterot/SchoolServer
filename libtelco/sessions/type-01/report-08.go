// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package type01

import (
	"bytes"
	"fmt"

	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/data-types"
	"github.com/masyagin1998/SchoolServer/libtelco/sessions/inner"
	ss "github.com/masyagin1998/SchoolServer/libtelco/sessions/session"

	"github.com/pkg/errors"

	gr "github.com/levigross/grequests"
)

/*
08 тип.
*/

// GetParentInfoLetterReport возвращает шаблон письма родителям с сервера первого типа.
func GetParentInfoLetterReport(s *ss.Session, reportTypeID, periodID, studentID string) (*dt.ParentInfoLetterReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	r0 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"LoginType": "0",
				"RPTID":     "4",
				"ThmID":     "2",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportParentInfoLetter.asp", ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r.Close()
		}()
		return checkResponse(s, r)
	}
	flag, err := r0()
	if err != nil {
		return nil, errors.Wrap(err, "0 POST")
	}
	if !flag {
		flag, err = r0()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 0 POST")
		}
		if !flag {
			return nil, fmt.Errorf("retry didn't work for 0 POST")
		}
	}

	// 1-ый Post-запрос.
	r1 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"A":          "",
				"AT":         s.AT,
				"BACK":       "/asp/Reports/ReportParentInfoLetter.asp",
				"LoginType":  "0",
				"NA":         "",
				"PCLID":      "",
				"PP":         "/asp/Reports/ReportParentInfoLetter.asp",
				"RP":         "",
				"RPTID":      "4",
				"RT":         "",
				"ReportType": reportTypeID,
				"SID":        studentID,
				"TA":         "",
				"TERMID":     periodID,
				"ThmID":      "2",
				"VER":        s.VER,
				"drWeek":     "",
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/ReportParentInfoLetter.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportParentInfoLetter.asp", ro)
		if err != nil {
			return false, err
		}
		defer func() {
			_ = r.Close()
		}()
		return checkResponse(s, r)
	}
	flag, err = r1()
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

	// 2-ой Post-запрос.
	r2 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"A":          "",
				"AT":         s.AT,
				"BACK":       "/asp/Reports/ReportParentInfoLetter.asp",
				"LoginType":  "0",
				"NA":         "",
				"PCLID":      "",
				"PP":         "/asp/Reports/ReportParentInfoLetter.asp",
				"RP":         "",
				"RPTID":      "4",
				"RT":         "",
				"ReportType": reportTypeID,
				"SID":        studentID,
				"TA":         "",
				"TERMID":     periodID,
				"ThmID":      "2",
				"VER":        s.VER,
				"drWeek":     "",
			},
			Headers: map[string]string{
				"Origin":           p + s.Serv.Link,
				"X-Requested-With": "XMLHttpRequest",
				"at":               s.AT,
				"Referer":          p + s.Serv.Link + "/asp/Reports/ReportParentInfoLetter.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ParentInfoLetter.asp", ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := checkResponse(s, r)
		return r.Bytes(), flag, err
	}
	b, flag, err := r2()
	if err != nil {
		return nil, errors.Wrap(err, "2 POST")
	}
	if !flag {
		b, flag, err = r2()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 2 POST")
		}
		if !flag {
			return nil, fmt.Errorf("retry didn't work for 2 POST")
		}
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа и найти в ней шаблон письма родителям.
	return inner.ParentInfoLetterReportParser(bytes.NewReader(b))
}
