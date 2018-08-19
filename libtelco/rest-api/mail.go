// mail
package restapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// getMailRequest используется в GetMailHandler
type getMailRequest struct {
	Section    int    `json:"section"`
	PageSize   int    `json:"pageSize"`
	StartIndex int    `json:"startIndex"`
	Order      string `json:"order"`
}

// GetMailHandler обрабатывает запросы на получение списка писем
func (rest *RestAPI) GetMailHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetMailHandler called", "IP", req.RemoteAddr)
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
	var rReq getMailRequest
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
		remoteSession = rest.remoteLogin(respwr, req, session)
		if remoteSession == nil {
			return
		}
	}
	// Если поле page пустое, page = 1
	if rReq.Section == 0 {
		rReq.Section = 1
	}
	// Проверим валидность данных
	if rReq.Section < 1 || rReq.Section > 4 {
		rest.logger.Info("REST: Invalid data", "Error", "must be between 1 and 4", "Section", rReq.Section, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		resp := "invalid page"
		status, err := respwr.Write([]byte(resp))
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
		}
		return
	}
	// Сходить за списком писем по удаленной сессии
	emailsList, err := remoteSession.GetEmailsList(strconv.Itoa(rReq.Section), strconv.Itoa(rReq.StartIndex), strconv.Itoa(rReq.PageSize), rReq.Order)
	if err != nil {
		if strings.Contains(err.Error(), "You was logged out from server") {
			// Если удаленная сессия есть, но не активна
			rest.logger.Info("REST: Remote connection timed out", "IP", req.RemoteAddr)
			// Создать новую
			remoteSession = rest.remoteLogin(respwr, req, session)
			if remoteSession == nil {
				return
			}
			// Повторно получить с сайта школы
			emailsList, err = remoteSession.GetEmailsList(strconv.Itoa(rReq.Section), strconv.Itoa(rReq.StartIndex), strconv.Itoa(rReq.PageSize), rReq.Order)
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
	bytes, err := json.Marshal(emailsList)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", emailsList, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", emailsList, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", emailsList, "IP", req.RemoteAddr)
	}
}