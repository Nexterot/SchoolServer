package type01

import (
	dt "SchoolServer/libtelco/sessions/data-types"
	"SchoolServer/libtelco/sessions/inner"
	ss "SchoolServer/libtelco/sessions/session"
	"bytes"

	gr "github.com/levigross/grequests"
)

/*
08 тип.
*/

// GetParentInfoLetterReport возвращает шаблон письма родителям с сервера первого типа.
func GetParentInfoLetterReport(s *ss.Session, reportTypeID, periodID, studentID string) (*dt.ParentInfoLetterReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
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
	response0, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportParentInfoLetter.asp", requestOptions0)
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
	requestOptions1 := &gr.RequestOptions{
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
	response1, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportParentInfoLetter.asp", requestOptions1)
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
	requestOptions2 := &gr.RequestOptions{
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
	response2, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ParentInfoLetter.asp", requestOptions2)
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
	// находящуюся в теле ответа и найти в ней шаблон письма родителям.
	return inner.ParentInfoLetterReportParser(bytes.NewReader(response2.Bytes()))
}
