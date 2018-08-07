// rest-api.go

/*
Package restapi содержит handler'ы для взаимодействия сервера с клиентами.
*/
package restapi

import (
	cp "SchoolServer/libtelco/config-parser"
	red "SchoolServer/libtelco/in-memory-db"
	"SchoolServer/libtelco/log"
	ss "SchoolServer/libtelco/sessions"
	db "SchoolServer/libtelco/sql-db"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	redistore "gopkg.in/boj/redistore.v1"
)

// RestAPI struct содержит конфигурацию Rest API.
// sessionsMap содержит отображения идентификаторов сессий Rest API
// в объекты сессий на удаленном сервере.
type RestAPI struct {
	config      *cp.Config
	Store       *redistore.RediStore
	logger      *log.Logger
	sessionsMap map[string]*ss.Session
	Db          *db.Database
	Redis       *red.Database
}

// NewRestAPI создает структуру для работы с Rest API.
func NewRestAPI(logger *log.Logger, config *cp.Config) *RestAPI {
	key := make([]byte, 32)
	rand.Read(key)
	logger.Info("REST: Successfully generated secure key", "Key", key)
	// redistore
	newStore, err := redistore.NewRediStoreWithDB(
		1,
		"tcp",
		":"+config.CookieStore.Port,
		config.CookieStore.Password,
		config.CookieStore.DBname,
		key,
	)
	if err != nil {
		logger.Error("REST: Error occured when creating redistore", "Error", err)
	} else {
		logger.Info("REST: Successfully initialized redistore")
	}
	newStore.SetMaxAge(86400 * 365)
	// gorm
	database, err := db.NewDatabase(logger, config)
	if err != nil {
		logger.Error("REST: Error occured when initializing database", "Error", err)
	} else {
		logger.Info("REST: Successfully initialized database")
	}
	// redis
	redis, err := red.NewDatabase(config)
	if err != nil {
		logger.Error("REST: Error occured when initializing redis", "Error", err)
	} else {
		logger.Info("REST: Successfully initialized redis")
	}
	return &RestAPI{
		config:      config,
		Store:       newStore,
		logger:      logger,
		sessionsMap: make(map[string]*ss.Session),
		Db:          database,
		Redis:       redis,
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
	rest.logger.Info("REST: Successfully bound handlers")
}

// permissionCheckRequest используется в CheckPermissionHandler
type permissionCheckRequest struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
}

// permissionCheckResponse используется в CheckPermissionHandler
type permissionCheckResponse struct {
	Permission bool `json:"permission"`
}

