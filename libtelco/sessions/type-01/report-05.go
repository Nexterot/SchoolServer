// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package type01

import (
	dt "SchoolServer/libtelco/sessions/data-types"
	"SchoolServer/libtelco/sessions/inner"
	ss "SchoolServer/libtelco/sessions/session"
	"bytes"
	"fmt"

	gr "github.com/levigross/grequests"
)

/*
05 тип.
*/

// GetStudentTotalReport возвращает отчет о посещениях ученика с сервера первого типа.
func GetStudentTotalReport(s *ss.Session, dateBegin, dateEnd, studentID string) (*dt.StudentTotalReport, error) {
	p := "http://"

	r0 := func() (bool, error) {
		// 0-ой Post-запрос.
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"LoginType": "0",
				"RPTID":     "1",
				"ThmID":     "2",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentTotal.asp", ro)
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
		return nil, err
	}
	if !flag {
		flag, err = r0()
		if err != nil {
			return nil, err
		}
		if !flag {
			return nil, fmt.Errorf("Retry didn't work")
		}
	}

	// 1-ый Post-запрос.
	r1 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"A":         "",
				"ADT":       dateBegin,
				"AT":        s.AT,
				"BACK":      "/asp/Reports/ReportStudentTotal.asp",
				"DDT":       dateEnd,
				"LoginType": "0",
				"NA":        "",
				"PCLID":     "",
				"PP":        "/asp/Reports/ReportStudentTotal.asp",
				"RP":        "",
				"RPTID":     "1",
				"RT":        "",
				"SID":       studentID,
				"TA":        "",
				"ThmID":     "2",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/ReportStudentTotal.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentTotal.asp", ro)
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
		return nil, err
	}
	if !flag {
		flag, err = r1()
		if err != nil {
			return nil, err
		}
		if !flag {
			return nil, fmt.Errorf("Retry didn't work")
		}
	}

	r2 := func() ([]byte, bool, error) {
		// 2-ой Post-запрос.
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"A":         "",
				"ADT":       dateBegin,
				"AT":        s.AT,
				"BACK":      "/asp/Reports/ReportStudentTotal.asp",
				"DDT":       dateEnd,
				"LoginType": "0",
				"NA":        "",
				"PCLID":     "",
				"PP":        "/asp/Reports/ReportStudentTotal.asp",
				"RP":        "",
				"RPTID":     "1",
				"RT":        "",
				"SID":       studentID,
				"TA":        "",
				"ThmID":     "2",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":           p + s.Serv.Link,
				"X-Requested-With": "XMLHttpRequest",
				"at":               s.AT,
				"Referer":          p + s.Serv.Link + "/asp/Reports/ReportStudentTotal.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/StudentTotal.asp", ro)
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
		return nil, err
	}
	if !flag {
		b, flag, err = r2()
		if err != nil {
			return nil, err
		}
		if !flag {
			return nil, fmt.Errorf("Retry didn't work")
		}
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней отчет о посещених ученика.
	return inner.StudentTotalReportParser(bytes.NewReader(b))
}
