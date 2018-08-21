// children-map
package restapi

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// student используется в GetChildrenMapResponse
type Student struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// GetChildrenMapResponse используется в GetChildrenMapHandler
type GetChildrenMapResponse struct {
	Students []Student `json:"students"`
}

// GetChildrenMapHandler обрабатывает запрос на получение списка детей
func (rest *RestAPI) GetChildrenMapHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetChildrenMapHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "GET" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	sessionName, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
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
	// Заполнить пакет ответа
	var stud Student
	studs := make([]Student, 0)
	// Проверить наличие мапы у сессии парсеров
	if remoteSession.Children == nil || len(remoteSession.Children) == 0 {
		// Если мапа не существует или пустая, полезем в БД
		userName := session.Values["userName"]
		schoolID := session.Values["schoolID"]
		res, err := rest.Db.GetStudents(userName.(string), schoolID.(int))
		if err != nil {
			rest.logger.Error("REST: Error occured when getting children map from db", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
			return
		}
		for k, v := range res {
			stud = Student{Name: k, ID: v}
			studs = append(studs, stud)
		}
	} else {
		// Если мапа есть
		res := remoteSession.Children
		for k, v := range res {
			vs, err := strconv.Atoi(v.SID)
			if err != nil {
				rest.logger.Error("REST: Error occured when converting student id", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
			stud = Student{Name: k, ID: vs}
			studs = append(studs, stud)
		}
	}
	// Замаршалить
	resp := GetChildrenMapResponse{Students: studs}
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