// CheckPermissionHandler проверяет, есть ли разрешение на работу с школой
func (rest *RestAPI) CheckPermissionHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: CheckPermissionHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "POST" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Чтение json'a
	var rReq permissionCheckRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Info("REST: Malformed request data", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		resp := "malformed data"
		status, err := respwr.Write([]byte(resp))
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
		}
		return
	}
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// Проверим разрешение у школы
	perm, err := rest.Db.GetSchoolPermission(rReq.ID)
	if err != nil {
		if err.Error() == "record not found" {
			// Школа не найдена
			rest.logger.Info("REST: Invalid school id specified", "Id", rReq.ID, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadRequest)
			resp := "invalid id"
			status, err := respwr.Write([]byte(resp))
			if err != nil {
				rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
			} else {
				rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting school permission from db", "Error", err, "Id", rReq.ID, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if !perm {
		// Если у школы нет разрешения, проверить разрешение пользователя
		userPerm, err := rest.Db.GetUserPermission(rReq.Login, rReq.ID)
		if err != nil {
			if err.Error() == "record not found" {
				// Пользователь новый: вернем true, чтобы он мог залогиниться и
				// купить подписку
				perm = true
			} else {
				// Другая ошибка
				rest.logger.Error("REST: Error occured when getting user permission from db", "Error", err, "Login", rReq.Login, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			perm = userPerm
		}
	}
	// Закодировать ответ в JSON
	resp := permissionCheckResponse{perm}
	bytes, err := json.Marshal(resp)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", resp, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
	}
}

// ErrorHandler обрабатывает некорректные запросы.
func (rest *RestAPI) ErrorHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("Wrong request", "Path", req.URL.EscapedPath(), "IP", req.RemoteAddr)
	respwr.WriteHeader(http.StatusNotFound)
}

// reportStudentTotalMarksGetRequest используется в GetReportStudentTotalMarksHandler
type reportStudentTotalMarksGetRequest struct {
	ID int `json:"id"`
}

// getLocalSession читает куки и получает объект локальной сессии
func (rest *RestAPI) getLocalSession(respwr http.ResponseWriter, req *http.Request) (string, *sessions.Session) {
	// Прочитать куку
	cookie, err := req.Cookie("sessionName")
	if err != nil {
		rest.logger.Info("REST: User not authorized", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusUnauthorized)
		return "", nil
	}
	sessionName := cookie.Value
	// Получить существующий объект сессии
	session, err := rest.Store.Get(req, sessionName)
	if err != nil {
		rest.logger.Error("REST: Error occured when getting session from cookiestore", "Error", err, "Session name", sessionName, "IP", req.RemoteAddr)
		delete(rest.sessionsMap, sessionName)
		session.Options.MaxAge = -1
		session.Save(req, respwr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return "", nil
	}
	if session.IsNew {
		rest.logger.Info("REST: No session exists", "Session name", sessionName, "IP", req.RemoteAddr)
		delete(rest.sessionsMap, sessionName)
		session.Options.MaxAge = -1
		session.Save(req, respwr)
		respwr.WriteHeader(http.StatusUnauthorized)
		return "", nil
	}
	return sessionName, session
}

// remoteLogin авторизуется на сайте школы
func (rest *RestAPI) remoteLogin(respwr http.ResponseWriter, req *http.Request, session *sessions.Session) *ss.Session {
	rest.logger.Info("REST: Remote signing in", "IP", req.RemoteAddr)
	// Полезть в базу данных за данными для авторизации
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	school, err := rest.Db.GetUserAuthData(userName.(string), schoolID.(int))
	if err != nil {
		// Ошибок тут быть не должно
		rest.logger.Error("REST: Error occured when getting user auth data from db", "Username", userName, "SchoolID", schoolID, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	// Создать удаленную сессию и залогиниться
	remoteSession := ss.NewSession(school)
	err = remoteSession.Login()
	if err != nil {
		rest.logger.Error("REST: Error occured when remote signing in", "Error", err, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadGateway)
		return nil
	}
	rest.sessionsMap[session.Name()] = remoteSession
	rest.logger.Info("REST: Successfully created new remote session", "IP", req.RemoteAddr)
	return remoteSession
}

// GetReportStudentTotalMarksHandler обрабатывает запрос на получение отчета
// об итоговых оценках
func (rest *RestAPI) GetReportStudentTotalMarksHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetReportStudentTotalMarksHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "POST" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	sessionName, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
	}
	// Чтение запроса от клиента
	var rReq reportStudentTotalMarksGetRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Info("REST: Malformed request data", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		resp := "malformed data"
		status, err := respwr.Write([]byte(resp))
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
		}
		return
	}
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// Получим удаленную сессию
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		// Если нет удаленной сессии
		rest.logger.Info("REST: No remote session", "IP", req.RemoteAddr)
		// Создать новую
		remoteSession = rest.remoteLogin(respwr, req, session)
		if remoteSession == nil {
			return
		}
	}
	// Получить отчет с сайта школы
	totalMarkReport, err := remoteSession.GetTotalMarkReport(strconv.Itoa(rReq.ID))
	if err != nil {
		if err.Error() == "You was logged out from server" {
			// Если удаленная сессия есть, но не активна
			rest.logger.Info("REST: Remote connection timed out", "IP", req.RemoteAddr)
			// Создать новую
			remoteSession = rest.remoteLogin(respwr, req, session)
			if remoteSession == nil {
				return
			}
			// Повторно получить отчет с сайта школы
			totalMarkReport, err = remoteSession.GetTotalMarkReport(strconv.Itoa(rReq.ID))
			if err != nil {
				// Ошибка
				rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadGateway)
				return
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadGateway)
			return
		}
	}
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(totalMarkReport)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", totalMarkReport, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", totalMarkReport, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", totalMarkReport, "IP", req.RemoteAddr)
	}
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
	rest.logger.Info("REST: GetReportStudentAverageMarkHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "POST" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	sessionName, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
	}
	// Чтение запроса от клиента
	var rReq getReportStudentAverageMarkRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Info("REST: Malformed request data", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		resp := "malformed data"
		status, err := respwr.Write([]byte(resp))
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
		}
		return
	}
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// Получим удаленную сессию
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		// Если нет удаленной сессии
		rest.logger.Info("REST: No remote session", "IP", req.RemoteAddr)
		// Создать новую
		remoteSession = rest.remoteLogin(respwr, req, session)
		if remoteSession == nil {
			return
		}
	}
	// Получить отчет с сайта школы
	averageMarkReport, err := remoteSession.GetAverageMarkReport(rReq.From, rReq.To, rReq.Type, strconv.Itoa(rReq.ID))
	if err != nil {
		if err.Error() == "You was logged out from server" {
			// Если удаленная сессия есть, но не активна
			rest.logger.Info("REST: Remote connection timed out", "IP", req.RemoteAddr)
			// Создать новую
			remoteSession = rest.remoteLogin(respwr, req, session)
			if remoteSession == nil {
				return
			}
			// Повторно получить отчет с сайта школы
			averageMarkReport, err = remoteSession.GetAverageMarkReport(rReq.From, rReq.To, rReq.Type, strconv.Itoa(rReq.ID))
			if err != nil {
				// Ошибка
				rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadGateway)
				return
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadGateway)
			return
		}
	}
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(averageMarkReport)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", averageMarkReport, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", averageMarkReport, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", averageMarkReport, "IP", req.RemoteAddr)
	}
}

