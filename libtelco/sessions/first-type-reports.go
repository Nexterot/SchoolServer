// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

/*
Package sessions - данный файл содержит в себе функции для обработки 1 типа сайтов.
*/
package sessions

import (
	"bytes"
	"fmt"

	gr "github.com/levigross/grequests"
)

// getTotalMarkReportFirst возвращает успеваемость ученика с сервера первого типа.
func (s *Session) getTotalMarkReportFirst() (*TotalMarkReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
		Data: map[string]string{
			"AT":        s.at,
			"LoginType": "0",
			"RPTID":     "0",
			"ThmID":     "1",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
		},
	}
	response0, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentTotalMarks.asp", requestOptions0)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response0.Close()
	}()
	if err := s.checkResponse(response0); err != nil {
		return nil, err
	}

	// 1-ый Post-запрос.
	requestOption1 := &gr.RequestOptions{
		Data: map[string]string{
			"A":         "",
			"AT":        s.at,
			"BACK":      "/asp/Reports/ReportStudentTotalMarks.asp",
			"ISTF":      "0",
			"LoginType": "0",
			"NA":        "",
			"PCLID":     "",
			"PP":        "/asp/Reports/ReportStudentTotalMarks.asp",
			"RP":        "0",
			"RPTID":     "0",
			"RT":        "0",
			"SID":       "11198",
			"TA":        "",
			"ThmID":     "1",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Origin":           p + s.Serv.Link,
			"X-Requested-With": "XMLHttpRequest",
			"at":               s.at,
			"Referer":          p + s.Serv.Link + "/asp/Reports/ReportStudentTotalMarks.asp",
		},
	}
	response1, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/StudentTotalMarks.asp", requestOption1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()
	if err := s.checkResponse(response1); err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа и найти в ней оценки.
	return totalMarkReportParser(bytes.NewReader(response1.Bytes()))
}

// getAverageMarkReportFirst возвращает средние баллы ученика с сервера первого типа.
func (s *Session) getAverageMarkReportFirst(dateBegin, dateEnd, Type string) (*AverageMarkReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
		Data: map[string]string{
			"AT":        s.at,
			"LoginType": "0",
			"RPTID":     "1",
			"ThmID":     "1",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
		},
	}
	response0, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentAverageMark.asp", requestOptions0)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response0.Close()
	}()
	if err := s.checkResponse(response0); err != nil {
		return nil, err
	}

	// 1-ый Post-запрос.
	requestOptions1 := &gr.RequestOptions{
		Data: map[string]string{
			"A":         "",
			"ADT":       dateBegin,
			"AT":        s.at,
			"BACK":      "/asp/Reports/ReportStudentAverageMark.asp",
			"DDT":       dateEnd,
			"LoginType": "0",
			"MT":        Type,
			"NA":        "",
			"PCLID":     "",
			"PP":        "/asp/Reports/ReportStudentAverageMark.asp",
			"RP":        "",
			"RPTID":     "1",
			"RT":        "",
			"SID":       "11198",
			"TA":        "",
			"ThmID":     "1",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Origin":           p + s.Serv.Link,
			"X-Requested-With": "XMLHttpRequest",
			"at":               s.at,
			"Referer":          p + s.Serv.Link + "/asp/Reports/ReportStudentAverageMark.asp",
		},
	}
	response1, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/StudentAverageMark.asp", requestOptions1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()
	if err := s.checkResponse(response1); err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа и найти в ней оценки.
	return averageMarkReportParser(bytes.NewReader(response1.Bytes()))
}

