// rest-api.go

/*
Package restapi содержит handler'ы для взаимодействия сервера с клиентами.
*/
package restapi

import (
	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"
	ss "SchoolServer/libtelco/sessions"
	db "SchoolServer/libtelco/sql-db"
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
	db          *db.Database
}

// NewRestAPI создает структуру для работы с Rest API.
func NewRestAPI(logger *log.Logger, config *cp.Config) *RestAPI {
	key := make([]byte, 32)
	rand.Read(key)
	logger.Info("Generated secure key: ", key)
	newStore := sessions.NewCookieStore(key)
	newStore.MaxAge(86400 * 365)
	database, err := db.NewDatabase(logger, config)
	if err != nil {
		logger.Error("Error when creating database!", err)
	}
	return &RestAPI{
		config:      config,
		store:       newStore,
		logger:      logger,
		sessionsMap: make(map[string]*ss.Session),
		db:          database,
	}
}

// BindHandlers привязывает все handler'ы Rest API
func (rest *RestAPI) BindHandlers() {
	http.HandleFunc("/", rest.ErrorHandler)

	// Общее: Запрос списка школ, запрос доступа, авторизация, выход
	http.HandleFunc("/get_school_list", rest.GetSchoolListHandler)    // done
	http.HandleFunc("/check_permission", rest.CheckPermissionHandler) // done
	http.HandleFunc("/sign_in", rest.SignInHandler)                   // done
	http.HandleFunc("/log_out", rest.LogOutHandler)                   // done
	// Дневник: список учеников, задания и оценки на неделю, отметить задание
	// как выполненное/невыполненное
	http.HandleFunc("/get_children_map", rest.GetChildrenMapHandler)             // done
	http.HandleFunc("/get_tasks_and_marks", rest.GetTasksAndMarksHandler)        // done
	http.HandleFunc("/get_lesson_description", rest.GetLessonDescriptionHandler) // done
	http.HandleFunc("/mark_as_done", rest.MarkAsDoneHandler)                     // done
	http.HandleFunc("/unmark_as_done", rest.UnmarkAsDoneHandler)                 // done
	// Объявления: получение списка объявлений
	http.HandleFunc("/get_posts", rest.Handler)
	// Расписание: получение расписания на N дней
	http.HandleFunc("/get_schedule", rest.GetScheduleHandler) // done
	// Отчеты
	http.HandleFunc("/get_report_student_total_marks", rest.GetReportStudentTotalMarksHandler)              // done
	http.HandleFunc("/get_report_student_average_mark", rest.GetReportStudentAverageMarkHandler)            // done
	http.HandleFunc("/get_report_student_average_mark_dyn", rest.GetReportStudentAverageMarkDynHandler)     // done
	http.HandleFunc("/get_report_student_grades_lesson_list", rest.GetReportStudentGradesLessonListHandler) // done
	http.HandleFunc("/get_report_student_grades", rest.GetReportStudentGradesHandler)                       // done
	http.HandleFunc("/get_report_student_total", rest.GetReportStudentTotalHandler)                         // done
	http.HandleFunc("/get_report_journal_access", rest.GetReportJournalAccessHandler)                       // done
	http.HandleFunc("/get_report_parent_info_letter_data", rest.Handler)
	http.HandleFunc("/get_report_parent_info_letter", rest.GetReportParentInfoLetterHandler) // done
	// Школьные ресурсы
	http.HandleFunc("/get_resources", rest.Handler)
	// Почта
	http.HandleFunc("/get_mail", rest.Handler)
	http.HandleFunc("/get_mail_description", rest.Handler)
	http.HandleFunc("/delete_mail", rest.Handler)
	http.HandleFunc("/send_letter", rest.Handler)
	http.HandleFunc("/get_address_book", rest.Handler)
	// Форум
	http.HandleFunc("/get_forum", rest.Handler)
	http.HandleFunc("/get_forum_messages", rest.Handler)
	http.HandleFunc("/create_topic", rest.Handler)
	http.HandleFunc("/create_message_in_topic", rest.Handler)
	// Настройки
	http.HandleFunc("/change_password", rest.Handler)
}

// checkPermissionRequest используется в CheckPermissionHandler
type checkPermissionRequest struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
}

// checkPermissionResponse используется в CheckPermissionHandler
type checkPermissionResponse struct {
	Permission bool `json:"permission"`
}