// GetReportStudentAverageMarkDynHandler обрабатывает запрос на получение отчета
// о динамике среднего балла
func (rest *RestAPI) GetReportStudentAverageMarkDynHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetReportStudentAverageMarkDynHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "POST" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	sessionName, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
	}
	// Чтение запроса от клиента
	var rReq getReportStudentAverageMarkRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Info("REST: Malformed request data", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		resp := "malformed data"
		status, err := respwr.Write([]byte(resp))
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
		}
		return
	}
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// Получим удаленную сессию
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		// Если нет удаленной сессии
		rest.logger.Info("REST: No remote session", "IP", req.RemoteAddr)
		// Создать новую
		remoteSession = rest.remoteLogin(respwr, req, session)
		if remoteSession == nil {
			return
		}
	}
	// Получить отчет с сайта школы
	averageMarkDynReport, err := remoteSession.GetAverageMarkDynReport(rReq.From, rReq.To, rReq.Type, strconv.Itoa(rReq.ID))
	if err != nil {
		if err.Error() == "You was logged out from server" {
			// Если удаленная сессия есть, но не активна
			rest.logger.Info("REST: Remote connection timed out", "IP", req.RemoteAddr)
			// Создать новую
			remoteSession = rest.remoteLogin(respwr, req, session)
			if remoteSession == nil {
				return
			}
			// Повторно получить отчет с сайта школы
			averageMarkDynReport, err = remoteSession.GetAverageMarkDynReport(rReq.From, rReq.To, rReq.Type, strconv.Itoa(rReq.ID))
			if err != nil {
				// Ошибка
				rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadGateway)
				return
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadGateway)
			return
		}
	}
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(averageMarkDynReport)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", averageMarkDynReport, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", averageMarkDynReport, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", averageMarkDynReport, "IP", req.RemoteAddr)
	}
}

// getReportStudentGradesLessonListRequest используется в GetReportStudentGradesLessonListHandler
type getReportStudentGradesLessonListRequest struct {
	ID int `json:"id"`
}

// GetReportStudentGradesLessonListHandler обрабатывает запрос на получение
// списка предметов для отчета 'Об успеваемости'
func (rest *RestAPI) GetReportStudentGradesLessonListHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetReportStudentGradesLessonListHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "POST" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	sessionName, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
	}
	// Чтение запроса от клиента
	var rReq getReportStudentGradesLessonListRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Info("REST: Malformed request data", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		resp := "malformed data"
		status, err := respwr.Write([]byte(resp))
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
		}
		return
	}
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// Получим удаленную сессию
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		// Если нет удаленной сессии
		rest.logger.Info("REST: No remote session", "IP", req.RemoteAddr)
		// Создать новую
		remoteSession = rest.remoteLogin(respwr, req, session)
		if remoteSession == nil {
			return
		}
	}
	// Получить отчет с сайта школы
	lessonsMap, err := remoteSession.GetLessonsMap(strconv.Itoa(rReq.ID))
	if err != nil {
		if err.Error() == "You was logged out from server" {
			// Если удаленная сессия есть, но не активна
			rest.logger.Info("REST: Remote connection timed out", "IP", req.RemoteAddr)
			// Создать новую
			remoteSession = rest.remoteLogin(respwr, req, session)
			if remoteSession == nil {
				return
			}
			// Повторно получить отчет с сайта школы
			lessonsMap, err = remoteSession.GetLessonsMap(strconv.Itoa(rReq.ID))
			if err != nil {
				// Ошибка
				rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadGateway)
				return
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadGateway)
			return
		}
	}
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(lessonsMap)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", lessonsMap, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", lessonsMap, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", lessonsMap, "IP", req.RemoteAddr)
	}
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
	rest.logger.Info("REST: GetReportStudentGradesHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "POST" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	sessionName, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
	}
	// Чтение запроса от клиента
	var rReq getReportStudentGradesRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Info("REST: Malformed request data", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		resp := "malformed data"
		status, err := respwr.Write([]byte(resp))
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
		}
		return
	}
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// Получим удаленную сессию
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		// Если нет удаленной сессии
		rest.logger.Info("REST: No remote session", "IP", req.RemoteAddr)
		// Создать новую
		remoteSession = rest.remoteLogin(respwr, req, session)
		if remoteSession == nil {
			return
		}
	}
	// Получить отчет с сайта школы
	gradeReport, err := remoteSession.GetStudentGradeReport(rReq.From, rReq.To, strconv.Itoa(rReq.LessonID), strconv.Itoa(rReq.StudentID))
	if err != nil {
		if err.Error() == "You was logged out from server" {
			// Если удаленная сессия есть, но не активна
			rest.logger.Info("REST: Remote connection timed out", "IP", req.RemoteAddr)
			// Создать новую
			remoteSession = rest.remoteLogin(respwr, req, session)
			if remoteSession == nil {
				return
			}
			// Повторно получить отчет с сайта школы
			gradeReport, err = remoteSession.GetStudentGradeReport(rReq.From, rReq.To, strconv.Itoa(rReq.LessonID), strconv.Itoa(rReq.StudentID))
			if err != nil {
				// Ошибка
				rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadGateway)
				return
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadGateway)
			return
		}
	}
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(gradeReport)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", gradeReport, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", gradeReport, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", gradeReport, "IP", req.RemoteAddr)
	}
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
	rest.logger.Info("REST: GetReportStudentTotalHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "POST" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	sessionName, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
	}
	// Чтение запроса от клиента
	var rReq getReportStudentTotalRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Info("REST: Malformed request data", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		resp := "malformed data"
		status, err := respwr.Write([]byte(resp))
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
		}
		return
	}
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// Получим удаленную сессию
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		// Если нет удаленной сессии
		rest.logger.Info("REST: No remote session", "IP", req.RemoteAddr)
		// Создать новую
		remoteSession = rest.remoteLogin(respwr, req, session)
		if remoteSession == nil {
			return
		}
	}
	// Получить отчет с сайта школы
	studentID := strconv.Itoa(rReq.ID)
	totalReport, err := remoteSession.GetStudentTotalReport(rReq.From, rReq.To, studentID)
	if err != nil {
		if err.Error() == "You was logged out from server" {
			// Если удаленная сессия есть, но не активна
			rest.logger.Info("REST: Remote connection timed out", "IP", req.RemoteAddr)
			// Создать новую
			remoteSession = rest.remoteLogin(respwr, req, session)
			if remoteSession == nil {
				return
			}
			// Повторно получить отчет с сайта школы
			totalReport, err = remoteSession.GetStudentTotalReport(rReq.From, rReq.To, studentID)
			if err != nil {
				// Ошибка
				rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadGateway)
				return
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadGateway)
			return
		}
	}
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(totalReport)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", totalReport, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", totalReport, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", totalReport, "IP", req.RemoteAddr)
	}
}

