package type01

import (
	dt "SchoolServer/libtelco/sessions/data-types"
	"SchoolServer/libtelco/sessions/inner"
	ss "SchoolServer/libtelco/sessions/session"
	"bytes"

	gr "github.com/levigross/grequests"
)

/*
03 тип.
*/

// GetAverageMarkDynReport возвращает динамику среднего балла ученика с сервера первого типа.
func GetAverageMarkDynReport(s *ss.Session, dateBegin, dateEnd, Type, studentID string) (*dt.AverageMarkDynReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
		Data: map[string]string{
			"AT":        s.AT,
			"LoginType": "0",
			"RPTID":     "2",
			"ThmID":     "1",
			"VER":       s.VER,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
		},
	}
	response0, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentAverageMarkDyn.asp", requestOptions0)
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
			"A":         "",
			"ADT":       dateBegin,
			"AT":        s.AT,
			"BACK":      "/asp/Reports/ReportStudentAverageMarkDyn.asp",
			"DDT":       dateEnd,
			"LoginType": "0",
			"MT":        Type,
			"NA":        "",
			"PCLID":     "",
			"PP":        "/asp/Reports/ReportStudentAverageMarkDyn.asp",
			"RP":        "",
			"RPTID":     "2",
			"RT":        "",
			"SID":       studentID,
			"TA":        "",
			"ThmID":     "1",
			"VER":       s.VER,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Reports/ReportStudentAverageMarkDyn.asp",
		},
	}
	response1, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentAverageMarkDyn.asp", requestOptions1)
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
			"A":         "",
			"ADT":       dateBegin,
			"AT":        s.AT,
			"BACK":      "/asp/Reports/ReportStudentAverageMarkDyn.asp",
			"DDT":       dateEnd,
			"LoginType": "0",
			"MT":        Type,
			"NA":        "",
			"PCLID":     "",
			"PP":        "/asp/Reports/ReportStudentAverageMarkDyn.asp",
			"RP":        "",
			"RPTID":     "2",
			"RT":        "",
			"SID":       studentID,
			"TA":        "",
			"ThmID":     "1",
			"VER":       s.VER,
		},
		Headers: map[string]string{
			"Origin":           p + s.Serv.Link,
			"X-Requested-With": "XMLHttpRequest",
			"at":               s.AT,
			"Referer":          p + s.Serv.Link + "/asp/Reports/ReportStudentAverageMarkDyn.asp",
		},
	}
	response2, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/StudentAverageMarkDyn.asp", requestOptions2)
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
	// находящуюся в теле ответа, и найти в ней отчет о динамике среднего балла.
	return inner.AverageMarkDynReportParser(bytes.NewReader(response2.Bytes()))
}
