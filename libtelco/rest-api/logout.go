// logout
package restapi

import (
	"net/http"
	"time"
)

// LogOutHandler обрабатывает удаление сессии клиента и отвязку устройства
func (rest *RestAPI) LogOutHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: LogOutHandler called", "IP", req.RemoteAddr)
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
	// Удалить куки
	expiration := time.Now().Add(-24 * time.Hour)
	cookie := http.Cookie{Name: "sessionName", Value: sessionName, Expires: expiration}
	http.SetCookie(respwr, &cookie)
	// Отправить ответ клиенту
	rest.logger.Info("REST: Successfully logged out", "IP", req.RemoteAddr)
	respwr.WriteHeader(http.StatusOK)
}