// getAverageMarkDynReportFirst возвращает динамику среднего балла ученика с сервера первого типа.
func (s *Session) getAverageMarkDynReportFirst(dateBegin, dateEnd, Type string) (*AverageMarkDynReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
		Data: map[string]string{
			"AT":        s.at,
			"LoginType": "0",
			"RPTID":     "2",
			"ThmID":     "1",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
		},
	}
	response0, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentAverageMarkDyn.asp", requestOptions0)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response0.Close()
	}()
	if err := s.checkResponse(response0); err != nil {
		return nil, err
	}

	// 1-ый Post-запрос.
	requestOptions1 := &gr.RequestOptions{
		Data: map[string]string{
			"A":         "",
			"ADT":       dateBegin,
			"AT":        s.at,
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
			"SID":       "11198",
			"TA":        "",
			"ThmID":     "1",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Origin":           p + s.Serv.Link,
			"X-Requested-With": "XMLHttpRequest",
			"at":               s.at,
			"Referer":          p + s.Serv.Link + "/asp/Reports/ReportStudentAverageMarkDyn.asp",
		},
	}
	response1, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/StudentAverageMarkDyn.asp", requestOptions1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()
	if err := s.checkResponse(response1); err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа и найти в ней оценки.
	return averageMarkDynReportParser(bytes.NewReader(response1.Bytes()))
}

// getStudentGradeReportFirst возвращает отчет об успеваемости ученика по предмету с сервера первого типа.
func (s *Session) getStudentGradeReportFirst(dateBegin, dateEnd, SubjectName string) (*StudentGradeReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
		Data: map[string]string{
			"AT":        s.at,
			"LoginType": "0",
			"RPTID":     "0",
			"ThmID":     "2",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
		},
	}
	response0, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentGrades.asp", requestOptions0)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response0.Close()
	}()
	if err := s.checkResponse(response0); err != nil {
		return nil, err
	}

	// 1-ый Post-запрос.
	requestOptions1 := &gr.RequestOptions{
		Data: map[string]string{
			"A":         "",
			"ADT":       dateBegin,
			"AT":        s.at,
			"BACK":      "/asp/Reports/ReportStudentGrades.asp",
			"DDT":       dateEnd,
			"LoginType": "0",
			"NA":        "",
			"PCLID_IUP": "10169_0",
			"PP":        "/asp/Reports/ReportStudentGrades.asp",
			"RP":        "",
			"RPTID":     "0",
			"RT":        "",
			"SCLID":     SubjectName,
			"SID":       "11198",
			"TA":        "",
			"ThmID":     "2",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Reports/ReportStudentGrades.asp",
		},
	}
	response1, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentGrades.asp", requestOptions1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()
	if err := s.checkResponse(response1); err != nil {
		return nil, err
	}

	// 2-ой POST-запрос.
	requestOptions2 := &gr.RequestOptions{
		Data: map[string]string{
			"A":         "",
			"ADT":       dateBegin,
			"AT":        s.at,
			"BACK":      "/asp/Reports/ReportStudentGrades.asp",
			"DDT":       dateEnd,
			"LoginType": "0",
			"NA":        "",
			"PCLID_IUP": "10169_0",
			"PP":        "/asp/Reports/ReportStudentGrades.asp",
			"RP":        "",
			"RPTID":     "0",
			"RT":        "",
			"SCLID":     SubjectName,
			"SID":       "11198",
			"TA":        "",
			"ThmID":     "2",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Origin":           p + s.Serv.Link,
			"X-Requested-With": "XMLHttpRequest",
			"at":               s.at,
			"Referer":          p + s.Serv.Link + "/asp/Reports/ReportStudentGrades.asp",
		},
	}
	response2, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/StudentGrades.asp", requestOptions2)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response2.Close()
	}()
	if err := s.checkResponse(response2); err != nil {
		return nil, err
	}
	return studentGradeReportParser(bytes.NewReader(response2.Bytes()))
}

// getStudentTotalReportFirst возвращает отчет о посещениях ученика с сервера первого типа.
func (s *Session) getStudentTotalReportFirst(dateBegin, dateEnd string) (*StudentTotalReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
		Data: map[string]string{
			"AT":        s.at,
			"LoginType": "0",
			"RPTID":     "1",
			"ThmID":     "2",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
		},
	}
	response0, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentTotal.asp", requestOptions0)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response0.Close()
	}()
	if err := s.checkResponse(response0); err != nil {
		return nil, err
	}

	// 1-ый Post-запрос.
	requestOptions1 := &gr.RequestOptions{
		Data: map[string]string{
			"A":         "",
			"ADT":       dateBegin,
			"AT":        s.at,
			"BACK":      "/asp/Reports/ReportStudentTotal.asp",
			"DDT":       dateEnd,
			"LoginType": "0",
			"NA":        "",
			"PCLID":     "",
			"PP":        "/asp/Reports/ReportStudentTotal.asp",
			"RP":        "",
			"RPTID":     "1",
			"RT":        "",
			"SID":       "11198",
			"TA":        "",
			"ThmID":     "2",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Origin":           p + s.Serv.Link,
			"X-Requested-With": "XMLHttpRequest",
			"at":               s.at,
			"Referer":          p + s.Serv.Link + "/asp/Reports/ReportStudentTotal.asp",
		},
	}
	response1, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/StudentTotal.asp", requestOptions1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()
	if err := s.checkResponse(response1); err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа и найти в ней оценки.
	return studentTotalReportParser(bytes.NewReader(response1.Bytes()))
}