// getReportJournalAccessRequest используется в GetReportJournalAccessHandler
type getReportJournalAccessRequest struct {
	ID int `json:"id"`
}

// GetReportJournalAccessHandler обрабатывает запрос на получение отчета о доступе
// к классному журналу
func (rest *RestAPI) GetReportJournalAccessHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetReportJournalAccessHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "POST" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	sessionName, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
	}
	// Чтение запроса от клиента
	var rReq getReportJournalAccessRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Info("REST: Malformed request data", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		resp := "malformed data"
		status, err := respwr.Write([]byte(resp))
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
		}
		return
	}
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// Получим удаленную сессию
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		// Если нет удаленной сессии
		rest.logger.Info("REST: No remote session", "IP", req.RemoteAddr)
		// Создать новую
		remoteSession = rest.remoteLogin(respwr, req, session)
		if remoteSession == nil {
			return
		}
	}
	// Получить отчет с сайта школы
	studentID := strconv.Itoa(rReq.ID)
	accessReport, err := remoteSession.GetJournalAccessReport(studentID)
	if err != nil {
		if err.Error() == "You was logged out from server" {
			// Если удаленная сессия есть, но не активна
			rest.logger.Info("REST: Remote connection timed out", "IP", req.RemoteAddr)
			// Создать новую
			remoteSession = rest.remoteLogin(respwr, req, session)
			if remoteSession == nil {
				return
			}
			// Повторно получить отчет с сайта школы
			accessReport, err = remoteSession.GetJournalAccessReport(studentID)
			if err != nil {
				// Ошибка
				rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadGateway)
				return
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadGateway)
			return
		}
	}
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(accessReport)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", accessReport, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", accessReport, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", accessReport, "IP", req.RemoteAddr)
	}
}

// getReportParentInfoLetterRequest используется в GetReportParentInfoLetterHandler
type getReportParentInfoLetterRequest struct {
	StudentID    int `json:"student_id"`
	ReportTypeID int `json:"report_type_id"`
	PeriodID     int `json:"period_id"`
}

// GetReportParentInfoLetterHandler обрабатывает запрос на получение шаблона
// письма родителям
func (rest *RestAPI) GetReportParentInfoLetterHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetReportParentInfoLetterHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "POST" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	sessionName, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
	}
	// Чтение запроса от клиента
	var rReq getReportParentInfoLetterRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Info("REST: Malformed request data", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		resp := "malformed data"
		status, err := respwr.Write([]byte(resp))
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
		}
		return
	}
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// Получим удаленную сессию
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		// Если нет удаленной сессии
		rest.logger.Info("REST: No remote session", "IP", req.RemoteAddr)
		// Создать новую
		remoteSession = rest.remoteLogin(respwr, req, session)
		if remoteSession == nil {
			return
		}
	}
	// Получить отчет с сайта школы
	studentID := strconv.Itoa(rReq.StudentID)
	reportID := strconv.Itoa(rReq.ReportTypeID)
	periodID := strconv.Itoa(rReq.PeriodID)
	parentLetter, err := remoteSession.GetParentInfoLetterReport(reportID, periodID, studentID)
	if err != nil {
		if err.Error() == "You was logged out from server" {
			// Если удаленная сессия есть, но не активна
			rest.logger.Info("REST: Remote connection timed out", "IP", req.RemoteAddr)
			// Создать новую
			remoteSession = rest.remoteLogin(respwr, req, session)
			if remoteSession == nil {
				return
			}
			// Повторно получить отчет с сайта школы
			parentLetter, err = remoteSession.GetParentInfoLetterReport(reportID, periodID, studentID)
			if err != nil {
				// Ошибка
				rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadGateway)
				return
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadGateway)
			return
		}
	}
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(parentLetter)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", parentLetter, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", parentLetter, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", parentLetter, "IP", req.RemoteAddr)
	}
}

// school struct используется в GetSchoolListHandler
type school struct {
	Name     string `json:"name"`
	ID       int    `json:"id"`
	Website  string `json:"website"`
	Shortcut string `json:"shortcut"`
}

// SchoolListResponse используется в GetSchoolListHandler
type SchoolListResponse struct {
	Schools []school `json:"schools"`
}

// GetSchoolListHandler обрабатывает запрос на получение списка обслуживаемых школ
func (rest *RestAPI) GetSchoolListHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetSchoolListHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "GET" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Залезть в БД за списком школ
	schools, err := rest.Db.GetSchools()
	if err != nil {
		rest.logger.Error("REST: Error occured when getting school list from db", "Error", err, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Заполняем пакет ответа
	schoolList := make([]school, 0)
	for _, sch := range schools {
		schoolList = append(schoolList, school{sch.Name, int(sch.ID), sch.Address, sch.Initials})
	}
	resp := SchoolListResponse{schoolList}
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(resp)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", resp, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
	}
}

