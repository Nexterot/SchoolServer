// school-list
package restapi

import (
	"encoding/json"
	"net/http"
)

// school struct используется в GetSchoolListHandler
type school struct {
	Name     string `json:"name"`
	ID       int    `json:"id"`
	Website  string `json:"website"`
	Shortcut string `json:"shortcut"`
}

// SchoolListResponse используется в GetSchoolListHandler
type SchoolListResponse struct {
	Schools []school `json:"schools"`
}

// GetSchoolListHandler обрабатывает запрос на получение списка обслуживаемых школ
func (rest *RestAPI) GetSchoolListHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetSchoolListHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "GET" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Залезть в БД за списком школ
	schools, err := rest.Db.GetSchools()
	if err != nil {
		rest.logger.Error("REST: Error occured when getting school list from db", "Error", err, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Заполняем пакет ответа
	schoolList := make([]school, 0)
	for _, sch := range schools {
		schoolList = append(schoolList, school{sch.Name, int(sch.ID), sch.Address, sch.Initials})
	}
	resp := SchoolListResponse{schoolList}
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
