// logout
package restapi

import (
	"encoding/json"
	"net/http"
	"time"
)

// logOutRequest используется в LogOutHandler
type logOutRequest struct {
	Token      string `json:"token"`
	SystemType int    `json:"systemType"`
}

// LogOutHandler обрабатывает удаление сессии клиента и отвязку устройства
func (rest *RestAPI) LogOutHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: LogOutHandler called", "IP", req.RemoteAddr)
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
	var rReq logOutRequest
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
	// Удалить устройство из БД
	userName := session.Values["userName"].(string)
	schoolID := session.Values["schoolID"].(int)
	err = rest.Db.PseudoDeleteDevice(userName, schoolID, rReq.Token, rReq.SystemType)
	if err != nil {
		rest.logger.Error("REST: Error when deleting device from DB", "Error", err, "userName", userName, "schoolID", schoolID, "token", rReq.Token, "IP", req.RemoteAddr)
	}
	// Удалить куки
	expiration := time.Now().Add(-24 * time.Hour)
	cookie := http.Cookie{Name: "sessionName", Value: sessionName, Expires: expiration}
	http.SetCookie(respwr, &cookie)
	// Отправить ответ клиенту
	rest.logger.Info("REST: Successfully logged out", "IP", req.RemoteAddr)
	respwr.WriteHeader(http.StatusOK)
}
