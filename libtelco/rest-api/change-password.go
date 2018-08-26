// change-password
package restapi

import (
	"encoding/json"
	"net/http"
	"strings"
)

// changePasswordRequest используется в ChangePasswordHandler
type changePasswordRequest struct {
	OldPasskey string `json:"oldPasskey"`
	NewPasskey string `json:"newPasskey"`
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
		remoteSession = rest.remoteRelogin(respwr, req, session)
		if remoteSession == nil {
			return
		}
	}
	// Сходить по удаленной сессии
	err = remoteSession.ChangePassword(rReq.OldPasskey, rReq.NewPasskey)
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
			err = remoteSession.ChangePassword(rReq.OldPasskey, rReq.NewPasskey)
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
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	// Обновить пароль в БД
	err = rest.Db.UpdateUser(userName.(string), rReq.NewPasskey, false, schoolID.(int), nil, nil)
	if err != nil {
		rest.logger.Error("REST: Error occured when saving updated password to DB", "Error", err, "userName", userName, "schoolID", schoolID, "newPasskey", rReq.NewPasskey, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	rest.logger.Info("REST: Successfully updated password", "IP", req.RemoteAddr)
	respwr.WriteHeader(http.StatusOK)
}
