// change-password
package restapi

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
)

// changePasswordRequest используется в ChangePasswordHandler
type changePasswordRequest struct {
	Old string `json:"old"`
	New string `json:"new"`
}

// getMD5Hash получает md5 сумму.
func getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// ChangePasswordHandler обрабатывает запросы на удаление письма
func (rest *RestAPI) ChangePasswordHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: ChangePasswordHandler called", "IP", req.RemoteAddr)
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
	var rReq changePasswordRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rReq)
	if err != nil {
		rest.logger.Info("REST: Malformed request data", "Error", err.Error(), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		status, err := respwr.Write(rest.Errors.MalformedData)
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "IP", req.RemoteAddr)
		}
		return
	}
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// Проверим валидность данных
	if rReq.New == "" || rReq.Old == "" {
		rest.logger.Info("REST: Invalid data: empty passwords", "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		status, err := respwr.Write(rest.Errors.InvalidData)
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "IP", req.RemoteAddr)
		}
		return
	}
	// Получим удаленную сессию
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		// Если нет удаленной сессии
		rest.logger.Info("REST: No remote session", "IP", req.RemoteAddr)
		// Создать новую
		remoteSession = rest.remoteRelogin(respwr, req, session)
		if remoteSession == nil {
			return
		}
	}
	// Применить md5 к паролям
	oldPasskey := getMD5Hash(rReq.Old)
	newPasskey := getMD5Hash(rReq.New)
	// Сходить по удаленной сессии
	err = remoteSession.ChangePassword(oldPasskey, newPasskey)
	if err != nil {
		if strings.Contains(err.Error(), "You was logged out from server") {
			// Если удаленная сессия есть, но не активна
			rest.logger.Info("REST: Remote connection timed out", "IP", req.RemoteAddr)
			// Создать новую
			remoteSession = rest.remoteRelogin(respwr, req, session)
			if remoteSession == nil {
				return
			}
			// Повторно получить с сайта школы
			err = remoteSession.ChangePassword(oldPasskey, newPasskey)
			if err != nil {
				// Ошибка
				rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadGateway)
				return
			}
		} else if strings.Contains(err.Error(), "Invalid old password") {
			// Если старый пароль введен неверно
			rest.logger.Info("REST: Invalid old password")
			respwr.WriteHeader(http.StatusBadRequest)
			status, err := respwr.Write(rest.Errors.WrongOldPassword)
			if err != nil {
				rest.logger.Error("REST: Error occured when sending response", "Error", err, "Status", status, "IP", req.RemoteAddr)
			} else {
				rest.logger.Info("REST: Successfully sent response", "IP", req.RemoteAddr)
			}
			return
		} else if strings.Contains(err.Error(), "Equal new and old passwords") {
			// Если пароли одинаковы
			rest.logger.Info("REST: Equal new and old passwords")
			respwr.WriteHeader(http.StatusBadRequest)
			status, err := respwr.Write(rest.Errors.SamePassword)
			if err != nil {
				rest.logger.Error("REST: Error occured when sending response", "Error", err, "Status", status, "IP", req.RemoteAddr)
			} else {
				rest.logger.Info("REST: Successfully sent response", "IP", req.RemoteAddr)
			}
			return
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadGateway)
			return
		}
	}
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	// Обновить пароль в БД
	err = rest.Db.UpdateUser(userName.(string), newPasskey, schoolID.(int), "", 0, nil, nil)
	if err != nil {
		rest.logger.Error("REST: Error occured when saving updated password to DB", "Error", err, "userName", userName, "schoolID", schoolID, "newPasskey", newPasskey, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	rest.logger.Info("REST: Successfully updated password", "IP", req.RemoteAddr)
	respwr.WriteHeader(http.StatusOK)
}
