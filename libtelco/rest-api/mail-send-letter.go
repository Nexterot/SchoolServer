// mail-send-letter
package restapi

import (
	"encoding/json"
	"net/http"
	"strings"
)

// sendLetterRequest используется в SendLetterHandler
type sendLetterRequest struct {
	UserID  string `json:"userID"`
	LBC     string `json:"LBC"`
	LCC     string `json:"LCC"`
	LTO     string `json:"LTO"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

// SendLetterHandler обрабатывает запросы на получение подробностей письма
func (rest *RestAPI) SendLetterHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: SendLetterHandler called", "IP", req.RemoteAddr)
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
	var rReq sendLetterRequest
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
	err = remoteSession.CreateEmail(rReq.UserID, rReq.LBC, rReq.LCC, rReq.LTO, rReq.Name, rReq.Message)
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
			err = remoteSession.CreateEmail(rReq.UserID, rReq.LBC, rReq.LCC, rReq.LTO, rReq.Name, rReq.Message)
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
	// Отправить ответ клиенту
	rest.logger.Info("REST: Successfully sent letter", "IP", req.RemoteAddr)
	respwr.WriteHeader(http.StatusOK)
}
