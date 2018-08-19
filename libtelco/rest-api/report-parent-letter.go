// report-parent-letter
package restapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// getReportParentInfoLetterRequest используется в GetReportParentInfoLetterHandler
type getReportParentInfoLetterRequest struct {
	StudentID    int `json:"student_id"`
	ReportTypeID int `json:"report_type_id"`
	PeriodID     int `json:"period_id"`
}

// GetReportParentInfoLetterHandler обрабатывает запрос на получение шаблона
// письма родителям
func (rest *RestAPI) GetReportParentInfoLetterHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetReportParentInfoLetterHandler called", "IP", req.RemoteAddr)
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
	var rReq getReportParentInfoLetterRequest
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
	// Получить отчет с сайта школы
	studentID := strconv.Itoa(rReq.StudentID)
	reportID := strconv.Itoa(rReq.ReportTypeID)
	periodID := strconv.Itoa(rReq.PeriodID)
	parentLetter, err := remoteSession.GetParentInfoLetterReport(reportID, periodID, studentID)
	if err != nil {
		if strings.Contains(err.Error(), "You was logged out from server") {
			// Если удаленная сессия есть, но не активна
			rest.logger.Info("REST: Remote connection timed out", "IP", req.RemoteAddr)
			// Создать новую
			remoteSession = rest.remoteLogin(respwr, req, session)
			if remoteSession == nil {
				return
			}
			// Повторно получить отчет с сайта школы
			parentLetter, err = remoteSession.GetParentInfoLetterReport(reportID, periodID, studentID)
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
	bytes, err := json.Marshal(parentLetter)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", parentLetter, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", parentLetter, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", parentLetter, "IP", req.RemoteAddr)
	}
}
