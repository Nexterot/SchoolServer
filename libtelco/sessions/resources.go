// Copyright (C) 2018 Mikhail Masyagin

/*
Package sessions - данный файл содержит в себе получение ресурсов.
*/
package sessions

import (
	cp "SchoolServer/libtelco/config-parser"
	dt "SchoolServer/libtelco/sessions/data-types"
	t01 "SchoolServer/libtelco/sessions/type-01"
	"fmt"
)

// GetResourcesList возвращает список всех ресурсов.
func (s *Session) GetResourcesList() (*dt.Resources, error) {
	s.Base.MU.Lock()
	defer s.Base.MU.Unlock()
	var err error
	var resources *dt.Resources
	switch s.Base.Serv.Type {
	case cp.FirstType:
		resources, err = t01.GetResourcesList(s.Base)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Base.Serv.Type)
	}
	return resources, err
}

// GetFile выкачивает файл по заданной ссылке в заданную директорию (если его там ещё нет) и возвращает
// - true, если файл был скачан;
// - false, если файл уже был в директории.
func (s *Session) GetFile(link, path string) (bool, error) {

	return false, nil
}
