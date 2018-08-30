// delete-account
package restapi

import (
	"net/http"
	"strconv"
	"time"
)

// DeleteAccountHandler обрабатывает запрос на псевдоудаление аккаунта
func (rest *RestAPI) DeleteAccountHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: DeleteAccountHandler called", "IP", req.RemoteAddr)
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
	// Удалить из редиса
	userName := session.Values["userName"].(string)
	schoolID := session.Values["schoolID"].(int)
	key := strconv.Itoa(schoolID) + userName
	_, err := rest.Redis.DeleteCookie(key)
	if err != nil {
		rest.logger.Error("REST: Error deleting user key from Redis", "Error", err, "IP", req.RemoteAddr)
	}
	// Псевдоудалить из БД
	err = rest.Db.PseudoDeleteUser(userName, schoolID)
	if err != nil {
		rest.logger.Error("REST: Error pseudo-deleting user from DB", "Error", err, "IP", req.RemoteAddr)
	}
	// Удалить куки
	expiration := time.Now().Add(-24 * time.Hour)
	cookie := http.Cookie{Name: "sessionName", Value: sessionName, Expires: expiration}
	http.SetCookie(respwr, &cookie)
	// Отправить ответ клиенту
	rest.logger.Info("REST: Successfully deleted account", "IP", req.RemoteAddr)
	respwr.WriteHeader(http.StatusOK)
}
