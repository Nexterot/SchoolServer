// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package type01

import (
	dt "SchoolServer/libtelco/sessions/data-types"
	"SchoolServer/libtelco/sessions/inner"
	ss "SchoolServer/libtelco/sessions/session"
	"bytes"

	gr "github.com/levigross/grequests"
)

/*
01 тип.
*/

// GetTotalMarkReport возвращает успеваемость ученика с сервера первого типа.
func GetTotalMarkReport(s *ss.Session, studentID string) (*dt.TotalMarkReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
		Data: map[string]string{
			"AT":        s.AT,
			"LoginType": "0",
			"RPTID":     "0",
			"ThmID":     "1",
			"VER":       s.VER,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
		},
	}
	response0, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentTotalMarks.asp", requestOptions0)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response0.Close()
	}()
	if err := checkResponse(s, response0); err != nil {
		return nil, err
	}

	// 1-ый Post-запрос.
	requestOption1 := &gr.RequestOptions{
		Data: map[string]string{
			"A":         "",
			"AT":        s.AT,
			"BACK":      "/asp/Reports/ReportStudentTotalMarks.asp",
			"LoginType": "0",
			"NA":        "",
			"PCLID":     "",
			"PP":        "/asp/Reports/ReportStudentTotalMarks.asp",
			"RP":        "",
			"RPTID":     "0",
			"RT":        "",
			"SID":       studentID,
			"TA":        "",
			"ThmID":     "1",
			"VER":       s.VER,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Reports/ReportStudentTotalMarks.asp",
		},
	}
	response1, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentTotalMarks.asp", requestOption1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()

	if err := checkResponse(s, response1); err != nil {
		return nil, err
	}

	// 2-ой Post-запрос.
	requestOption2 := &gr.RequestOptions{
		Data: map[string]string{
			"A":         "",
			"AT":        s.AT,
			"BACK":      "/asp/Reports/ReportStudentTotalMarks.asp",
			"ISTF":      "0",
			"LoginType": "0",
			"NA":        "",
			"PCLID":     "",
			"PP":        "/asp/Reports/ReportStudentTotalMarks.asp",
			"RP":        "0",
			"RPTID":     "0",
			"RT":        "0",
			"SID":       studentID,
			"TA":        "",
			"ThmID":     "1",
			"VER":       s.VER,
		},
		Headers: map[string]string{
			"Origin":           p + s.Serv.Link,
			"X-Requested-With": "XMLHttpRequest",
			"at":               s.AT,
			"Referer":          p + s.Serv.Link + "/asp/Reports/ReportStudentTotalMarks.asp",
		},
	}
	response2, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/StudentTotalMarks.asp", requestOption2)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response2.Close()
	}()

	if err := checkResponse(s, response2); err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней отчет об успеваемости ученика.
	return inner.TotalMarkReportParser(bytes.NewReader(response2.Bytes()))
}