// CheckPermissionHandler проверяет, есть ли разрешение на работу с школой
func (rest *RestAPI) CheckPermissionHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("CheckPermissionHandler called")
	if req.Method != "POST" {
		rest.logger.Error("Wrong method: ", req.Method)
		respwr.WriteHeader(http.StatusBadRequest)
		return
	}
	// Чтение json'a
	var rReq checkPermissionRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Error("Malformed request data")
		respwr.WriteHeader(http.StatusBadRequest)
		return
	}
	// Проверим разрешение у школы
	perm, err := rest.db.GetSchoolPermission(rReq.ID)
	if err != nil {
		rest.logger.Error("Invalid id param specified")
		respwr.WriteHeader(http.StatusBadRequest)
		return
	}
	if !perm {
		// Если у школы нет разрешения, проверить разрешение пользователя
		userPerm, err := rest.db.GetUserPermission(rReq.Login, rReq.ID)
		if err != nil {
			if err.Error() == "record not found" {
				// Пользователь новый, вернем true
				perm = true
			} else {
				rest.logger.Error("Getting permission from db: ", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		perm = userPerm
	}
	resp := checkPermissionResponse{perm}
	bytes, err := json.Marshal(resp)
	if err != nil {
		rest.logger.Error("Error marshalling permission check response")
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent permission check response: ", resp)
}

// ErrorHandler обрабатывает некорректные запросы.
func (rest *RestAPI) ErrorHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("Wrong request:", req.URL.EscapedPath())
}

// getReportStudentTotalMarksRequest используется в GetReportStudentTotalMarksHandler
type getReportStudentTotalMarksRequest struct {
	ID int `json:"id"`
}

// GetReportStudentTotalMarksHandler обрабатывает запрос на получение отчета
// об итоговых оценках
func (rest *RestAPI) GetReportStudentTotalMarksHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("GetReportStudentTotalMarksHandler called")
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
	var rReq getReportStudentTotalMarksRequest
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
		rest.logger.Info("No remote session, creating new one")
		userName := session.Values["userName"]
		schoolID := session.Values["schoolID"]
		school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
		if err != nil {
			rest.logger.Error("Error reading database")
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		remoteSession = ss.NewSession(school)
		if err = remoteSession.Login(); err != nil {
			rest.logger.Error("Error remote signing in")
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		rest.sessionsMap[sessionName] = remoteSession
	}
	id := strconv.Itoa(rReq.ID)
	totalMarkReport, err := remoteSession.GetTotalMarkReport(id)
	// Если удаленная сессия есть в mapSessions, но не активна, создать новую
	if err != nil {
		if err.Error() == "You was logged out from server" {
			rest.logger.Info("Remote connection broken, creation new one")
			userName := session.Values["userName"]
			schoolID := session.Values["schoolID"]
			school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
			if err != nil {
				rest.logger.Error("Error reading database")
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			remoteSession = ss.NewSession(school)
			if err = remoteSession.Login(); err != nil {
				rest.logger.Error("Error remote signing in")
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			rest.sessionsMap[sessionName] = remoteSession
			rest.logger.Info("Successfully created new remote session")
		} else {
			rest.logger.Error("Unable to get total marks: ", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	bytes, err := json.Marshal(totalMarkReport)
	if err != nil {
		rest.logger.Error("Error marshalling totalMarkReport")
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent total marks report: ", totalMarkReport)
}

// getReportStudentAverageMarkRequest используется в GetReportStudentAverageMarkHandler
// и GetReportStudentAverageMarkDynHandler
type getReportStudentAverageMarkRequest struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
	From string `json:"from"`
	To   string `json:"to"`
}

// GetReportStudentAverageMarkHandler обрабатывает запрос на получение отчета
// о среднем балле
func (rest *RestAPI) GetReportStudentAverageMarkHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("GetReportStudentAverageMarkHandler called")
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
	var rReq getReportStudentAverageMarkRequest
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&rReq)
	if err != nil {
		respwr.WriteHeader(http.StatusBadRequest)
		rest.logger.Error("Malformed request data")
		return
	}
	id := strconv.Itoa(rReq.ID)
	// Если нет удаленной сессии, создать
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		rest.logger.Info("No remote session, creating new one")
		userName := session.Values["userName"]
		schoolID := session.Values["schoolID"]
		school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
		if err != nil {
			rest.logger.Error("Error reading database", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		remoteSession = ss.NewSession(school)
		if err = remoteSession.Login(); err != nil {
			rest.logger.Error("Error remote signing in", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		rest.sessionsMap[sessionName] = remoteSession
	}

	averageMarkReport, err := remoteSession.GetAverageMarkReport(rReq.From, rReq.To, rReq.Type, id)
	// Если удаленная сессия есть в mapSessions, но не активна, создать новую
	if err != nil {
		if err.Error() == "You was logged out from server" {
			rest.logger.Info("Remote connection broken, creation new one")
			userName := session.Values["userName"]
			schoolID := session.Values["schoolID"]
			school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
			if err != nil {
				rest.logger.Error("Error reading database", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			remoteSession = ss.NewSession(school)
			if err = remoteSession.Login(); err != nil {
				rest.logger.Error("Error remote signing in", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			rest.sessionsMap[sessionName] = remoteSession
			rest.logger.Info("Successfully created new remote session")
		} else {
			rest.logger.Error("Unable to get average marks: ", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	bytes, err := json.Marshal(averageMarkReport)
	if err != nil {
		rest.logger.Error("Error marshalling averageMarkReport")
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent average marks report: ", averageMarkReport)
}

// GetReportStudentAverageMarkDynHandler обрабатывает запрос на получение отчета
// о динамике среднего балла
func (rest *RestAPI) GetReportStudentAverageMarkDynHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("GetReportStudentAverageMarkDynHandler called")
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
	var rReq getReportStudentAverageMarkRequest
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&rReq)
	if err != nil {
		respwr.WriteHeader(http.StatusBadRequest)
		rest.logger.Error("Malformed request data")
		return
	}
	id := strconv.Itoa(rReq.ID)
	// Если нет удаленной сессии, создать
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		rest.logger.Info("No remote session, creating new one")
		userName := session.Values["userName"]
		schoolID := session.Values["schoolID"]
		school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
		if err != nil {
			rest.logger.Error("Error reading database", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		remoteSession = ss.NewSession(school)
		if err = remoteSession.Login(); err != nil {
			rest.logger.Error("Error remote signing in", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		rest.sessionsMap[sessionName] = remoteSession
	}
	averageMarkDynReport, err := remoteSession.GetAverageMarkDynReport(rReq.From, rReq.To, rReq.Type, id)
	// Если удаленная сессия есть в mapSessions, но не активна, создать новую
	if err != nil {
		if err.Error() == "You was logged out from server" {
			rest.logger.Info("Remote connection broken, creation new one")
			userName := session.Values["userName"]
			schoolID := session.Values["schoolID"]
			school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
			if err != nil {
				rest.logger.Error("Error reading database", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			remoteSession = ss.NewSession(school)
			if err = remoteSession.Login(); err != nil {
				rest.logger.Error("Error remote signing in", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			rest.sessionsMap[sessionName] = remoteSession
			rest.logger.Info("Successfully created new remote session")
		} else {
			rest.logger.Error("Unable to get average dyn marks: ", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	bytes, err := json.Marshal(averageMarkDynReport)
	if err != nil {
		rest.logger.Error("Error marshalling averageMarkDynReport")
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent average marks dynamic report: ", averageMarkDynReport)
}

// getReportStudentGradesLessonListRequest используется в GetReportStudentGradesLessonListHandler
type getReportStudentGradesLessonListRequest struct {
	ID int `json:"id"`
}

// GetReportStudentGradesLessonListHandler обрабатывает запрос на получение
// списка предметов для отчета 'Об успеваемости'
func (rest *RestAPI) GetReportStudentGradesLessonListHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("GetReportStudentGradesLessonListHandler called")
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
	var rReq getReportStudentGradesLessonListRequest
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&rReq)
	if err != nil {
		respwr.WriteHeader(http.StatusBadRequest)
		rest.logger.Error("Malformed request data")
		return
	}
	studentID := strconv.Itoa(rReq.ID)
	// Если нет удаленной сессии, создать
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		rest.logger.Info("No remote session, creating new one")
		userName := session.Values["userName"]
		schoolID := session.Values["schoolID"]
		school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
		if err != nil {
			rest.logger.Error("Error reading database", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		remoteSession = ss.NewSession(school)
		if err = remoteSession.Login(); err != nil {
			rest.logger.Error("Error remote signing in", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		rest.sessionsMap[sessionName] = remoteSession
	}
	lessonsMap, err := remoteSession.GetLessonsMap(studentID)
	// Если удаленная сессия есть в mapSessions, но не активна, создать новую
	if err != nil {
		if err.Error() == "You was logged out from server" {
			rest.logger.Info("Remote connection broken, creation new one")
			userName := session.Values["userName"]
			schoolID := session.Values["schoolID"]
			school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
			if err != nil {
				rest.logger.Error("Error reading database", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			remoteSession = ss.NewSession(school)
			if err = remoteSession.Login(); err != nil {
				rest.logger.Error("Error remote signing in", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			rest.sessionsMap[sessionName] = remoteSession
			rest.logger.Info("Successfully created new remote session")
		} else {
			rest.logger.Error("Unable to get lessons map: ", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	bytes, err := json.Marshal(lessonsMap)
	if err != nil {
		rest.logger.Error("Error marshalling student lessons map")
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent lessons map: ", lessonsMap)
}

// getReportStudentGradesRequest используется в GetReportStudentGradesHandler
type getReportStudentGradesRequest struct {
	StudentID int    `json:"student_id"`
	LessonID  int    `json:"lesson_id"`
	From      string `json:"from"`
	To        string `json:"to"`
}

// GetReportStudentGradesHandler обрабатывает запрос на получение
// отчета 'Об успеваемости'
func (rest *RestAPI) GetReportStudentGradesHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("GetReportStudentGradesHandler called")
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
	var rReq getReportStudentGradesRequest
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&rReq)
	if err != nil {
		respwr.WriteHeader(http.StatusBadRequest)
		rest.logger.Error("Malformed request data")
		return
	}
	studentID := strconv.Itoa(rReq.StudentID)
	lessonID := strconv.Itoa(rReq.LessonID)
	// Если нет удаленной сессии, создать
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		rest.logger.Info("No remote session, creating new one")
		userName := session.Values["userName"]
		schoolID := session.Values["schoolID"]
		school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
		if err != nil {
			rest.logger.Error("Error reading database", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		remoteSession = ss.NewSession(school)
		if err = remoteSession.Login(); err != nil {
			rest.logger.Error("Error remote signing in", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		rest.sessionsMap[sessionName] = remoteSession
	}
	gradeReport, err := remoteSession.GetStudentGradeReport(rReq.From, rReq.To, lessonID, studentID)
	// Если удаленная сессия есть в mapSessions, но не активна, создать новую
	if err != nil {
		if err.Error() == "You was logged out from server" {
			rest.logger.Info("Remote connection broken, creation new one")
			userName := session.Values["userName"]
			schoolID := session.Values["schoolID"]
			school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
			if err != nil {
				rest.logger.Error("Error reading database", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			remoteSession = ss.NewSession(school)
			if err = remoteSession.Login(); err != nil {
				rest.logger.Error("Error remote signing in", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			rest.sessionsMap[sessionName] = remoteSession
			rest.logger.Info("Successfully created new remote session")
		} else {
			rest.logger.Error("Unable to get student grades report: ", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	bytes, err := json.Marshal(gradeReport)
	if err != nil {
		rest.logger.Error("Error marshalling student grades report")
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent student grades report: ", gradeReport)
}

// getReportStudentTotalRequest используется в GetReportStudentTotalHandler
type getReportStudentTotalRequest struct {
	ID   int    `json:"id"`
	From string `json:"from"`
	To   string `json:"to"`
}

// GetReportStudentTotalHandler обрабатывает запрос на получение отчета об успеваемости
// и посещаемости
func (rest *RestAPI) GetReportStudentTotalHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("GetReportStudentTotalHandler called")
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
	var rReq getReportStudentTotalRequest
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
		rest.logger.Info("No remote session, creating new one")
		userName := session.Values["userName"]
		schoolID := session.Values["schoolID"]
		school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
		if err != nil {
			rest.logger.Error("Error reading database", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		remoteSession = ss.NewSession(school)
		if err = remoteSession.Login(); err != nil {
			rest.logger.Error("Error remote signing in", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		rest.sessionsMap[sessionName] = remoteSession
	}
	studentID := strconv.Itoa(rReq.ID)
	totalReport, err := remoteSession.GetStudentTotalReport(rReq.From, rReq.To, studentID)
	// Если удаленная сессия есть в mapSessions, но не активна, создать новую
	if err != nil {
		if err.Error() == "You was logged out from server" {
			rest.logger.Info("Remote connection broken, creation new one")
			userName := session.Values["userName"]
			schoolID := session.Values["schoolID"]
			school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
			if err != nil {
				rest.logger.Error("Error reading database", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			remoteSession = ss.NewSession(school)
			if err = remoteSession.Login(); err != nil {
				rest.logger.Error("Error remote signing in", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			rest.sessionsMap[sessionName] = remoteSession
			rest.logger.Info("Successfully created new remote session")
		} else {
			rest.logger.Error("Unable to get total student report: ", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	bytes, err := json.Marshal(totalReport)
	if err != nil {
		rest.logger.Error("Error marshalling totalReport")
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent total student report: ", totalReport)
}

// getReportJournalAccessRequest используется в GetReportJournalAccessHandler
type getReportJournalAccessRequest struct {
	ID int `json:"id"`
}

// GetReportJournalAccessHandler обрабатывает запрос на получение отчета о доступе
// к классному журналу
func (rest *RestAPI) GetReportJournalAccessHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("GetReportJournalAccessHandler called")
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
	var rReq getReportJournalAccessRequest
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
		rest.logger.Info("No remote session, creating new one")
		userName := session.Values["userName"]
		schoolID := session.Values["schoolID"]
		school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
		if err != nil {
			rest.logger.Error("Error reading database", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		remoteSession = ss.NewSession(school)
		if err = remoteSession.Login(); err != nil {
			rest.logger.Error("Error remote signing in", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		rest.sessionsMap[sessionName] = remoteSession
	}
	studentID := strconv.Itoa(rReq.ID)
	accessReport, err := remoteSession.GetJournalAccessReport(studentID)
	if err != nil {
		// Если удаленная сессия есть в mapSessions, но не активна, создать новую
		if err.Error() == "You was logged out from server" {
			rest.logger.Info("Remote connection broken, creation new one")
			userName := session.Values["userName"]
			schoolID := session.Values["schoolID"]
			school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
			if err != nil {
				rest.logger.Error("Error reading database", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			remoteSession = ss.NewSession(school)
			if err = remoteSession.Login(); err != nil {
				rest.logger.Error("Error remote signing in", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			rest.sessionsMap[sessionName] = remoteSession
			rest.logger.Info("Successfully created new remote session")
		} else {
			rest.logger.Error("Unable to get journal access report: ", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	bytes, err := json.Marshal(accessReport)
	if err != nil {
		rest.logger.Error("Error marshalling accessReport")
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent total access report: ", accessReport)
}

// getReportParentInfoLetterRequest используется в GetReportParentInfoLetterHandler
type getReportParentInfoLetterRequest struct {
	StudentID    int `json:"student_id"`
	ReportTypeID int `json:"report_type_id"`
	PeriodID     int `json:"period_id"`
}

// GetReportParentInfoLetterHandler обрабатывает запрос на получение шаблона для
// письма родителям
func (rest *RestAPI) GetReportParentInfoLetterHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("GetReportParentInfoLetterHandler called")
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
	var rReq getReportParentInfoLetterRequest
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
		rest.logger.Info("No remote session, creating new one")
		userName := session.Values["userName"]
		schoolID := session.Values["schoolID"]
		school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
		if err != nil {
			rest.logger.Error("Error reading database", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		remoteSession = ss.NewSession(school)
		if err = remoteSession.Login(); err != nil {
			rest.logger.Error("Error remote signing in", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		rest.sessionsMap[sessionName] = remoteSession
	}
	studentID := strconv.Itoa(rReq.StudentID)
	reportID := strconv.Itoa(rReq.ReportTypeID)
	periodID := strconv.Itoa(rReq.PeriodID)
	parentLetter, err := remoteSession.GetParentInfoLetterReport(reportID, periodID, studentID)
	// Если удаленная сессия есть в mapSessions, но не активна, создать новую
	if err != nil {
		if err.Error() == "You was logged out from server" {
			rest.logger.Info("Remote connection broken, creation new one")
			userName := session.Values["userName"]
			schoolID := session.Values["schoolID"]
			school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
			if err != nil {
				rest.logger.Error("Error reading database", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			remoteSession = ss.NewSession(school)
			if err = remoteSession.Login(); err != nil {
				rest.logger.Error("Error remote signing in", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			rest.sessionsMap[sessionName] = remoteSession
			rest.logger.Info("Successfully created new remote session")
		} else {
			rest.logger.Error("Unable to get parent info letter report: ", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	bytes, err := json.Marshal(parentLetter)
	if err != nil {
		rest.logger.Error("Error marshalling parentLetter")
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent parent info letter report: ", parentLetter)
}

// school struct используется в GetSchoolListHandler
type school struct {
	Name    string `json:"name"`
	ID      int    `json:"id"`
	Website string `json:"website"`
}

// SchoolListResponse используется в GetSchoolListHandler
type SchoolListResponse struct {
	Schools []school `json:"schools"`
}

// GetSchoolListHandler обрабатывает запрос на получение списка обслуживаемых школ
func (rest *RestAPI) GetSchoolListHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("GetSchoolListHandler called")
	if req.Method != "GET" {
		rest.logger.Error("Wrong method: ", req.Method)
		respwr.WriteHeader(http.StatusBadRequest)
		return
	}
	schools, err := rest.db.GetSchools()
	if err != nil {
		rest.logger.Error("Error getting school list from db")
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	schoolList := make([]school, 0)
	for _, sch := range schools {
		schoolList = append(schoolList, school{sch.Name, int(sch.ID), sch.Address})
	}
	resp := SchoolListResponse{schoolList}
	bytes, err := json.Marshal(resp)
	if err != nil {
		rest.logger.Error("Error marshalling list of schools")
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent list of schools: ", resp)
}

// GetChildrenMapHandler обрабатывает запрос на получение списка детей
func (rest *RestAPI) GetChildrenMapHandler(respwr http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		rest.logger.Error("Wrong method: ", req.Method)
		respwr.WriteHeader(http.StatusBadRequest)
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
	// Если нет удаленной сессии, создать
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		rest.logger.Info("No remote session, creating new one")
		userName := session.Values["userName"]
		schoolID := session.Values["schoolID"]
		school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
		if err != nil {
			rest.logger.Error("Error reading database")
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		remoteSession = ss.NewSession(school)
		if err = remoteSession.Login(); err != nil {
			rest.logger.Error("Error remote signing in")
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		rest.sessionsMap[sessionName] = remoteSession
	}
	err = remoteSession.GetChildrenMap()
	// Если удаленная сессия есть в mapSessions, но не активна, создать новую
	if err != nil {
		if err.Error() == "You was logged out from server" {
			rest.logger.Info("Remote connection broken, creation new one")
			userName := session.Values["userName"]
			schoolID := session.Values["schoolID"]
			school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
			if err != nil {
				rest.logger.Error("Error reading database")
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			remoteSession = ss.NewSession(school)
			if err = remoteSession.Login(); err != nil {
				rest.logger.Error("Error remote signing in")
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			rest.sessionsMap[sessionName] = remoteSession
			rest.logger.Info("Successfully created new remote session")
		} else {
			rest.logger.Error("Unable to get children map: ", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	bytes, err := json.Marshal(remoteSession.Base.ChildrenIDS)
	if err != nil {
		rest.logger.Error("Error marshalling childrenMap")
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent children map: ", remoteSession.Base.ChildrenIDS)
}

// tasksMarksRequest используется в GetTasksAndMarksHandler
type tasksMarksRequest struct {
	Week string `json:"week"`
	ID   int    `json:"id"`
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
	id := strconv.Itoa(rReq.ID)
	// Если нет удаленной сессии, создать
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		rest.logger.Info("No remote session, creating new one")
		userName := session.Values["userName"]
		schoolID := session.Values["schoolID"]
		school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
		if err != nil {
			rest.logger.Error("Error reading database")
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		remoteSession = ss.NewSession(school)
		if err = remoteSession.Login(); err != nil {
			rest.logger.Error("Error remote signing in")
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		rest.sessionsMap[sessionName] = remoteSession
	}
	week := rReq.Week
	if week == "" {
		t2 := time.Now()
		t1 := t2.Weekday()
		switch t1 {
		case time.Monday:
		case time.Tuesday:
			t2 = t2.AddDate(0, 0, -1)
		case time.Wednesday:
			t2 = t2.AddDate(0, 0, -2)
		case time.Thursday:
			t2 = t2.AddDate(0, 0, -3)
		case time.Friday:
			t2 = t2.AddDate(0, 0, -4)
		case time.Saturday:
			t2 = t2.AddDate(0, 0, -5)
		case time.Sunday:
			t2 = t2.AddDate(0, 0, -6)
		}
		week = t2.Format("02.01.2006")
	}
	weekMarks, err := remoteSession.GetWeekSchoolMarks(week, id)
	// Если удаленная сессия есть в mapSessions, но не активна, создать новую
	if err != nil {
		if err.Error() == "You was logged out from server" {
			rest.logger.Info("Remote connection broken, creation new one")
			userName := session.Values["userName"]
			schoolID := session.Values["schoolID"]
			school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
			if err != nil {
				rest.logger.Error("Error reading database")
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			remoteSession = ss.NewSession(school)
			if err = remoteSession.Login(); err != nil {
				rest.logger.Error("Error remote signing in")
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			rest.sessionsMap[sessionName] = remoteSession
			rest.logger.Info("Successfully created new remote session")
		} else {
			rest.logger.Error("Unable to get tasks and marks: ", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	// Обновить статусы заданий
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	err = rest.db.UpdateTasksStatuses(userName.(string), schoolID.(int), rReq.ID, weekMarks)
	if err != nil {
		rest.logger.Error("Error updating statuses for weekMarks")
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Замаршалить
	bytes, err := json.Marshal(weekMarks)
	if err != nil {
		rest.logger.Error("Error marshalling weekMarks")
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent tasks and marks for a week: ", weekMarks)
}

// getLessonDescriptionRequest используется в GetLessonDescriptionHandler
type getLessonDescriptionRequest struct {
	ID string `json:"id"`
}

// GetLessonDescriptionHandler обрабатывает запрос на получение подробностей дз
func (rest *RestAPI) GetLessonDescriptionHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("GetLessonDescriptionHandler called")
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
	var rReq getLessonDescriptionRequest
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&rReq)
	if err != nil {
		respwr.WriteHeader(http.StatusBadRequest)
		rest.logger.Error("Malformed request data", err)
		return
	}
	taskID, err := strconv.Atoi(rReq.ID)
	if err != nil {
		respwr.WriteHeader(http.StatusBadRequest)
		rest.logger.Error("Malformed request data", err)
		return
	}
	// Если нет удаленной сессии, создать
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		rest.logger.Info("No remote session, creating new one")
		userName := session.Values["userName"]
		schoolID := session.Values["schoolID"]
		school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
		if err != nil {
			rest.logger.Error("Error reading database")
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		remoteSession = ss.NewSession(school)
		if err = remoteSession.Login(); err != nil {
			rest.logger.Error("Error remote signing in")
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		rest.sessionsMap[sessionName] = remoteSession
	}
	// Сходить в бд за информацией о таске
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	date, cid, tp, studentID, err := rest.db.GetTaskInfo(userName.(string), schoolID.(int), taskID)
	if err != nil {
		if err.Error() == "record not found" {
			rest.logger.Info("Invalid task specified: it's not in db", err)
			respwr.WriteHeader(http.StatusBadRequest)
			return
		}
		rest.logger.Error("Error getting task date from db")
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Получить описание таска
	lessonDescription, err := remoteSession.GetLessonDescription(date, taskID, cid, tp, strconv.Itoa(studentID))
	// Если удаленная сессия есть в mapSessions, но не активна, создать новую
	if err != nil {
		if err.Error() == "You was logged out from server" {
			rest.logger.Info("Remote connection broken, creation new one")
			userName := session.Values["userName"]
			schoolID := session.Values["schoolID"]
			school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
			if err != nil {
				rest.logger.Error("Error reading database")
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			remoteSession = ss.NewSession(school)
			if err = remoteSession.Login(); err != nil {
				rest.logger.Error("Error remote signing in")
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			rest.sessionsMap[sessionName] = remoteSession
			rest.logger.Info("Successfully created new remote session")
		} else {
			rest.logger.Error("Unable to get lesson description: ", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	// Замаршалить
	bytes, err := json.Marshal(lessonDescription)
	if err != nil {
		rest.logger.Error("Error marshalling lesson description", err)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	respwr.Write(bytes)
	rest.logger.Info("Sent lesson description: ", lessonDescription)

}

// markAsDoneRequest используется в MarkAsDoneHandler
type MarkAsDoneRequest struct {
	ID int `json:"id"`
}

// MarkAsDoneHandler обрабатывает запрос на отметку задания как сделанного
func (rest *RestAPI) MarkAsDoneHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("MarkAsDoneHandler called")
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
	var rReq MarkAsDoneRequest
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
		rest.logger.Info("No remote session, creating new one")
		userName := session.Values["userName"]
		schoolID := session.Values["schoolID"]
		school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
		if err != nil {
			rest.logger.Error("Error reading database")
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		remoteSession = ss.NewSession(school)
		if err = remoteSession.Login(); err != nil {
			rest.logger.Error("Error remote signing in")
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		rest.sessionsMap[sessionName] = remoteSession
	}
	// Если удаленная сессия есть в mapSessions, но не активна, создать новую
	if err != nil {
		if err.Error() == "You was logged out from server" {
			rest.logger.Info("Remote connection broken, creation new one")
			userName := session.Values["userName"]
			schoolID := session.Values["schoolID"]
			school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
			if err != nil {
				rest.logger.Error("Error reading database")
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			remoteSession = ss.NewSession(school)
			if err = remoteSession.Login(); err != nil {
				rest.logger.Error("Error remote signing in")
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			rest.sessionsMap[sessionName] = remoteSession
			rest.logger.Info("Successfully created new remote session")
		} else {
			rest.logger.Error("Unable to mark task as done: ", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	// Лезть в БД
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	err = rest.db.TaskMarkDone(userName.(string), schoolID.(int), rReq.ID)
	if err != nil {
		rest.logger.Error("Error when marking task as done in db", err)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	rest.logger.Info("Successfully marked task as done")
	respwr.WriteHeader(http.StatusOK)
}

// unmarkAsDoneRequest используется в UnmarkAsDoneHandler
type UnmarkAsDoneRequest struct {
	ID int `json:"id"`
}

// UnmarkAsDoneHandler обрабатывает запрос на отметку задания как просмотренного
func (rest *RestAPI) UnmarkAsDoneHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("UnmarkAsDoneHandler called")
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
	var rReq MarkAsDoneRequest
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
		rest.logger.Info("No remote session, creating new one")
		userName := session.Values["userName"]
		schoolID := session.Values["schoolID"]
		school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
		if err != nil {
			rest.logger.Error("Error reading database")
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		remoteSession = ss.NewSession(school)
		if err = remoteSession.Login(); err != nil {
			rest.logger.Error("Error remote signing in")
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		rest.sessionsMap[sessionName] = remoteSession
	}
	// Если удаленная сессия есть в mapSessions, но не активна, создать новую
	if err != nil {
		if err.Error() == "You was logged out from server" {
			rest.logger.Info("Remote connection broken, creation new one")
			userName := session.Values["userName"]
			schoolID := session.Values["schoolID"]
			school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
			if err != nil {
				rest.logger.Error("Error reading database")
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			remoteSession = ss.NewSession(school)
			if err = remoteSession.Login(); err != nil {
				rest.logger.Error("Error remote signing in")
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			rest.sessionsMap[sessionName] = remoteSession
			rest.logger.Info("Successfully created new remote session")
		} else {
			rest.logger.Error("Unable to mark task as undone: ", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	// Лезть в БД
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	err = rest.db.TaskMarkUndone(userName.(string), schoolID.(int), rReq.ID)
	if err != nil {
		rest.logger.Error("Error when marking task as undone in db", err)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	rest.logger.Info("Successfully marked task as undone")
	respwr.WriteHeader(http.StatusOK)
}

// scheduleRequest используется в GetScheduleHandler
type scheduleRequest struct {
	Days int `json:"days"`
	ID   int `json:"id"`
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
	id := strconv.Itoa(rReq.ID)
	// Если нет удаленной сессии, создать
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		rest.logger.Info("No remote session, creating new one")
		userName := session.Values["userName"]
		schoolID := session.Values["schoolID"]
		school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
		if err != nil {
			rest.logger.Error("Error reading database", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		remoteSession = ss.NewSession(school)
		if err = remoteSession.Login(); err != nil {
			rest.logger.Error("Error remote signing in", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		rest.sessionsMap[sessionName] = remoteSession
	}
	today := time.Now().Format("02.01.2006")
	timeTable, err := remoteSession.GetTimeTable(today, rReq.Days, id)
	// Если удаленная сессия есть в mapSessions, но не активна, создать новую
	if err != nil {
		if err.Error() == "You was logged out from server" {
			rest.logger.Info("Remote connection broken, creation new one")
			userName := session.Values["userName"]
			schoolID := session.Values["schoolID"]
			school, err := rest.db.GetUserAuthData(userName.(string), schoolID.(int))
			if err != nil {
				rest.logger.Error("Error reading database", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			remoteSession = ss.NewSession(school)
			if err = remoteSession.Login(); err != nil {
				rest.logger.Error("Error remote signing in", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			rest.sessionsMap[sessionName] = remoteSession
			rest.logger.Info("Successfully created new remote session")
		} else {
			rest.logger.Error("Unable to get schedule: ", err)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	bytes, err := json.Marshal(timeTable)
	if err != nil {
		rest.logger.Error("Error marshalling timeTable")
		respwr.WriteHeader(http.StatusInternalServerError)
		return
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
		rest.logger.Info("Error getting session: ", err)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Вызвать logout для удаленной сессии
	err = rest.sessionsMap[sessionName].Logout()
	if err != nil {
		rest.logger.Error("Error remote log out", err)
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
	ID      int    `json:"id"`
}

// SignInHandler обрабатывает вход в учетную запись на сайте школы
func (rest *RestAPI) SignInHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("SignInHandler called")
	if req.Method != "POST" {
		rest.logger.Error("Wrong method: ", req.Method)
		return
	}
	// Чтение запроса от клиента
	var rReq signInRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		respwr.WriteHeader(http.StatusBadRequest)
		rest.logger.Error("Malformed request data")
		return
	}
	// Проверим разрешение у школы
	perm, err := rest.db.GetSchoolPermission(rReq.ID)
	if err != nil {
		rest.logger.Error("Invalid id param specified")
		respwr.WriteHeader(http.StatusBadRequest)
		return
	}
	if !perm {
		// Если у школы нет разрешения, проверить разрешение пользователя
		userPerm, err := rest.db.GetUserPermission(rReq.Login, rReq.ID)
		if err != nil {
			if err.Error() == "record not found" {
				// Пользователь новый, вернем true
				perm = true
			} else {
				rest.logger.Error("Getting permission from db: ", err)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		perm = userPerm
	}
	if !perm {
		rest.logger.Info("Access to service denied!")
		respwr.WriteHeader(http.StatusBadRequest)
		return
	}
	rest.logger.Info("Valid data:", rReq)
	school := rest.config.Schools[rReq.ID-1]
	school.Login = rReq.Login
	school.Password = rReq.Passkey
	// Создание удаленной сессии
	newRemoteSession := ss.NewSession(&school)
	err = newRemoteSession.Login()
	if err != nil {
		rest.logger.Error("Error remote signing in", err)
		respwr.WriteHeader(http.StatusBadRequest)
		return
	}
	// Сразу получим мапу имен детей в их ID
	err = newRemoteSession.GetChildrenMap()
	if err != nil {
		rest.logger.Error("Error: can't get children map", err)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Если удаленная авторизация прошла успешно, создать локальную сессию
	timeString := time.Now().String()
	hasher := md5.New()
	if _, err = hasher.Write([]byte(timeString)); err != nil {
		rest.logger.Error("Md5 hashing error: ", err)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	newSessionName := hex.EncodeToString(hasher.Sum(nil))
	newLocalSession, err := rest.store.Get(req, newSessionName)
	if err != nil {
		rest.logger.Error("Error creating new local session")
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// ... и привязать к ней удаленную сессию
	rest.sessionsMap[newSessionName] = newRemoteSession
	newLocalSession.Values["userName"] = rReq.Login
	newLocalSession.Values["schoolID"] = rReq.ID
	newLocalSession.Save(req, respwr)
	// Устанавливаем в куки значение sessionName
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{
		Name: "sessionName", Value: newSessionName, Expires: expiration,
	}
	http.SetCookie(respwr, &cookie)
	// Обновляем базу данных
	isParent := true
	err = rest.db.UpdateUser(rReq.Login, rReq.Passkey, isParent, rReq.ID, newRemoteSession.Base.ChildrenIDS)
	if err != nil {
		rest.logger.Error("Error updating database: ", err)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	rest.logger.Info("Successfully signed in as user: ", rReq.Login)
}

// Handler временный абстрактный handler для некоторых еще не реализованных
// обработчиков запросов.
func (rest *RestAPI) Handler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("Handler called (not implemented yet)", req.URL.EscapedPath())
}
