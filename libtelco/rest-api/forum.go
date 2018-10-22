// forum
package restapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// getForumRequest используется в GetForumHandler
type getForumRequest struct {
	Page int `json:"page"`
}

// GetForumHandler обрабатывает запросы на получение тем форума
func (rest *RestAPI) GetForumHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetForumHandler called", "IP", req.RemoteAddr)
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
	var rReq getForumRequest
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
	// Сходить по удаленной сессии
	forumThemes, err := remoteSession.GetForumThemesList(strconv.Itoa(rReq.Page))
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
			forumThemes, err = remoteSession.GetForumThemesList(strconv.Itoa(rReq.Page))
			if err != nil {
				// Ошибка
				rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadGateway)
				return
			}
		} else if strings.Contains(err.Error(), "current page number is more than max page number") {
			// Запрашиваемая страница не существует
			rest.logger.Info("REST: Invalid page number", "Error", err.Error(), "Page", rReq.Page, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadRequest)
			status, err := respwr.Write(rest.Errors.InvalidPage)
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
	// Сравнить с БД
	userName := session.Values["userName"].(string)
	schoolID := session.Values["schoolID"].(int)
	err = rest.Db.UpdateTopicsStatuses(userName, schoolID, forumThemes)
	if err != nil {
		rest.logger.Error("REST: Error occured when updating statuses for forum themes", "Error", err, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(forumThemes)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", forumThemes, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", forumThemes, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", forumThemes, "IP", req.RemoteAddr)
	}
	// Отправить пуш на удаление пушей с сообщениями форума
	err = rest.pushDelete(userName, schoolID, "forum_new_message")
	if err != nil {
		rest.logger.Error("REST: Error occured when sending deleting push", "Error", err, "Category", "forum_new_message", "IP", req.RemoteAddr)
	}
}
