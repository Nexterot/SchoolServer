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
07 тип.
*/

// GetJournalAccessReport возвращает отчет о доступе к журналу с сервера первого типа.
func GetJournalAccessReport(s *ss.Session, studentID string) (*dt.JournalAccessReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
		Data: map[string]string{
			"AT":        s.AT,
			"LoginType": "0",
			"RPTID":     "3",
			"ThmID":     "2",
			"VER":       s.VER,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
		},
	}
	response0, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/JournalAccess.asp", requestOptions0)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response0.Close()
	}()
	if err := checkResponse(s, response0); err != nil {
		return nil, err
	}

	type Filter struct {
		ID    string `json:"filterId"`
		Value string `json:"filterValue"`
	}

	type SelectedData struct {
		SelectedData []Filter `json:"selectedData"`
	}

	json := SelectedData{
		SelectedData: []Filter{Filter{"SID", studentID}},
	}

	// 1-ый Post-запрос.
	requestOptions1 := &gr.RequestOptions{
		JSON: json,
		Headers: map[string]string{
			"Origin":           p + s.Serv.Link,
			"X-Requested-With": "XMLHttpRequest",
			"at":               s.AT,
			"Referer":          p + s.Serv.Link + "/asp/Reports/ReportJournalAccess.asp",
		},
	}
	response1, err := s.Sess.Post(p+s.Serv.Link+"/webapi/reports/journal_access/initfilters", requestOptions1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()
	fmt.Println(string(response1.Bytes()))
	fmt.Println()
	if err := checkResponse(s, response1); err != nil {
		return nil, err
	}

	// 2-ой Post-запрос.
	requestOptions2 := &gr.RequestOptions{
		Data: map[string]string{
			"A":         "",
			"AT":        s.AT,
			"BACK":      "/asp/Reports/ReportJournalAccess.asp",
			"LoginType": "0",
			"NA":        "",
			"PCLID_IUP": "10169_0",
			"PP":        "/asp/Reports/ReportJournalAccess.asp",
			"RP":        "",
			"RPTID":     "3",
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
			"Referer":          p + s.Serv.Link + "/asp/Reports/ReportJournalAccess.asp",
		},
	}
	response2, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/JournalAccess.asp", requestOptions2)
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
	// находящуюся в теле ответа, и найти в ней отчет о доступе к журналу.
	return inner.JournalAccessReportParser(bytes.NewReader(response2.Bytes()))
}
