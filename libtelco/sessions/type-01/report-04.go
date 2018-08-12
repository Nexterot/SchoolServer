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
04 тип.
*/

// GetStudentGradeReport возвращает отчет об успеваемости ученика по предмету с сервера первого типа.
func GetStudentGradeReport(s *ss.Session, dateBegin, dateEnd, subjectID, studentID string) (*dt.StudentGradeReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	r0 := func() (bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"AT":        s.AT,
				"LoginType": "0",
				"RPTID":     "0",
				"ThmID":     "2",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentGrades.asp", ro)
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

	r1 := func() (bool, error) {
		// 1-ый Post-запрос.
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"A":         "",
				"ADT":       dateBegin,
				"AT":        s.AT,
				"BACK":      "/asp/Reports/ReportStudentGrades.asp",
				"DDT":       dateEnd,
				"LoginType": "0",
				"NA":        "",
				"PCLID_IUP": "10169_0",
				"PP":        "/asp/Reports/ReportStudentGrades.asp",
				"RP":        "",
				"RPTID":     "0",
				"RT":        "",
				"SCLID":     subjectID,
				"SID":       studentID,
				"TA":        "",
				"ThmID":     "2",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":                    p + s.Serv.Link,
				"Upgrade-Insecure-Requests": "1",
				"Referer":                   p + s.Serv.Link + "/asp/Reports/ReportStudentGrades.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentGrades.asp", ro)
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

	// 2-ой POST-запрос.
	r2 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"A":         "",
				"ADT":       dateBegin,
				"AT":        s.AT,
				"BACK":      "/asp/Reports/ReportStudentGrades.asp",
				"DDT":       dateEnd,
				"LoginType": "0",
				"NA":        "",
				"PCLID_IUP": "10169_0",
				"PP":        "/asp/Reports/ReportStudentGrades.asp",
				"RP":        "",
				"RPTID":     "0",
				"RT":        "",
				"SCLID":     subjectID,
				"SID":       studentID,
				"TA":        "",
				"ThmID":     "2",
				"VER":       s.VER,
			},
			Headers: map[string]string{
				"Origin":           p + s.Serv.Link,
				"X-Requested-With": "XMLHttpRequest",
				"at":               s.AT,
				"Referer":          p + s.Serv.Link + "/asp/Reports/ReportStudentGrades.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/StudentGrades.asp", ro)
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
	// находящуюся в теле ответа, и найти в ней отчет об успеваемости ученика по предмету.
	return inner.StudentGradeReportParser(bytes.NewReader(b))
}
