// tasks-and-marks
package restapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// tasksMarksRequest используется в GetTasksAndMarksHandler
type tasksMarksRequest struct {
	Week string `json:"week"`
	ID   int    `json:"id"`
}

// lesson используется в day
type lesson struct {
	AID    int    `json:"AID"`
	CID    int    `json:"CID"`
	TP     int    `json:"TP"`
	Status int    `json:"status"`
	InTime bool   `json:"inTime"`
	Name   string `json:"name"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	Mark   string `json:"mark"`
	Weight string `json:"weight"`
}

// day используется в tasksMarksResponse
type day struct {
	Date    string   `json:"date"`
	Lessons []lesson `json:"lessons"`
}

// tasksMarksResponse используется в GetTasksAndMarksHandler
type tasksMarksResponse struct {
	Days []day `json:"days"`
}

// GetTasksAndMarksHandler возвращает задания и оценки на неделю
func (rest *RestAPI) GetTasksAndMarksHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetTasksAndMarksHandler called", "IP", req.RemoteAddr)
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
	var rReq tasksMarksRequest
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
	// По протоколу пустое поле заменить на дату первого дня текущей недели
	// (пока просто текущий день)
	if rReq.Week == "" {
		rReq.Week = time.Now().Format("02.01.2006")
	}
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
	// Получить с сайта школы
	weekMarks, err := remoteSession.GetWeekSchoolMarks(rReq.Week, strconv.Itoa(rReq.ID))
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
			weekMarks, err = remoteSession.GetWeekSchoolMarks(rReq.Week, strconv.Itoa(rReq.ID))
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
	// Обновить статусы заданий
	userName := session.Values["userName"]
	schoolID := session.Values["schoolID"]
	// Сходить в бд
	err = rest.Db.UpdateTasksStatuses(userName.(string), schoolID.(int), rReq.ID, weekMarks)
	if err != nil {
		rest.logger.Error("REST: Error occured when updating statuses for tasks and marks", "Error", err, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Сформировать ответ по протоколу
	var resp tasksMarksResponse
	days := make([]day, 0)
	for _, d := range weekMarks.Data {
		lessons := make([]lesson, 0)
		for _, l := range d.Lessons {
			newLesson := lesson{AID: l.AID, CID: l.CID, TP: l.TP, Status: l.Status, InTime: l.InTime, Name: l.Name, Title: l.Title, Type: l.Type, Mark: l.Mark, Weight: l.Weight}
			lessons = append(lessons, newLesson)
		}
		newDay := day{Date: d.Date, Lessons: lessons}
		days = append(days, newDay)
	}
	resp.Days = days
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(resp)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", resp, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
	}
}
