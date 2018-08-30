// rest-api.go

/*
Package restapi содержит handler'ы для взаимодействия сервера с клиентами.
*/
package restapi

import (
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/context"
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
	Errors      *marshalledErrors
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
		logger.Fatal("REST: Error occured when creating redistore", "Error", err)
	} else {
		logger.Info("REST: Successfully initialized redistore")
	}
	newStore.SetMaxAge(86400 * 365)
	// gorm
	database, err := db.NewDatabase(logger, config)
	if err != nil {
		logger.Fatal("REST: Error occured when initializing database", "Error", err)
	} else {
		logger.Info("REST: Successfully initialized database")
	}
	// redis
	redis, err := red.NewDatabase(config)
	if err != nil {
		logger.Fatal("REST: Error occured when initializing redis", "Error", err)
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
		Errors:      NewMarshalledErrors(logger),
	}
}

// BindHandlers привязывает все handler'ы Rest API
func (rest *RestAPI) BindHandlers() http.Handler {
	mux := http.NewServeMux()

	// Сайт
	mux.Handle("/", http.FileServer(http.Dir("./static"))) // done
	// Общее: Запрос списка школ, запрос доступа, авторизация, выход
	mux.HandleFunc("/get_school_list", rest.GetSchoolListHandler)    // done
	mux.HandleFunc("/check_permission", rest.CheckPermissionHandler) // done
	mux.HandleFunc("/sign_in", rest.SignInHandler)                   // done
	mux.HandleFunc("/log_out", rest.LogOutHandler)                   // done
	mux.HandleFunc("/delete_account", rest.DeleteAccountHandler)     // done
	// Дневник: задания и оценки на неделю, отметить задание
	// как выполненное/невыполненное
	mux.HandleFunc("/get_tasks_and_marks", rest.GetTasksAndMarksHandler)        // done
	mux.HandleFunc("/get_lesson_description", rest.GetLessonDescriptionHandler) // done
	mux.HandleFunc("/mark_as_done", rest.MarkAsDoneHandler)                     // done
	mux.HandleFunc("/unmark_as_done", rest.UnmarkAsDoneHandler)                 // done
	// Объявления: получение списка объявлений
	mux.HandleFunc("/get_posts", rest.Handler)
	// Расписание: получение расписания на N дней
	mux.HandleFunc("/get_schedule", rest.GetScheduleHandler) // done
	// Отчеты
	mux.HandleFunc("/get_report_student_total_marks", rest.GetReportStudentTotalMarksHandler)              // done
	mux.HandleFunc("/get_report_student_average_mark", rest.GetReportStudentAverageMarkHandler)            // done
	mux.HandleFunc("/get_report_student_average_mark_dyn", rest.GetReportStudentAverageMarkDynHandler)     // done
	mux.HandleFunc("/get_report_student_grades_lesson_list", rest.GetReportStudentGradesLessonListHandler) // done
	mux.HandleFunc("/get_report_student_grades", rest.GetReportStudentGradesHandler)                       // done
	mux.HandleFunc("/get_report_student_total", rest.GetReportStudentTotalHandler)                         // done
	mux.HandleFunc("/get_report_journal_access", rest.GetReportJournalAccessHandler)                       // done
	mux.HandleFunc("/get_report_parent_info_letter_data", rest.GetReportParentInfoLetterDataHandler)       // done
	mux.HandleFunc("/get_report_parent_info_letter", rest.GetReportParentInfoLetterHandler)                // done
	// Школьные ресурсы
	mux.HandleFunc("/get_resources", rest.GetResourcesHandler) // done
	// Почта
	mux.HandleFunc("/get_mail", rest.GetMailHandler)                        // done
	mux.HandleFunc("/get_mail_description", rest.GetMailDescriptionHandler) // done
	mux.HandleFunc("/delete_mail", rest.DeleteMailHandler)                  // done
	mux.HandleFunc("/send_letter", rest.SendLetterHandler)                  // done
	mux.HandleFunc("/get_address_book", rest.GetAddressBookHandler)         // done
	// Форум
	mux.HandleFunc("/get_forum", rest.GetForumHandler)                         // done
	mux.HandleFunc("/get_forum_messages", rest.GetForumMessagesHandler)        // done
	mux.HandleFunc("/create_topic", rest.CreateTopicHandler)                   // done
	mux.HandleFunc("/create_message_in_topic", rest.CreateTopicMessageHandler) // done
	// Настройки
	mux.HandleFunc("/change_password", rest.ChangePasswordHandler) // done
	// Файлы
	mux.Handle("/doc/", http.StripPrefix("/doc/", http.FileServer(http.Dir(".")))) // done

	rest.logger.Info("REST: Successfully bound handlers")
	return context.ClearHandler(mux)
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
	// Создать удаленную сессию и залогиниться
	remoteSession := ss.NewSession(school)
	err = remoteSession.Login()
	if err != nil {
		if strings.Contains(err.Error(), "invalid login or passkey") {
			// Пароль неверный
			rest.logger.Info("REST: Error occured when remote signing in", "Error", "invalid login or password", "Config", school, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadRequest)
			status, err := respwr.Write(rest.Errors.InvalidLoginData)
			if err != nil {
				rest.logger.Error("REST: Error occured when sending response", "Error", err, "Status", status, "IP", req.RemoteAddr)
			} else {
				rest.logger.Info("REST: Successfully sent response", "IP", req.RemoteAddr)
			}
			return nil
		}
		rest.logger.Error("REST: Error occured when remote signing in", "Error", err, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadGateway)
		return nil
	}
	// Получить childrenMap
	err = remoteSession.GetChildrenMap()
	if err != nil {
		// Ошибка
		rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadGateway)
		return nil
	}
	rest.sessionsMap[session.Name()] = remoteSession
	rest.logger.Info("REST: Successfully created new remote session", "IP", req.RemoteAddr)
	return remoteSession
}