// getJournalAccessReportFirst возвращает отчет о доступе к журналу с сервера первого типа.
func (s *Session) getJournalAccessReportFirst() (*JournalAccessReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
		Data: map[string]string{
			"AT":        s.at,
			"LoginType": "0",
			"RPTID":     "3",
			"ThmID":     "2",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
		},
	}
	response0, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentTotal.asp", requestOptions0)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response0.Close()
	}()
	if err := s.checkResponse(response0); err != nil {
		return nil, err
	}

	// 1-ый Post-запрос.
	requestOptions1 := &gr.RequestOptions{
		Data: map[string]string{
			"A":         "",
			"AT":        s.at,
			"BACK":      "/asp/Reports/ReportJournalAccess.asp",
			"LoginType": "0",
			"NA":        "",
			"PCLID":     "",
			"PP":        "/asp/Reports/ReportJournalAccess.asp",
			"RP":        "",
			"RPTID":     "3",
			"RT":        "",
			"SID":       "11198",
			"TA":        "",
			"ThmID":     "2",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Origin":           p + s.Serv.Link,
			"X-Requested-With": "XMLHttpRequest",
			"at":               s.at,
			"Referer":          p + s.Serv.Link + "/asp/Reports/ReportJournalAccess.asp",
		},
	}
	response1, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/JournalAccess.asp", requestOptions1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()
	if err := s.checkResponse(response1); err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа и найти в ней оценки.
	// Сделай парсер.
	fmt.Println(string(response1.Bytes()))
	fmt.Println()
	fmt.Println()
	return nil, nil
}

// getParentInfoLetterReportFirst возвращает шаблон письма родителям с сервера первого типа.
func (s *Session) getParentInfoLetterReportFirst(studentID, reportTypeID, periodID string) (*ParentInfoLetterReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
		Data: map[string]string{
			"AT":        s.at,
			"LoginType": "0",
			"RPTID":     "4",
			"ThmID":     "2",
			"VER":       s.ver,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
		},
	}
	response0, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/ReportParentInfoLetter.asp", requestOptions0)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response0.Close()
	}()
	if err := s.checkResponse(response0); err != nil {
		return nil, err
	}

	// 1-ый Post-запрос.
	requestOptions1 := &gr.RequestOptions{
		Data: map[string]string{
			"A":          "",
			"AT":         s.at,
			"BACK":       "/asp/Reports/ReportParentInfoLetter.asp",
			"LoginType":  "0",
			"NA":         "",
			"PCLID":      "",
			"PP":         "/asp/Reports/ReportParentInfoLetter.asp",
			"RP":         "",
			"RPTID":      "4",
			"RT":         "",
			"ReportType": "1",
			"SID":        "11198",
			"TA":         "",
			"TERMID":     "10067",
			"ThmID":      "2",
			"VER":        s.ver,
			"drWeek":     "",
		},
		Headers: map[string]string{
			"Origin":           p + s.Serv.Link,
			"X-Requested-With": "XMLHttpRequest",
			"at":               s.at,
			"Referer":          p + s.Serv.Link + "/asp/Reports/ReportParentInfoLetter.asp",
		},
	}
	response1, err := s.sess.Post(p+s.Serv.Link+"/asp/Reports/ParentInfoLetter.asp", requestOptions1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()
	if err := s.checkResponse(response1); err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа и найти в ней оценки.
	// Сделай парсер.
	fmt.Println(string(response1.Bytes()))
	fmt.Println()
	fmt.Println()
	return nil, nil
}
