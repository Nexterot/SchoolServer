// mark-unmark-done
package restapi

import (
	"encoding/json"
	"net/http"
)

// markAsDoneRequest используется в MarkAsDoneHandler и UnmarkAsDoneHandler
type markAsDoneRequest struct {
	AID int `json:"AID"`
	CID int `json:"CID"`
	TP  int `json:"TP"`
}

// MarkAsDoneHandler обрабатывает запрос на отметку задания как сделанного
func (rest *RestAPI) MarkAsDoneHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: MarkAsDoneHandler called", "IP", req.RemoteAddr)
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
	var rReq markAsDoneRequest
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
	// Лезть в БД
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	err = rest.Db.TaskMarkDone(userName.(string), schoolID.(int), rReq.AID, rReq.CID, rReq.TP)
	if err != nil {
		if err.Error() == "record not found" {
			// Такого таска нет в БД
			rest.logger.Info("REST: Task with specified id not found in db", "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadRequest)
			resp := "invalid id"
			status, err := respwr.Write([]byte(resp))
			if err != nil {
				rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
			} else {
				rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when marking task as done in db", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// Отправить ответ клиенту
	rest.logger.Info("REST: Successfully marked task as done", "IP", req.RemoteAddr)
	respwr.WriteHeader(http.StatusOK)
}

// UnmarkAsDoneHandler обрабатывает запрос на отметку задания как просмотренного
func (rest *RestAPI) UnmarkAsDoneHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: UnmarkAsDoneHandler called", "IP", req.RemoteAddr)
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
	var rReq markAsDoneRequest
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
	// Лезть в БД
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	err = rest.Db.TaskMarkUndone(userName.(string), schoolID.(int), rReq.AID, rReq.CID, rReq.TP)
	if err != nil {
		if err.Error() == "record not found" {
			// Такого таска нет в БД
			rest.logger.Info("REST: Task with specified id not found in db", "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadRequest)
			resp := "invalid id"
			status, err := respwr.Write([]byte(resp))
			if err != nil {
				rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
			} else {
				rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when marking task as undone in db", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// Отправить ответ клиенту
	rest.logger.Info("REST: Successfully marked task as undone", "IP", req.RemoteAddr)
	respwr.WriteHeader(http.StatusOK)
}
