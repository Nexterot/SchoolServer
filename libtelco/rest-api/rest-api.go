// rest-api.go

/*
Package restapi содержит handler'ы для взаимодействия сервера с клиентами.
*/
package restapi

import (
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"
	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	red "github.com/masyagin1998/SchoolServer/libtelco/in-memory-db"
	"github.com/masyagin1998/SchoolServer/libtelco/log"
	ss "github.com/masyagin1998/SchoolServer/libtelco/sessions"
	db "github.com/masyagin1998/SchoolServer/libtelco/sql-db"
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
		strconv.Itoa(config.CookieStore.DBname),
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
	http.HandleFunc("/get_mail", rest.GetMailHandler)                        // in dev
	http.HandleFunc("/get_mail_description", rest.GetMailDescriptionHandler) // in dev
	http.HandleFunc("/delete_mail", rest.Handler)
	http.HandleFunc("/send_letter", rest.Handler)
	http.HandleFunc("/get_address_book", rest.Handler)
	// Форум
	http.HandleFunc("/get_forum", rest.GetForumHandler) // in dev
	http.HandleFunc("/get_forum_messages", rest.Handler)
	http.HandleFunc("/create_topic", rest.Handler)
	http.HandleFunc("/create_message_in_topic", rest.Handler)
	// Настройки
	http.HandleFunc("/change_password", rest.Handler)
	// Файлы
	http.HandleFunc("/doc/", rest.FileHandler) // done

	rest.logger.Info("REST: Successfully bound handlers")
}

// Handler временный абстрактный handler для некоторых еще не реализованных
// обработчиков запросов.
func (rest *RestAPI) Handler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: Handler called (not implemented yet)", "Path", req.URL.Path)
	respwr.WriteHeader(http.StatusNotImplemented)
}

// ErrorHandler обрабатывает некорректные запросы.
func (rest *RestAPI) ErrorHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("Wrong request", "Path", req.URL.Path, "IP", req.RemoteAddr)
	respwr.WriteHeader(http.StatusNotFound)
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

// remoteRelogin повторно авторизуется на сайте школы
func (rest *RestAPI) remoteRelogin(respwr http.ResponseWriter, req *http.Request, session *sessions.Session) *ss.Session {
	rest.logger.Info("REST: Remote signing in", "IP", req.RemoteAddr)
	// Полезть в базу данных за данными для авторизации
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	rest.logger.Info("go to database")
	school, err := rest.Db.GetUserAuthData(userName.(string), schoolID.(int))
	if err != nil {
		// Ошибок тут быть не должно
		rest.logger.Error("REST: Error occured when getting user auth data from db", "Username", userName, "SchoolID", schoolID, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	rest.logger.Info("successfully gone to database")
	// Создать удаленную сессию и залогиниться
	remoteSession := ss.NewSession(school)
	rest.logger.Info("create session")
	err = remoteSession.Login()
	if err != nil {
		rest.logger.Error("REST: Error occured when remote signing in", "Error", err, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadGateway)
		return nil
	}
	rest.logger.Info("successfully created session")
	// Получить childrenMap
	rest.logger.Info("getting childrenMap")
	err = remoteSession.GetChildrenMap()
	if err != nil {
		// Ошибка
		rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadGateway)
		return nil
	}
	rest.logger.Info("successfully got childrenMap")
	rest.sessionsMap[session.Name()] = remoteSession
	rest.logger.Info("REST: Successfully created new remote session", "IP", req.RemoteAddr)
	return remoteSession
}
