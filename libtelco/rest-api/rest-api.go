// Copyright (C) 2018 Barluka Alexander

/*
Package restapi содержит handler'ы для взаимодействия сервера с клиентами.
*/
package restapi

import (
	"SchoolServer/libtelco/log"
	"math/rand"
	"net/http"

	"github.com/gorilla/sessions"
)

// RestAPI struct содержит конфигурацию Rest API.
type RestAPI struct {
	store  *sessions.CookieStore
	logger *log.Logger
}

// NewRestAPI создает структуру для работы с Rest API.
func NewRestAPI(logger *log.Logger) *RestAPI {
	key := make([]byte, 32)
	rand.Read(key)
	logger.Info("Generated secure key: ", key)
	return &RestAPI{
		store:  sessions.NewCookieStore(key),
		logger: logger,
	}
}

// BindHandlers привязывает все handler'ы (с помощью http.HandleFunc).
func (rest *RestAPI) BindHandlers() {
	http.HandleFunc("/", rest.ErrorHandler)

	http.HandleFunc("/get_school_list", rest.GetSchoolListHandler)
	http.HandleFunc("/check_permission", rest.Handler)
	http.HandleFunc("/sign_in", rest.Handler)
	http.HandleFunc("/log_out", rest.Handler)

	http.HandleFunc("/get_tasks_and_marks", rest.Handler)
	http.HandleFunc("/get_lesson_description", rest.Handler)
	http.HandleFunc("/mark_as_done", rest.Handler)
	http.HandleFunc("/unmark_as_done", rest.Handler)

	http.HandleFunc("/get_posts", rest.Handler)

	http.HandleFunc("/get_schedule", rest.Handler)

	http.HandleFunc("/get_report_student_total_mark", rest.Handler)
	http.HandleFunc("/get_report_student_average_mark", rest.Handler)
	http.HandleFunc("/get_report_student_average_mark_dyn", rest.Handler)
	http.HandleFunc("/get_report_student_grades_lesson_list", rest.Handler)
	http.HandleFunc("/get_report_student_grades", rest.Handler)
	http.HandleFunc("/get_report_student_total", rest.Handler)
	http.HandleFunc("/get_report_journal_access_classes_list", rest.Handler)
	http.HandleFunc("/get_report_journal_access", rest.Handler)
	http.HandleFunc("/get_report_parent_info_letter_data", rest.Handler)
	http.HandleFunc("/get_report_parent_info_letter", rest.Handler)

	http.HandleFunc("/get_resources", rest.Handler)

	http.HandleFunc("/get_mail", rest.Handler)
	http.HandleFunc("/get_mail_description", rest.Handler)
	http.HandleFunc("/delete_mail", rest.Handler)
	http.HandleFunc("/send_letter", rest.Handler)
	http.HandleFunc("/get_address_book", rest.Handler)

	http.HandleFunc("/get_forum", rest.Handler)
	http.HandleFunc("/get_forum_messages", rest.Handler)
	http.HandleFunc("/create_topic", rest.Handler)
	http.HandleFunc("/create_message_in_topic", rest.Handler)

	http.HandleFunc("/change_password", rest.Handler)
}

// ErrorHandler обрабатывает некорректные запросы.
func (rest *RestAPI) ErrorHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("Wrong request:", req.URL.EscapedPath())
}

// GetSchoolListHandler обрабатывает запрос на получение списка обслуживаемых школ.
func (rest *RestAPI) GetSchoolListHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("GetSchoolListHandler called (not implemented yet)")
	if req.Method != "GET" {
		rest.logger.Error("Wrong method: ", req.Method)
		return
	}
}

// Handler временный абстрактный handler для некоторых еще не реализованных
// обработчиков запросов.
func (rest *RestAPI) Handler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("Handler called (not implemented yet)", req.URL.EscapedPath())
}
