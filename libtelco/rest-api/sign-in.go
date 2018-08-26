// sign-in
package restapi

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	ss "github.com/masyagin1998/SchoolServer/libtelco/sessions"
)

// signInRequest используется в SignInHandler
type signInRequest struct {
	Login   string `json:"login"`
	Passkey string `json:"passkey"`
	ID      int    `json:"id"`
}

// student используется в signInResponse
type student struct {
	Name    string `json:"name"`
	ID      int    `json:"id"`
	ClassID string `json:"classID"`
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
		// Подразумевается, что уже был успешный вход
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
				// Получить childrenMap
				err = newRemoteSession.GetChildrenMap()
				if err != nil {
					// Ошибка
					rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
					respwr.WriteHeader(http.StatusBadGateway)
					return
				}
				rest.sessionsMap[sessionName] = newRemoteSession
			}
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
			if strings.Contains(err.Error(), "invalid login or password") {
				// Пароль неверный
				rest.logger.Info("REST: Error occured when remote signing in", "Error", "invalid login or password", "Config", school, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadRequest)
				resp := "invalid login or password"
				status, err := respwr.Write([]byte(resp))
				if err != nil {
					rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
				} else {
					rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
				}
				return
			}
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
		if newRemoteSession.Children == nil || len(newRemoteSession.Children) == 0 {
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
		err = rest.Db.UpdateUser(rReq.Login, rReq.Passkey, isParent, rReq.ID, newRemoteSession.Children)
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
	if remoteSession == nil || remoteSession.Children == nil || len(remoteSession.Children) == 0 {
		// К этому моменту мапа должна существовать
		rest.logger.Error("REST: Error occured when getting children map from session.Children", "Error", "map is empty or nil", "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Если мапа есть, сформировать ответ
	var stud student
	studs := make([]student, 0)
	for k, v := range remoteSession.Children {
		vs, err := strconv.Atoi(v.SID)
		if err != nil {
			rest.logger.Error("REST: Error occured when converting student id", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		stud = student{Name: k, ID: vs, ClassID: v.CLID}
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
