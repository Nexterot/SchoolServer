// check-permission
package restapi

import (
	"encoding/json"
	"net/http"
)

// permissionCheckRequest используется в CheckPermissionHandler
type permissionCheckRequest struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
}

// permissionCheckResponse используется в CheckPermissionHandler
type permissionCheckResponse struct {
	Permission bool `json:"permission"`
}

// CheckPermissionHandler проверяет, есть ли разрешение на работу с школой
func (rest *RestAPI) CheckPermissionHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: CheckPermissionHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "POST" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Чтение json'a
	var rReq permissionCheckRequest
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
	// Проверим разрешение у школы
	perm, err := rest.Db.GetSchoolPermission(rReq.ID)
	if err != nil {
		if err.Error() == "record not found" {
			// Школа не найдена
			rest.logger.Info("REST: Invalid school id specified", "Id", rReq.ID, "IP", req.RemoteAddr)
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
			rest.logger.Error("REST: Error occured when getting school permission from db", "Error", err, "Id", rReq.ID, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if !perm {
		// Если у школы нет разрешения, проверить разрешение пользователя
		userPerm, err := rest.Db.GetUserPermission(rReq.Login, rReq.ID)
		if err != nil {
			if err.Error() == "record not found" {
				// Пользователь новый: вернем true, чтобы он мог залогиниться и
				// купить подписку
				perm = true
			} else {
				// Другая ошибка
				rest.logger.Error("REST: Error occured when getting user permission from db", "Error", err, "Login", rReq.Login, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			perm = userPerm
		}
	}
	// Закодировать ответ в JSON
	resp := permissionCheckResponse{perm}
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
