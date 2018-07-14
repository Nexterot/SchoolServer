// Copyright (C) 2018 Barluka Alexander

/*
Package restapi содержит handler'ы для взаимодействия сервера с клиентами.
*/
package restapi

import (
	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"
	ss "SchoolServer/libtelco/sessions"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
)

// RestAPI struct содержит конфигурацию Rest API.
// sessionsMap содержит отображения идентификаторов сессий Rest API
// в объекты сессий на удаленном сервере.
type RestAPI struct {
	config      *cp.Config
	store       *sessions.CookieStore
	logger      *log.Logger
	sessionsMap map[string]*ss.Session
}

// NewRestAPI создает структуру для работы с Rest API.
func NewRestAPI(logger *log.Logger, config *cp.Config) *RestAPI {
	key := make([]byte, 32)
	rand.Read(key)
	logger.Info("Generated secure key: ", key)
	return &RestAPI{
		config:      config,
		store:       sessions.NewCookieStore(key),
		logger:      logger,
		sessionsMap: make(map[string]*ss.Session),
	}
}

// BindHandlers привязывает все handler'ы (с помощью http.HandleFunc).
func (rest *RestAPI) BindHandlers() {
	http.HandleFunc("/", rest.ErrorHandler)

	http.HandleFunc("/get_school_list", rest.GetSchoolListHandler) // done
	http.HandleFunc("/check_permission", rest.Handler)
	http.HandleFunc("/sign_in", rest.SignInHandler) // done
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
		rest.logger.Error("Error marshalling list of schools")
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

// SessionResponse используется в SignInHandler
type SessionResponse struct {
	SessionID string `json:"session_id"`
}

// SignInHandler обрабатывает вход в учетную запись на сайте школы
func (rest *RestAPI) SignInHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("SignInHandler called")
	if req.Method != "POST" {
		rest.logger.Error("Wrong method: ", req.Method)
		return
	}
	// Чтение запроса от клиента
	var rReq SignInRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		respwr.WriteHeader(http.StatusBadRequest)
		rest.logger.Error("Malformed request data")
		return
	}
	schoolId, _ := strconv.Atoi(rReq.Id)
	if schoolId >= len(rest.config.Schools) {
		respwr.WriteHeader(http.StatusBadRequest)
		rest.logger.Error("No school with such id: ", rReq.Id)
		return
	}
	rest.logger.Info("Valid data:", rReq)
	school := rest.config.Schools[schoolId]
	school.Login = rReq.Login
	school.Password = rReq.Passkey
	// Создание удаленной сессии
	newRemoteSession := ss.NewSession(&school)
	if err = newRemoteSession.Login(); err != nil {
		rest.logger.Error("Error remote signing in")
		respwr.WriteHeader(http.StatusBadRequest)
		return
	}
	// Если удаленная авторизация прошла успешно, создать локальную сессию
	// и привязать к ней удаленную сессию
	timeString := time.Now().String()
	hasher := md5.New()
	if _, err = hasher.Write([]byte(timeString)); err != nil {
		rest.logger.Error("Md5 hashing error: ", err)
		return
	}
	newSessionId := hex.EncodeToString(hasher.Sum(nil))
	newLocalSession, err := rest.store.Get(req, newSessionId)
	if err != nil {
		rest.logger.Error("Error creating new local session")
		return
	}
	rest.sessionsMap[newSessionId] = newRemoteSession
	newLocalSession.Save(req, respwr)
	// Устанавливаем в куки значение sessionId
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{
		Name: "sessionId", Value: newSessionId, Expires: expiration,
	}
	http.SetCookie(respwr, &cookie)
	rest.logger.Info("Successfully signed in as user: ", rReq.Login)
}

// Handler временный абстрактный handler для некоторых еще не реализованных
// обработчиков запросов.
func (rest *RestAPI) Handler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("Handler called (not implemented yet)", req.URL.EscapedPath())
}
