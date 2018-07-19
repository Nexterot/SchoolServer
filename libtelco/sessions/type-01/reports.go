// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

/*
Package type01 - данный файл содержит в себе функции для обработки 1 типа сайтов.
*/
package type01

import (
	dt "SchoolServer/libtelco/sessions/data-types"
	"SchoolServer/libtelco/sessions/inner"
	ss "SchoolServer/libtelco/sessions/session"
	"bytes"

	gr "github.com/levigross/grequests"
)

/*
Получение отчетов.
*/

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
	response1, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/StudentTotalMarks.asp", requestOption1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()
	if err := checkResponse(s, response1); err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней отчет об успеваемости ученика.
	return inner.TotalMarkReportParser(bytes.NewReader(response1.Bytes()))
}

/*
02 тип.
*/

// GetAverageMarkReport возвращает средние баллы ученика с сервера первого типа.
func GetAverageMarkReport(s *ss.Session, dateBegin, dateEnd, Type, studentID string) (*dt.AverageMarkReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
		Data: map[string]string{
			"AT":        s.AT,
			"LoginType": "0",
			"RPTID":     "1",
			"ThmID":     "1",
			"VER":       s.VER,
		},
		Headers: map[string]string{
			"Origin":                    p + s.Serv.Link,
			"Upgrade-Insecure-Requests": "1",
			"Referer":                   p + s.Serv.Link + "/asp/Reports/Reports.asp",
		},
	}
	response0, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentAverageMark.asp", requestOptions0)
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
			"SID":       studentID,
			"TA":        "",
			"ThmID":     "1",
			"VER":       s.VER,
		},
		Headers: map[string]string{
			"Origin":           p + s.Serv.Link,
			"X-Requested-With": "XMLHttpRequest",
			"at":               s.AT,
			"Referer":          p + s.Serv.Link + "/asp/Reports/ReportStudentAverageMark.asp",
		},
	}
	response1, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/StudentAverageMark.asp", requestOptions1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()
	if err := checkResponse(s, response1); err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней отчет о средних баллах ученика.
	return inner.AverageMarkReportParser(bytes.NewReader(response1.Bytes()))
}

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
			"Origin":           p + s.Serv.Link,
			"X-Requested-With": "XMLHttpRequest",
			"at":               s.AT,
			"Referer":          p + s.Serv.Link + "/asp/Reports/ReportStudentAverageMarkDyn.asp",
		},
	}
	response1, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/StudentAverageMarkDyn.asp", requestOptions1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()
	if err := checkResponse(s, response1); err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней отчет о динамике среднего балла.
	return inner.AverageMarkDynReportParser(bytes.NewReader(response1.Bytes()))
}

/*
04 тип.
*/

// GetStudentGradeReport возвращает отчет об успеваемости ученика по предмету с сервера первого типа.
func GetStudentGradeReport(s *ss.Session, dateBegin, dateEnd, subjectID, studentID string) (*dt.StudentGradeReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
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
	response0, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentGrades.asp", requestOptions0)
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
	response1, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentGrades.asp", requestOptions1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()
	if err := checkResponse(s, response1); err != nil {
		return nil, err
	}

	// 2-ой POST-запрос.
	requestOptions2 := &gr.RequestOptions{
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
	response2, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/StudentGrades.asp", requestOptions2)
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
	// находящуюся в теле ответа, и найти в ней отчет об успеваемости ученика по предмету.
	return inner.StudentGradeReportParser(bytes.NewReader(response2.Bytes()))
}

/*
05 тип.
*/

// GetStudentTotalReport возвращает отчет о посещениях ученика с сервера первого типа.
func GetStudentTotalReport(s *ss.Session, dateBegin, dateEnd, studentID string) (*dt.StudentTotalReport, error) {
	p := "http://"

	// 0-ой Post-запрос.
	requestOptions0 := &gr.RequestOptions{
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
	response0, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentTotal.asp", requestOptions0)
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
	response1, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/StudentTotal.asp", requestOptions1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()
	if err := checkResponse(s, response1); err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней отчет о посещених ученика.
	return inner.StudentTotalReportParser(bytes.NewReader(response1.Bytes()))
}

/*
06 тип.
*/

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// GetStudentAttendanceGradeReport - отчет шестого типа пока что пропускаем.
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

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
	response0, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ReportStudentTotal.asp", requestOptions0)
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
			"AT":        s.AT,
			"BACK":      "/asp/Reports/ReportJournalAccess.asp",
			"LoginType": "0",
			"NA":        "",
			"PCLID":     "",
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
	response1, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/JournalAccess.asp", requestOptions1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()
	if err := checkResponse(s, response1); err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа, и найти в ней отчет о доступе к журналу.
	return inner.JournalAccessReportParser(bytes.NewReader(response1.Bytes()))
}

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
			"ReportType": "1",
			"SID":        studentID,
			"TA":         "",
			"TERMID":     "10067",
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
	response1, err := s.Sess.Post(p+s.Serv.Link+"/asp/Reports/ParentInfoLetter.asp", requestOptions1)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response1.Close()
	}()
	if err := checkResponse(s, response1); err != nil {
		return nil, err
	}

	// Если мы дошли до этого места, то можно распарсить HTML-страницу,
	// находящуюся в теле ответа и найти в ней шаблон письма родителям.
	return inner.ParentInfoLetterReportParser(bytes.NewReader(response1.Bytes()))
}
