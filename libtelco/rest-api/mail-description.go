// mail-description
package restapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// getMailDescriptionRequest используется в GetMailDescriptionHandler
type getMailDescriptionRequest struct {
	Section int `json:"section"`
	ID      int `json:"id"`
}

// GetMailDescriptionHandler обрабатывает запросы на получение подробностей письма
func (rest *RestAPI) GetMailDescriptionHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetMailDescriptionHandler called", "IP", req.RemoteAddr)
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
	var rReq getMailDescriptionRequest
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
	// Сходить в бд за uid юзера
	schoolID := session.Values["schoolID"].(int)
	userName := session.Values["userName"].(string)
	uid, err := rest.Db.GetUserUID(userName, schoolID)
	if err != nil {
		if err.Error() == "record not found" {
			// Такого пользователя нет
			rest.logger.Info("REST: Invalid student id specified", "Error", err.Error(), "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadRequest)
			status, err := respwr.Write(rest.Errors.InvalidData)
			if err != nil {
				rest.logger.Error("REST: Error occured when sending response", "Error", err, "Status", status, "IP", req.RemoteAddr)
			} else {
				rest.logger.Info("REST: Successfully sent response", "IP", req.RemoteAddr)
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting class id from db", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// Сходить по удаленной сессии
	emailDesc, err := remoteSession.GetEmailDescription(strconv.Itoa(schoolID), uid, strconv.Itoa(rReq.ID), strconv.Itoa(rReq.Section), rest.config.ServerName)
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
			emailDesc, err = remoteSession.GetEmailDescription(strconv.Itoa(schoolID), uid, strconv.Itoa(rReq.ID), strconv.Itoa(rReq.Section), rest.config.ServerName)
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
	// Лезть в БД
	err = rest.Db.MarkMailRead(userName, schoolID, rReq.Section, rReq.ID)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			// Такого письма нет в БД
			rest.logger.Info("REST: Mail with specified id not found in db", "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadRequest)
			status, err := respwr.Write(rest.Errors.InvalidData)
			if err != nil {
				rest.logger.Error("REST: Error occured when sending response", "Error", err, "Status", status, "IP", req.RemoteAddr)
			} else {
				rest.logger.Info("REST: Successfully sent response", "IP", req.RemoteAddr)
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when marking task as done in db", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(emailDesc)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", emailDesc, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", emailDesc, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", emailDesc, "IP", req.RemoteAddr)
	}
}