// GetChildrenMapResponse используется в GetChildrenMapHandler
type GetChildrenMapResponse struct {
	Students []student `json:"students"`
}

// GetChildrenMapHandler обрабатывает запрос на получение списка детей
func (rest *RestAPI) GetChildrenMapHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetChildrenMapHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "GET" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	sessionName, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
	}
	// Получим удаленную сессию
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		// Если нет удаленной сессии
		rest.logger.Info("REST: No remote session", "IP", req.RemoteAddr)
		// Создать новую
		remoteSession = rest.remoteLogin(respwr, req, session)
		if remoteSession == nil {
			return
		}
	}
	// Заполнить пакет ответа
	var stud student
	studs := make([]student, 0)
	// Проверить наличие мапы у сессии парсеров
	if remoteSession.Base.Children == nil || len(remoteSession.Base.Children) == 0 {
		// Если мапа не существует или пустая, полезем в БД
		userName := session.Values["userName"]
		schoolID := session.Values["schoolID"]
		res, err := rest.Db.GetStudents(userName.(string), schoolID.(int))
		if err != nil {
			rest.logger.Error("REST: Error occured when getting children map from db", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		for k, v := range res {
			stud = student{Name: k, ID: v}
			studs = append(studs, stud)
		}
	} else {
		// Если мапа есть
		res := remoteSession.Base.Children
		for k, v := range res {
			vs, err := strconv.Atoi(v.SID)
			if err != nil {
				rest.logger.Error("REST: Error occured when converting student id", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			stud = student{Name: k, ID: vs}
			studs = append(studs, stud)
		}
	}
	// Замаршалить
	resp := GetChildrenMapResponse{Students: studs}
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(resp)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", resp, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
	}
}

// tasksMarksRequest используется в GetTasksAndMarksHandler
type tasksMarksRequest struct {
	Week string `json:"week"`
	ID   int    `json:"id"`
}

// lesson используется в day
type lesson struct {
	ID     int    `json:"id"`
	Status int    `json:"status"`
	InTime bool   `json:"inTime"`
	Name   string `json:"name"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	Mark   string `json:"mark"`
	Weight string `json:"weight"`
}

// day используется в tasksMarksResponse
type day struct {
	Date    string   `json:"date"`
	Lessons []lesson `json:"lessons"`
}

// tasksMarksResponse используется в GetTasksAndMarksHandler
type tasksMarksResponse struct {
	Days []day `json:"days"`
}

// GetTasksAndMarksHandler возвращает задания и оценки на неделю
func (rest *RestAPI) GetTasksAndMarksHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetTasksAndMarksHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "POST" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	sessionName, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
	}
	// Чтение запроса от клиента
	var rReq tasksMarksRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Info("REST: Malformed request data", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		resp := "malformed data"
		status, err := respwr.Write([]byte(resp))
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
		}
		return
	}
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// По протоколу пустое поле заменить на дату первого дня текущей недели
	// (пока просто текущий день)
	if rReq.Week == "" {
		rReq.Week = time.Now().Format("02.01.2006")
	}
	// Получим удаленную сессию
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		// Если нет удаленной сессии
		rest.logger.Info("REST: No remote session", "IP", req.RemoteAddr)
		// Создать новую
		remoteSession = rest.remoteLogin(respwr, req, session)
		if remoteSession == nil {
			return
		}
	}
	// Получить с сайта школы
	weekMarks, err := remoteSession.GetWeekSchoolMarks(rReq.Week, strconv.Itoa(rReq.ID))
	if err != nil {
		if err.Error() == "You was logged out from server" {
			// Если удаленная сессия есть, но не активна
			rest.logger.Info("REST: Remote connection timed out", "IP", req.RemoteAddr)
			// Создать новую
			remoteSession = rest.remoteLogin(respwr, req, session)
			if remoteSession == nil {
				return
			}
			// Повторно получить с сайта школы
			weekMarks, err = remoteSession.GetWeekSchoolMarks(rReq.Week, strconv.Itoa(rReq.ID))
			if err != nil {
				// Ошибка
				rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadGateway)
				return
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadGateway)
			return
		}
	}
	// Обновить статусы заданий
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	// Сходить в бд
	err = rest.Db.UpdateTasksStatuses(userName.(string), schoolID.(int), rReq.ID, weekMarks)
	if err != nil {
		rest.logger.Error("REST: Error occured when updating statuses for tasks and marks", "Error", err, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Сформировать ответ по протоколу
	var resp tasksMarksResponse
	days := make([]day, 0)
	for _, d := range weekMarks.Data {
		lessons := make([]lesson, 0)
		for _, l := range d.Lessons {
			newLesson := lesson{ID: l.AID, Status: l.Status, InTime: l.InTime, Name: l.Name, Title: l.Title, Type: l.Type, Mark: l.Mark, Weight: l.Weight}
			lessons = append(lessons, newLesson)
		}
		newDay := day{Date: d.Date, Lessons: lessons}
		days = append(days, newDay)
	}
	resp.Days = days
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(resp)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", resp, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
	}
}

// getLessonDescriptionRequest используется в GetLessonDescriptionHandler
type getLessonDescriptionRequest struct {
	ID int `json:"id"`
}

// getLessonDescriptionResponse использутеся в GetLessonDescriptionHandler
type getLessonDescriptionResponse struct {
	Description string `json:"description"`
	Author      string `json:"author"`
	File        string `json:"file"`
	FileName    string `json:"fileName"`
}

// GetLessonDescriptionHandler обрабатывает запрос на получение подробностей дз
func (rest *RestAPI) GetLessonDescriptionHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetLessonDescriptionHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "POST" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	sessionName, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
	}
	// Чтение запроса от клиента
	var rReq getLessonDescriptionRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Info("REST: Malformed request data", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		resp := "malformed data"
		status, err := respwr.Write([]byte(resp))
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
		}
		return
	}
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// Получим удаленную сессию
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		// Если нет удаленной сессии
		rest.logger.Info("REST: No remote session", "IP", req.RemoteAddr)
		// Создать новую
		remoteSession = rest.remoteLogin(respwr, req, session)
		if remoteSession == nil {
			return
		}
	}
	// Сходить в бд за информацией о таске
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	date, cid, tp, studentID, err := rest.Db.GetTaskInfo(userName.(string), schoolID.(int), rReq.ID)
	if err != nil {
		if err.Error() == "record not found" {
			// Такого таска нет
			rest.logger.Info("REST: Invalid task id specified", "Error", err.Error(), "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadRequest)
			resp := "invalid id"
			status, err := respwr.Write([]byte(resp))
			if err != nil {
				rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
			} else {
				rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting task date from db", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// Получить описание таска
	lessonDescription, err := remoteSession.GetLessonDescription(date, rReq.ID, cid, tp, strconv.Itoa(studentID))
	if err != nil {
		if err.Error() == "You was logged out from server" {
			// Если удаленная сессия есть, но не активна
			rest.logger.Info("REST: Remote connection timed out", "IP", req.RemoteAddr)
			// Создать новую
			remoteSession = rest.remoteLogin(respwr, req, session)
			if remoteSession == nil {
				return
			}
			// Повторно получить с сайта школы
			lessonDescription, err = remoteSession.GetLessonDescription(date, rReq.ID, cid, tp, strconv.Itoa(studentID))
			if err != nil {
				// Ошибка
				rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadGateway)
				return
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadGateway)
			return
		}
	}
	// Сформировать ответ по протоколу
	// TODO переделать Comments
	s := ""
	for _, v := range lessonDescription.Comments {
		s += v
	}
	resp := getLessonDescriptionResponse{Description: s, Author: "Пока не реализовано", File: "http://Пока_не_реализовано", FileName: "Пока не реализовано"}
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(resp)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", lessonDescription, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", lessonDescription, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", lessonDescription, "IP", req.RemoteAddr)
	}
}

// markAsDoneRequest используется в MarkAsDoneHandler и UnmarkAsDoneHandler
type markAsDoneRequest struct {
	ID int `json:"id"`
}

// MarkAsDoneHandler обрабатывает запрос на отметку задания как сделанного
func (rest *RestAPI) MarkAsDoneHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: MarkAsDoneHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "POST" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	_, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
	}
	// Чтение запроса от клиента
	var rReq markAsDoneRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Info("REST: Malformed request data", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		resp := "malformed data"
		status, err := respwr.Write([]byte(resp))
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
		}
		return
	}
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// Лезть в БД
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	err = rest.Db.TaskMarkDone(userName.(string), schoolID.(int), rReq.ID)
	if err != nil {
		if err.Error() == "record not found" {
			// Такого таска нет в БД
			rest.logger.Info("REST: Task with specified id not found in db", "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadRequest)
			resp := "invalid id"
			status, err := respwr.Write([]byte(resp))
			if err != nil {
				rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
			} else {
				rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when marking task as done in db", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// Отправить ответ клиенту
	rest.logger.Info("REST: Successfully marked task as done", "IP", req.RemoteAddr)
	respwr.WriteHeader(http.StatusOK)
}

// UnmarkAsDoneHandler обрабатывает запрос на отметку задания как просмотренного
func (rest *RestAPI) UnmarkAsDoneHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: UnmarkAsDoneHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "POST" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	_, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
	}
	// Чтение запроса от клиента
	var rReq markAsDoneRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Info("REST: Malformed request data", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		resp := "malformed data"
		status, err := respwr.Write([]byte(resp))
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
		}
		return
	}
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// Лезть в БД
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	err = rest.Db.TaskMarkUndone(userName.(string), schoolID.(int), rReq.ID)
	if err != nil {
		if err.Error() == "record not found" {
			// Такого таска нет в БД
			rest.logger.Info("REST: Task with specified id not found in db", "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadRequest)
			resp := "invalid id"
			status, err := respwr.Write([]byte(resp))
			if err != nil {
				rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
			} else {
				rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when marking task as undone in db", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// Отправить ответ клиенту
	rest.logger.Info("REST: Successfully marked task as undone", "IP", req.RemoteAddr)
	respwr.WriteHeader(http.StatusOK)
}

// scheduleRequest используется в GetScheduleHandler
type scheduleRequest struct {
	Days int `json:"days"`
	ID   int `json:"id"`
}

// GetScheduleHandler возвращает расписание на неделю
func (rest *RestAPI) GetScheduleHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetScheduleHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "POST" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	sessionName, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
	}
	// Чтение запроса от клиента
	var rReq scheduleRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Info("REST: Malformed request data", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		resp := "malformed data"
		status, err := respwr.Write([]byte(resp))
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
		}
		return
	}
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// Получим удаленную сессию
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		// Если нет удаленной сессии
		rest.logger.Info("REST: No remote session", "IP", req.RemoteAddr)
		// Создать новую
		remoteSession = rest.remoteLogin(respwr, req, session)
		if remoteSession == nil {
			return
		}
	}
	today := time.Now().Format("02.01.2006")
	id := strconv.Itoa(rReq.ID)
	timeTable, err := remoteSession.GetTimeTable(today, rReq.Days, id)
	if err != nil {
		if err.Error() == "You was logged out from server" {
			// Если удаленная сессия есть, но не активна
			rest.logger.Info("REST: Remote connection timed out", "IP", req.RemoteAddr)
			// Создать новую
			remoteSession = rest.remoteLogin(respwr, req, session)
			if remoteSession == nil {
				return
			}
			// Повторно получить с сайта школы
			timeTable, err = remoteSession.GetTimeTable(today, rReq.Days, id)
			if err != nil {
				// Ошибка
				rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadGateway)
				return
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadGateway)
			return
		}
	}
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(timeTable)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", timeTable, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", timeTable, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", timeTable, "IP", req.RemoteAddr)
	}
}

// LogOutHandler обрабатывает удаление сессии клиента и отвязку устройства
func (rest *RestAPI) LogOutHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: LogOutHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "GET" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	sessionName, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
	}
	// Удалить куки
	expiration := time.Now().Add(-24 * time.Hour)
	cookie := http.Cookie{Name: "sessionName", Value: sessionName, Expires: expiration}
	http.SetCookie(respwr, &cookie)
	// Отправить ответ клиенту
	rest.logger.Info("REST: Successfully logged out", "IP", req.RemoteAddr)
	respwr.WriteHeader(http.StatusOK)
}

// signInRequest используется в SignInHandler
type signInRequest struct {
	Login   string `json:"login"`
	Passkey string `json:"passkey"`
	ID      int    `json:"id"`
}

// student используется в signInResponse
type student struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// signInResponse используется в SignInHandler
type signInResponse struct {
	Students []student `json:"students"`
}

// SignInHandler обрабатывает вход в учетную запись на сайте школы
func (rest *RestAPI) SignInHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: SignInHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "POST" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Чтение запроса от клиента
	var rReq signInRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Info("REST: Malformed request data", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		resp := "malformed data"
		status, err := respwr.Write([]byte(resp))
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
		}
		return
	}
	// Приведем логин к uppercase
	rReq.Login = strings.ToUpper(rReq.Login)
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// Проверим разрешение у школы
	perm, err := rest.Db.GetSchoolPermission(rReq.ID)
	if err != nil {
		if err.Error() == "record not found" {
			// Школа не найдена
			rest.logger.Info("REST: Invalid school id specified", "Id", rReq.ID, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadRequest)
			resp := "invalid data"
			status, err := respwr.Write([]byte(resp))
			if err != nil {
				rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
			} else {
				rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting school permission from db", "Error", err, "Id", rReq.ID, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if !perm {
		// Если у школы нет разрешения, проверить разрешение пользователя
		userPerm, err := rest.Db.GetUserPermission(rReq.Login, rReq.ID)
		if err != nil {
			if err.Error() == "record not found" {
				// Пользователь новый: вернем true, чтобы он мог залогиниться и
				// купить подписку
				perm = true
			} else {
				// Другая ошибка
				rest.logger.Error("REST: Error occured when getting user permission from db", "Error", err, "Login", rReq.Login, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			perm = userPerm
		}
	}
	if !perm {
		// Если у пользователя тоже нет разрешения
		rest.logger.Info("REST: Access to service denied", "Username", rReq.Login, "SchoolID", rReq.ID, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusPaymentRequired)
		return
	}
	// Взять из конфига структуру школы
	school := rest.config.Schools[rReq.ID-1]
	school.Login = rReq.Login
	school.Password = rReq.Passkey
	// Создать ключ
	uniqueUserKey := strconv.Itoa(rReq.ID) + rReq.Login
	// Проверить существование локальной сессии в редисе
	exists, err := rest.Redis.ExistsCookie(uniqueUserKey)
	if err != nil {
		rest.logger.Error("REST: Error occured when checking key existence in redis", "Error", err, "uniqueUserKey", uniqueUserKey, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	var sessionName string
	var session *sessions.Session
	var remoteSession *ss.Session
	if exists {
		rest.logger.Info("REST: exists in redis", "IP", req.RemoteAddr)
		// Если существует, проверим пароль
		isCorrect, err := rest.Db.CheckPassword(rReq.Login, rReq.ID, rReq.Passkey)
		if err != nil {
			if err.Error() == "record not found" {
				rest.logger.Info("REST: Invalid id or login specified", "Error", err.Error(), "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadRequest)
				resp := "invalid login or id"
				status, err := respwr.Write([]byte(resp))
				if err != nil {
					rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
				} else {
					rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
				}
				return
			}
			rest.logger.Error("REST: Error occured when checking password in db", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		if isCorrect {
			// Если пароль верный, достанем имя сессии
			rest.logger.Info("REST: correct pass", "IP", req.RemoteAddr)
			sessionName, err = rest.Redis.GetCookie(uniqueUserKey)
			if err != nil {
				rest.logger.Error("REST: Error occured when getting existing key from redis", "Error", err, "uniqueUserKey", uniqueUserKey, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			// Получить существующий объект сессии
			session, err = rest.Store.Get(req, sessionName)
			if err != nil {
				rest.logger.Error("REST: Error occured when getting session from cookiestore", "Error", err, "Session name", sessionName, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			// Получим удаленную сессию
			newRemoteSession, ok := rest.sessionsMap[sessionName]
			if !ok {
				// Если нет удаленной сессии
				rest.logger.Info("REST: No remote session", "IP", req.RemoteAddr)
				// Создать удаленную сессию и залогиниться
				newRemoteSession = ss.NewSession(&school)
				err = newRemoteSession.Login()
				if err != nil {
					rest.logger.Error("REST: Error remote signing in", "Error", err, "IP", req.RemoteAddr)
					respwr.WriteHeader(http.StatusBadGateway)
					return
				}
			}
			// Получить мапу
			err = newRemoteSession.GetChildrenMap()
			if err != nil {
				rest.logger.Error("REST: Error occured when getting children map from site", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadGateway)
				return
			}
			rest.sessionsMap[sessionName] = newRemoteSession
			remoteSession = newRemoteSession
		} else {
			// Если неверный, пошлем BadRequest
			rest.logger.Info("REST: incorrect pass", "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadRequest)
			// Отправить ответ клиенту
			resp := "invalid login or passkey"
			status, err := respwr.Write([]byte(resp))
			if err != nil {
				rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
			} else {
				rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
			}
			return
		}
	} else {
		rest.logger.Info("REST: doesn't exist in redis", "IP", req.RemoteAddr)
		// Если не существует, попробуем авторизоваться на удаленном сервере.
		// Создать удаленную сессию и залогиниться
		newRemoteSession := ss.NewSession(&school)
		err = newRemoteSession.Login()
		if err != nil {
			rest.logger.Error("REST: Error remote signing in", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadGateway)
			return
		}
		// Сразу получим мапу имен детей в их ID
		err = newRemoteSession.GetChildrenMap()
		if err != nil {
			rest.logger.Error("REST: Error occured when getting children map from site", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadGateway)
			return
		}
		// Проверить мапу на ошибки
		if newRemoteSession.Base.Children == nil || len(newRemoteSession.Base.Children) == 0 {
			rest.logger.Error("REST: Error occured when getting children map", "Error", "Children nil or empty", "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadGateway)
			return
		}
		// Если логин и пароль правильные, создадим локальную сессию
		// Сгенерировать имя локальной сессии
		timeString := time.Now().String()
		hasher := md5.New()
		if _, err = hasher.Write([]byte(timeString)); err != nil {
			rest.logger.Error("REST: Error occured when hashing for creating session name", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Создать локальную сессию
		newSessionName := hex.EncodeToString(hasher.Sum(nil))
		newLocalSession, err := rest.Store.Get(req, newSessionName)
		if err != nil {
			rest.logger.Error("REST: Error occured when creating local session", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Записать в редис
		err = rest.Redis.AddCookie(uniqueUserKey, newSessionName)
		if err != nil {
			rest.logger.Error("REST: Error occured when adding cookie to redis", "Error", err, "Key", uniqueUserKey, "Value", newSessionName, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Привязать к ней удаленную сессию
		rest.sessionsMap[newSessionName] = newRemoteSession
		remoteSession = newRemoteSession
		// Записать сессию
		sessionName = newSessionName
		session = newLocalSession
		// Обновляем базу данных
		isParent := true
		err = rest.Db.UpdateUser(rReq.Login, rReq.Passkey, isParent, rReq.ID, newRemoteSession.Base.Children)
		if err != nil {
			rest.logger.Error("REST: Error occured when updating user in db", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	// Запихать в сессионные переменные имя пользователя и номер школы
	session.Values["userName"] = rReq.Login
	session.Values["schoolID"] = rReq.ID
	session.Save(req, respwr)
	// Устанавливаем в куки значение sessionName
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: "sessionName", Value: sessionName, Expires: expiration}
	http.SetCookie(respwr, &cookie)
	// Проверить валидность мапы
	if remoteSession == nil || remoteSession.Base.Children == nil || len(remoteSession.Base.Children) == 0 {
		// К этому моменту мапа должна существовать
		rest.logger.Error("REST: Error occured when getting children map from session.Base.Children", "Error", "map is empty or nil", "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Если мапа есть, сформировать ответ
	var stud student
	studs := make([]student, 0)
	for k, v := range remoteSession.Base.Children {
		vs, err := strconv.Atoi(v.SID)
		if err != nil {
			rest.logger.Error("REST: Error occured when converting student id", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		stud = student{Name: k, ID: vs}
		studs = append(studs, stud)
	}
	// Замаршалить
	resp := signInResponse{Students: studs}
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(resp)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", resp, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
	}
}

// Handler временный абстрактный handler для некоторых еще не реализованных
// обработчиков запросов.
func (rest *RestAPI) Handler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: Handler called (not implemented yet)", "Path", req.URL.EscapedPath())
	respwr.WriteHeader(http.StatusNotImplemented)
}
