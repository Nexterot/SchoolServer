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
	//"SchoolServer/libtelco/db

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
	//db		*db.Database
}

// NewRestAPI создает структуру для работы с Rest API.
func NewRestAPI(logger *log.Logger, config *cp.Config) *RestAPI {
	key := make([]byte, 32)
	rand.Read(key)
	logger.Info("Generated secure key: ", key)
	newStore := sessions.NewCookieStore(key)
	newStore.MaxAge(86400 * 365)
	return &RestAPI{
		config:      config,
		store:       newStore,
		logger:      logger,
		sessionsMap: make(map[string]*ss.Session),
		//db	   : db.NewDatabase(),
	}
}

// BindHandlers привязывает все handler'ы (с помощью http.HandleFunc).
func (rest *RestAPI) BindHandlers() {
	http.HandleFunc("/", rest.ErrorHandler)

	http.HandleFunc("/get_school_list", rest.GetSchoolListHandler)    // done
	http.HandleFunc("/check_permission", rest.CheckPermissionHandler) // done
	http.HandleFunc("/sign_in", rest.SignInHandler)                   // done
	http.HandleFunc("/log_out", rest.LogOutHandler)                   // done

	http.HandleFunc("/get_tasks_and_marks", rest.GetTasksAndMarksHandler) // done
	http.HandleFunc("/get_lesson_description", rest.Handler)
	http.HandleFunc("/mark_as_done", rest.Handler)
	http.HandleFunc("/unmark_as_done", rest.Handler)

	http.HandleFunc("/get_posts", rest.Handler)

	http.HandleFunc("/get_schedule", rest.GetScheduleHandler) // done

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

// checkPermissionRequest используется в CheckPermissionHandler
type checkPermissionRequest struct {
	Login string `json:"login"`
	Id    string `json:"id"`
}

// checkPermissionResponse используется в CheckPermissionHandler
type checkPermissionResponse struct {
	Permission string `json:"permission"`
}

// CheckPermissionHandler проверяет, есть ли разрешение на работу с школой
func (rest *RestAPI) CheckPermissionHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("CheckPermissionHandler called")
	if req.Method != "POST" {
		rest.logger.Error("Wrong method: ", req.Method)
		return
	}
	// Чтение json'a
	var rReq checkPermissionRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		respwr.WriteHeader(http.StatusBadRequest)
		rest.logger.Error("Malformed request data")
		return
	}
	// TODO лазать в БД за соответствующим полем
	resp := checkPermissionResponse{"true"}
	bytes, err := json.Marshal(resp)
	if err != nil {
		rest.logger.Error("Error marshalling permission check response")
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent permission check response: ", resp)
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

// tasksMarksRequest используется в GetTasksAndMarksHandler
type tasksMarksRequest struct {
	Week string `json:"week"`
	Id   string `json:"id"`
}

// GetTasksAndMarksHandler возвращает задания и оценки на неделю
func (rest *RestAPI) GetTasksAndMarksHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("GetTasksAndMarksHandler called")
	if req.Method != "POST" {
		rest.logger.Error("Wrong method: ", req.Method)
		return
	}
	// Прочитать куку
	cookie, err := req.Cookie("sessionName")
	if err != nil {
		rest.logger.Info("User not authorized: sessionName absent")
		respwr.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionName := cookie.Value
	// Получить существующий объект сессии
	session, err := rest.store.Get(req, sessionName)
	if session.IsNew {
		rest.logger.Error("Local session broken")
		delete(rest.sessionsMap, sessionName)
		session.Options.MaxAge = -1
		session.Save(req, respwr)
		respwr.WriteHeader(http.StatusBadRequest)
		return
	}
	// Чтение запроса от клиента
	var rReq tasksMarksRequest
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&rReq)
	if err != nil {
		respwr.WriteHeader(http.StatusBadRequest)
		rest.logger.Error("Malformed request data")
		return
	}
	// Если нет удаленной сессии, создать
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		rest.logger.Info("No remote session, creating one (not implemented yet)")
		// сходить в БД за логином и паролем, создать новую сессию и войти
		// userName := session.Values["userName"]
		// school, err := db.GetAuthData(userName)
		// if err != nil {
		// TODO попросить пользователя войти
		// rest.logger.Error("Error reading database")
		// return
		// }
		// remoteSession = ss.NewSession(school)
		// if err = remoteSession.Login(); err != nil {
		// rest.logger.Error("Error remote signing in")
		// return
		// }
	}
	// TODO Если удаленная сессия есть, но не залогинена, снова войти
	week := rReq.Week
	if week == "" {
		week = time.Now().Format("02.01.2006")
	}
	weekMarks, err := remoteSession.GetWeekSchoolMarks(week)
	if err != nil {
		rest.logger.Info("Unable to get week tasks and marks: ", err)
		// TODO Добавить повторную авторизацию для удаленной сессии
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	bytes, err := json.Marshal(weekMarks)
	if err != nil {
		rest.logger.Error("Error marshalling weekMarks")
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent tasks and marks for a week: ", weekMarks)
}

// scheduleRequest используется в GetScheduleHandler
type scheduleRequest struct {
	Days string `json:"days"`
	Id   string `json:"id"`
}

// GetScheduleHandler возвращает расписание на неделю
func (rest *RestAPI) GetScheduleHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("GetScheduleHandler called")
	if req.Method != "POST" {
		rest.logger.Error("Wrong method: ", req.Method)
		return
	}
	// Прочитать куку
	cookie, err := req.Cookie("sessionName")
	if err != nil {
		rest.logger.Info("User not authorized: sessionName absent")
		respwr.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionName := cookie.Value
	// Получить существующий объект сессии
	session, err := rest.store.Get(req, sessionName)
	if session.IsNew {
		rest.logger.Error("Local session broken")
		delete(rest.sessionsMap, sessionName)
		session.Options.MaxAge = -1
		session.Save(req, respwr)
		respwr.WriteHeader(http.StatusBadRequest)
		return
	}
	// Чтение запроса от клиента
	var rReq scheduleRequest
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&rReq)
	if err != nil {
		respwr.WriteHeader(http.StatusBadRequest)
		rest.logger.Error("Malformed request data")
		return
	}
	// Если нет удаленной сессии, создать
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		rest.logger.Info("No remote session, creating one (not implemented yet)")
		// сходить в БД за логином и паролем, создать новую сессию и войти
		// userName := session.Values["userName"]
		// school, err := db.GetAuthData(userName)
		// if err != nil {
		// TODO попросить пользователя войти
		// rest.logger.Error("Error reading database")
		// return
		// }
		// remoteSession = ss.NewSession(school)
		// if err = remoteSession.Login(); err != nil {
		// rest.logger.Error("Error remote signing in")
		// return
		// }
	}
	// TODO Если удаленная сессия есть, но не залогинена, снова войти
	days, err := strconv.Atoi(rReq.Days)
	if err != nil {
		rest.logger.Error("Invalid param days specified: ", err)
		respwr.WriteHeader(http.StatusBadRequest)
		return
	}

	today := time.Now().Format("02.01.2006")

	timeTable, err := remoteSession.GetTimeTable(today, days)
	if err != nil {
		// TODO Добавить повторную авторизацию для удаленной сессии
		rest.logger.Error("Unable to get schedule: ", err)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	bytes, err := json.Marshal(timeTable)
	if err != nil {
		rest.logger.Error("Error marshalling timeTable")
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent schedule for a week: ", timeTable)
}

// LogOutHandler обрабатывает удаление сессии клиента и отвязку устройства
func (rest *RestAPI) LogOutHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("LogOutHandler called")
	if req.Method != "GET" {
		rest.logger.Error("Wrong method: ", req.Method)
		return
	}
	// Прочитать куку
	cookie, err := req.Cookie("sessionName")
	if err != nil {
		rest.logger.Info("User not authorized: sessionName absent")
		respwr.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionName := cookie.Value

	/* TODO
	не чистить sessionMap, сохранять в БД идентификатор последней сессии,
	чтобы можно было к случае logout+login восстановить удаленную сессию
	по имени пользователя
	*/

	// Если кука есть, удалить локальную и удаленную сессии
	session, err := rest.store.Get(req, sessionName)
	if err != nil {
		rest.logger.Info("Error getting session: ", sessionName)
		return
	}
	delete(rest.sessionsMap, sessionName)
	session.Options.MaxAge = -1
	session.Save(req, respwr)
	respwr.WriteHeader(http.StatusOK)
	rest.logger.Info("Successful logout for session ", sessionName)
}

// signInRequest используется в SignInHandler
type signInRequest struct {
	Login   string `json:"login"`
	Passkey string `json:"passkey"`
	Id      string `json:"id"`
}

// SignInHandler обрабатывает вход в учетную запись на сайте школы
func (rest *RestAPI) SignInHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("SignInHandler called")
	if req.Method != "POST" {
		rest.logger.Error("Wrong method: ", req.Method)
		return
	}
	// TODO вызывать checkPermision
	// Чтение запроса от клиента
	var rReq signInRequest
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
	newSessionName := hex.EncodeToString(hasher.Sum(nil))
	newLocalSession, err := rest.store.Get(req, newSessionName)
	if err != nil {
		rest.logger.Error("Error creating new local session")
		return
	}
	rest.sessionsMap[newSessionName] = newRemoteSession
	newLocalSession.Values["userName"] = rReq.Login
	newLocalSession.Save(req, respwr)
	// Устанавливаем в куки значение sessionName
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{
		Name: "sessionName", Value: newSessionName, Expires: expiration,
	}
	http.SetCookie(respwr, &cookie)
	rest.logger.Info("Successfully signed in as user: ", rReq.Login)
}

// Handler временный абстрактный handler для некоторых еще не реализованных
// обработчиков запросов.
func (rest *RestAPI) Handler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("Handler called (not implemented yet)", req.URL.EscapedPath())
}
