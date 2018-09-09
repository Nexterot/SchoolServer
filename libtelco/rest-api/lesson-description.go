// lesson-description
package restapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// getLessonDescriptionRequest используется в GetLessonDescriptionHandler
type getLessonDescriptionRequest struct {
	AID       int `json:"AID"`
	CID       int `json:"CID"`
	TP        int `json:"TP"`
	StudentID int `json:"studentID"`
}

// GetLessonDescriptionHandler обрабатывает запрос на получение подробностей дз
func (rest *RestAPI) GetLessonDescriptionHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetLessonDescriptionHandler called", "IP", req.RemoteAddr)
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
	var rReq getLessonDescriptionRequest
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
	// Сходить в бд за class id ученика
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	classID, err := rest.Db.GetStudentClassID(userName.(string), schoolID.(int), rReq.StudentID)
	if err != nil {
		if err.Error() == "record not found" {
			// Такого таска нет
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
	// Получить описание таска
	lessonDescription, err := remoteSession.GetLessonDescription(rReq.AID, rReq.CID, rReq.TP, strconv.Itoa(schoolID.(int)), strconv.Itoa(rReq.StudentID), classID, rest.config.ServerName, rest.Redis)
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
			lessonDescription, err = remoteSession.GetLessonDescription(rReq.AID, rReq.CID, rReq.TP, strconv.Itoa(schoolID.(int)), strconv.Itoa(rReq.StudentID), classID, rest.config.ServerName, rest.Redis)
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
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(lessonDescription)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", lessonDescription, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", lessonDescription, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", lessonDescription, "IP", req.RemoteAddr)
	}
}
