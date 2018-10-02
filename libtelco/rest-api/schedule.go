// schedule
package restapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// scheduleRequest используется в GetScheduleHandler
type scheduleRequest struct {
	Days int    `json:"days"`
	Date string `json:"date"`
	ID   int    `json:"id"`
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
	// Получить расписание
	timeTable, err := remoteSession.GetTimeTable(rReq.Date, rReq.Days, strconv.Itoa(rReq.ID))
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
			timeTable, err = remoteSession.GetTimeTable(rReq.Date, rReq.Days, strconv.Itoa(rReq.ID))
			if err != nil {
				// Ошибка
				rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadGateway)
				return
			}
		} else if strings.Contains(err.Error(), "Invalid days number") {
			respwr.WriteHeader(http.StatusBadRequest)
			status, err := respwr.Write(rest.Errors.InvalidData)
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
	// Записать в БД
	userName := session.Values["userName"].(string)
	schoolID := session.Values["schoolID"].(int)
	err = rest.Db.UpdateLessons(userName, schoolID, rReq.ID, timeTable)
	if err != nil {
		rest.logger.Error("REST: Error occured when updating schedule in DB", "Error", err, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
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
