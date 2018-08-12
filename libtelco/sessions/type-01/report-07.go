// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package type01

import (
	dt "SchoolServer/libtelco/sessions/data-types"
	"SchoolServer/libtelco/sessions/inner"
	ss "SchoolServer/libtelco/sessions/session"
	"bytes"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	gr "github.com/levigross/grequests"
)

/*
07 тип.
*/

// GetJournalAccessReport возвращает отчет о доступе к журналу с сервера первого типа.
func GetJournalAccessReport(s *ss.Session, studentID string) (*dt.JournalAccessReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	r0 := func() (bool, error) {
		ro := &gr.RequestOptions{
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
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/JournalAccess.asp", ro)
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
	r1 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			JSON: json,
			Headers: map[string]string{
				"Origin":           p + s.Serv.Link,
				"X-Requested-With": "XMLHttpRequest",
				"at":               s.AT,
				"Referer":          p + s.Serv.Link + "/asp/Reports/ReportJournalAccess.asp",
			},
		}
		r, err := s.Sess.Post(p+s.Serv.Link+"/webapi/reports/journal_access/initfilters", ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := checkResponse(s, r)
		return r.Bytes(), flag, err
	}
	b, flag, err := r1()
	if err != nil {
		return nil, errors.Wrap(err, "1 POST")
	}
	if !flag {
		b, flag, err = r1()
		if err != nil {
			return nil, errors.Wrap(err, "retrying 1 POST")
		}
		if !flag {
			return nil, fmt.Errorf("retry didn't work for 1 POST")
		}
	}
	CLID := string(b)
	index := strings.Index(CLID, "\"value\":\"")
	if index == -1 {
		return nil, fmt.Errorf("Invalid begin SID substring \"%s\"", CLID)
	}
	CLID = CLID[index+len("\"value\":\""):]
	index = strings.Index(CLID, "\"")
	if index == -1 {
		return nil, fmt.Errorf("Invalid end SID substring \"%s\"", CLID)
	}
	CLID = CLID[:index]

	// 2-ой Post-запрос.
	r2 := func() ([]byte, bool, error) {
		ro := &gr.RequestOptions{
			Data: map[string]string{
				"A":         "",
				"AT":        s.AT,
				"BACK":      "/asp/Reports/ReportJournalAccess.asp",
				"LoginType": "0",
				"NA":        "",
				"PCLID_IUP": CLID,
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
		r, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/JournalAccess.asp", ro)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			_ = r.Close()
		}()
		flag, err := checkResponse(s, r)
		return r.Bytes(), flag, err
	}
	b, flag, err = r2()
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
	// находящуюся в теле ответа, и найти в ней отчет о доступе к журналу.
	return inner.JournalAccessReportParser(bytes.NewReader(b))
}
