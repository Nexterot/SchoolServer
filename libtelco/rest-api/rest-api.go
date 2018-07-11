// Copyright (C) 2018 Barluka Alexander

/*
Package restapi содержит handler'ы для взаимодействия сервера с клиентами.
*/
package restapi

import (
	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"
)

// RestAPI struct содержит конфигурацию Rest API.
type RestAPI struct {
	config *cp.Config
	store  *sessions.CookieStore
	logger *log.Logger
}

// NewRestAPI создает структуру для работы с Rest API.
func NewRestAPI(logger *log.Logger, config *cp.Config) *RestAPI {
	key := make([]byte, 32)
	rand.Read(key)
	logger.Info("Generated secure key: ", key)
	return &RestAPI{
		config: config,
		store:  sessions.NewCookieStore(key),
		logger: logger,
	}
}

// BindHandlers привязывает все handler'ы (с помощью http.HandleFunc).
func (rest *RestAPI) BindHandlers() {
	http.HandleFunc("/", rest.ErrorHandler)

	http.HandleFunc("/get_school_list", rest.GetSchoolListHandler) // done
	http.HandleFunc("/check_permission", rest.Handler)
	http.HandleFunc("/sign_in", rest.SignInHandler) // in progress
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

// school используется в GetSchoolListHandler
type school struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

// schoolListResponse используется в GetSchoolListHandler
type schoolListResponse struct {
	Schools []school `json:"schools"`
}

// GetSchoolListHandler обрабатывает запрос на получение списка обслуживаемых школ.
func (rest *RestAPI) GetSchoolListHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("GetSchoolListHandler called")
	if req.Method != "GET" {
		rest.logger.Error("Wrong method: ", req.Method)
		return
	}
	schoolList := make([]school, len(rest.config.Schools))
	for id, sch := range rest.config.Schools {
		schoolList[id] = school{sch.Name, strconv.Itoa(id)}
	}
	resp := schoolListResponse{schoolList}
	bytes, err := json.Marshal(resp)
	if err != nil {
		rest.logger.Error("Error marshalling list of schools. May be critical!")
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent list of schools: ", resp)
}

// SignInRequest используется в SignInHandler
type SignInRequest struct {
	Login   string `json:"login"`
	Passkey string `json:"passkey"`
	Id      string `json:"id"`
}

// SignInHandler обрабатывает вход в учетную запись на сайте школы
func (rest *RestAPI) SignInHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("SignInHandler called (in development)")
	if req.Method != "POST" {
		rest.logger.Error("Wrong method: ", req.Method)
		return
	}
	var rReq SignInRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		// Тут надо ошибку отправлять клиенту
		rest.logger.Error("Malformed request data")
		return
	}
	rest.logger.Info("Valid data:", rReq)

}

// Handler временный абстрактный handler для некоторых еще не реализованных
// обработчиков запросов.
func (rest *RestAPI) Handler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("Handler called (not implemented yet)", req.URL.EscapedPath())
}
