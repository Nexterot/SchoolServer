// push-do-not-disturb
package restapi

import (
	"encoding/json"
	"net/http"
	"strings"
)

// pushDontDisturbRequest используется в PushDontDisturbHandler
type pushDontDisturbRequest struct {
	Minutes    int    `json:"minutes"`
	Token      string `json:"token"`
	SystemType int    `json:"systemType"`
}

// PushDontDisturbHandler обрабатывает запросы на удаление письма
func (rest *RestAPI) PushDontDisturbHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: PushDontDisturbHandler called", "IP", req.RemoteAddr)
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
	var rReq pushDontDisturbRequest
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
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	// Обновить время в БД
	err = rest.Db.UpdatePushTime(userName.(string), schoolID.(int), rReq.Token, rReq.SystemType, rReq.Minutes)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			rest.logger.Info("REST: Invalid device info specified", "Error", err.Error(), "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadRequest)
			status, err := respwr.Write(rest.Errors.InvalidDeviceInfo)
			if err != nil {
				rest.logger.Error("REST: Error occured when sending response", "Error", err, "Status", status, "IP", req.RemoteAddr)
			} else {
				rest.logger.Info("REST: Successfully sent response", "IP", req.RemoteAddr)
			}
			return
		}
		rest.logger.Error("REST: Error occured when saving do not disturb time to DB", "Error", err, "userName", userName, "schoolID", schoolID, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	rest.logger.Info("REST: Successfully updated do not disturb time", "IP", req.RemoteAddr)
	respwr.WriteHeader(http.StatusOK)
}
