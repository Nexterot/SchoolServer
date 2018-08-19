// file-download
package restapi

import (
	"bufio"
	"net/http"
	"os"
	"strings"
)

// FileHandler обрабатывает запросы на получение файлов
func (rest *RestAPI) FileHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: FileHandler called", "Path", req.URL.Path)
	// Проверка метода запроса
	if req.Method != "GET" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Откроем файл с именем fileName
	fileName := strings.TrimLeft(req.URL.Path, "/doc/")
	file, err := os.Open(fileName)
	if err != nil {
		rest.logger.Info("REST: Couldn't open file", "Error", err.Error(), "Filename", fileName, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusNotFound)
		return
	}
	defer file.Close()
	// Узнать его размер
	stats, statsErr := file.Stat()
	if statsErr != nil {
		rest.logger.Error("REST: Error occured when getting file stats", "Error", statsErr, "Filename", fileName, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	size := stats.Size()
	// Прочитать файл как байты
	bytes := make([]byte, size)
	bufr := bufio.NewReader(file)
	_, err = bufr.Read(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when reading file into memory", "Error", err, "Filename", fileName, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Установить MIME-тип в application/octet-stream
	respwr.Header().Set("Content-Type", "application/octet-stream")
	// Отдать байты клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending file", "Error", err, "Filename", fileName, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Filename", fileName, "IP", req.RemoteAddr)
	}
}
