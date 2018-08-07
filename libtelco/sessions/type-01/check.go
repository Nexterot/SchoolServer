// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

/*
Package type01 - данный файл содержит в себе проверку ответа на ошибки.
*/
package type01

import (
	ss "SchoolServer/libtelco/sessions/session"
	"fmt"
	"strings"

	gr "github.com/levigross/grequests"
)

// checkResponse проверяет, не выкинуло ли нас с сервера и не пришло ли предупреждение.
func checkResponse(s *ss.Session, response *gr.Response) error {
	str := string(response.Bytes())
	fmt.Println(str)
	// Проверка на "выброс с сервера".
	if (response.StatusCode != 200) || strings.Contains(strings.ToLower(str), "<title></title>") {
		return fmt.Errorf("You was logged out from server")
	}
	return nil
}
