// delete-account
package restapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	ss "github.com/masyagin1998/SchoolServer/libtelco/sessions"
)

// deleteAccountRequest используется в DeleteAccountHandler
type deleteAccountRequest struct {
	PassKey string `json:"passKey"`
}

// DeleteAccountHandler обрабатывает запрос на псевдоудаление аккаунта
func (rest *RestAPI) DeleteAccountHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: DeleteAccountHandler called", "IP", req.RemoteAddr)
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
	var rReq deleteAccountRequest
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
	// Проверим длину хеша пароля
	if len(rReq.PassKey) != 32 {
		rest.logger.Info("REST: Invalid passkey len", "Passkey", rReq.PassKey, "Len", len(rReq.PassKey), "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadRequest)
		status, err := respwr.Write(rest.Errors.InvalidLoginData)
		if err != nil {
			rest.logger.Error("REST: Error occured when sending response", "Error", err, "Status", status, "IP", req.RemoteAddr)
		} else {
			rest.logger.Info("REST: Successfully sent response", "IP", req.RemoteAddr)
		}
		return
	}
	userName := session.Values["userName"].(string)
	schoolID := session.Values["schoolID"].(int)
	// Создать новую сессию
	school := rest.config.Schools[schoolID-1]
	school.Login = userName
	school.Password = rReq.PassKey
	newSession := ss.NewSession(&school)
	// Залогиниться
	err = newSession.Login()
	if err != nil {
		if strings.Contains(err.Error(), "invalid login or password") {
			// Пароль неверный
			rest.logger.Info("REST: Error occured when remote signing in", "Error", "invalid login or password", "Config", school, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadRequest)
			status, err := respwr.Write(rest.Errors.InvalidLoginData)
			if err != nil {
				rest.logger.Error("REST: Error occured when sending response", "Error", err, "Status", status, "IP", req.RemoteAddr)
			} else {
				rest.logger.Info("REST: Successfully sent response", "IP", req.RemoteAddr)
			}
			return
		}
		rest.logger.Error("REST: Error remote signing in", "Error", err, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusBadGateway)
		return
	}
	// Вылогиниться
	err = newSession.Logout()
	if err != nil {
		rest.logger.Error("REST: Error logging out from netschool", "Error", err, "IP", req.RemoteAddr)
	}
	// Удалить из редиса
	key := strconv.Itoa(schoolID) + userName
	_, err = rest.Redis.DeleteCookie(key)
	if err != nil {
		rest.logger.Error("REST: Error deleting user key from Redis", "Error", err, "IP", req.RemoteAddr)
	}
	// Псевдоудалить из БД
	err = rest.Db.PseudoDeleteUser(userName, schoolID)
	if err != nil {
		rest.logger.Error("REST: Error pseudo-deleting user from DB", "Error", err, "IP", req.RemoteAddr)
	}
	// Удалить из мапы
	delete(rest.sessionsMap, sessionName)
	// Удалить куки
	expiration := time.Now().Add(-24 * time.Hour)
	cookie := http.Cookie{Name: "sessionName", Value: sessionName, Expires: expiration}
	http.SetCookie(respwr, &cookie)
	// Отправить ответ клиенту
	rest.logger.Info("REST: Successfully deleted account", "IP", req.RemoteAddr)
	respwr.WriteHeader(http.StatusOK)
}
