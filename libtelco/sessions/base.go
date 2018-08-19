// Copyright (C) 2018 Mikhail Masyagin

/*
Package sessions - данный файл содержит в себе сессии на серверах школ.
*/
package sessions

import (
	"fmt"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	dt "github.com/masyagin1998/SchoolServer/libtelco/sessions/datatypes"
	t01 "github.com/masyagin1998/SchoolServer/libtelco/sessions/servers/01/base"

	"github.com/pkg/errors"
)

/*
Вход в систему.
*/

// Login логинится к серверу и создает очередную сессию.
func (s *Session) Login() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	switch s.Serv.Type {
	case cp.FirstType:
		err = t01.Login(&s.Session)
		err = errors.Wrap(err, "type-01")
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return errors.Wrap(err, "from Login")
}

/*
Выход из системы.
*/

// Logout выходит с сервера.
func (s *Session) Logout() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	switch s.Serv.Type {
	case cp.FirstType:
		err = t01.Logout(&s.Session)
		err = errors.Wrap(err, "type-01")
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return errors.Wrap(err, "from Logout")
}

/*
Получение списка детей.
*/

// GetChildrenMap получает мапу детей в их ID.
func (s *Session) GetChildrenMap() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	switch s.Serv.Type {
	case cp.FirstType:
		err = t01.GetChildrenMap(&s.Session)
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return errors.Wrap(err, "from GetChildrenMap")
}

/*
Получение списка предметов.
*/

// GetLessonsMap возвращает список пар мапу предметов в их ID.
func (s *Session) GetLessonsMap(studentID string) (*dt.LessonsMap, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if studentID == "" {
		studentID = s.Child.SID
	}
	var err error
	var lessonsMap *dt.LessonsMap
	switch s.Serv.Type {
	case cp.FirstType:
		lessonsMap, err = t01.GetLessonsMap(&s.Session, studentID)
		err = errors.Wrap(err, "type-01")
	default:
		err = fmt.Errorf("Unknown SchoolServer Type: %d", s.Serv.Type)
	}
	return lessonsMap, errors.Wrap(err, "from GetLessonsMap")
}
