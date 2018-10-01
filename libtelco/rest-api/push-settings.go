// push-settings
package restapi

import (
	"encoding/json"
	"net/http"
	"strings"
)

// pushSettingsRequest используется в PushSettingsHandler
type pushSettingsRequest struct {
	Token      string `json:"token"`
	SystemType int    `json:"systemType"`
	Marks      int    `json:"marks"`
	Tasks      int    `json:"tasks"`
	Reports    bool   `json:"reports"`
	Schedule   bool   `json:"schedule"`
	Mail       bool   `json:"mail"`
	Forum      bool   `json:"forum"`
	Resources  bool   `json:"resources"`
}

// PushSettingsHandler обрабатывает запрос на отметку задания как сделанного
func (rest *RestAPI) PushSettingsHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: PushSettingsHandler called", "IP", req.RemoteAddr)
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
	var rReq pushSettingsRequest
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
	// Проверим systemType
	if rReq.SystemType != 1 && rReq.SystemType != 2 {
		rest.logger.Info("REST: Invalid system type", "systemType", rReq.SystemType, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		status, err := respwr.Write(rest.Errors.InvalidSystemType)
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "IP", req.RemoteAddr)
		}
		return
	}
	// Проверим token
	if rReq.Token == "" {
		rest.logger.Info("REST: Empty token", "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		status, err := respwr.Write(rest.Errors.InvalidToken)
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "IP", req.RemoteAddr)
		}
		return
	}
	// Проверим настройки
	if rReq.Tasks < 1 || rReq.Tasks > 3 {
		rest.logger.Info("REST: Invalid tasks setting specified", "IP", req.RemoteAddr, "tasks", rReq.Tasks)
		respwr.WriteHeader(http.StatusBadRequest)
		status, err := respwr.Write(rest.Errors.InvalidData)
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "IP", req.RemoteAddr)
		}
		return
	}
	if rReq.Marks < 1 || rReq.Marks > 3 {
		rest.logger.Info("REST: Invalid marks setting specified", "IP", req.RemoteAddr, "marks", rReq.Marks)
		respwr.WriteHeader(http.StatusBadRequest)
		status, err := respwr.Write(rest.Errors.InvalidData)
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "IP", req.RemoteAddr)
		}
		return
	}
	// Распечатаем запрос от клиента
	rest.logger.Info("REST: Request data", "Data", rReq, "IP", req.RemoteAddr)
	// Обновить значения в БД
	userName := session.Values["userName"].(string)
	schoolID := session.Values["schoolID"].(int)
	err = rest.Db.UpdatePushSettings(userName, schoolID, rReq.SystemType, rReq.Token, rReq.Marks, rReq.Tasks, rReq.Reports, rReq.Schedule, rReq.Mail, rReq.Forum, rReq.Resources)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			// Такого девайса нет в БД
			rest.logger.Info("REST: Device with specified id not found in db", "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadRequest)
			status, err := respwr.Write(rest.Errors.InvalidData)
			if err != nil {
				rest.logger.Error("REST: Error occured when sending response", "Error", err, "Status", status, "IP", req.RemoteAddr)
			} else {
				rest.logger.Info("REST: Successfully sent response", "IP", req.RemoteAddr)
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when updating push settings", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// Отправить ответ клиенту
	rest.logger.Info("REST: Successfully updated push settings for device", "IP", req.RemoteAddr, "SystemType", rReq.SystemType, "Token", rReq.Token)
	respwr.WriteHeader(http.StatusOK)
}
